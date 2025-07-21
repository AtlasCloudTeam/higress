package main

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/resp"

	"github.com/alibaba/higress/plugins/wasm-go/pkg/wrapper"
	"github.com/higress-group/proxy-wasm-go-sdk/proxywasm"
	"github.com/higress-group/proxy-wasm-go-sdk/proxywasm/types"
)

const (
	DefaultRejectedCode uint32 = 429
	DefaultRejectedMsg  string = "Too many requests"

	XAiModel    = "x-ai-model"
	XAiRejected = "x-ai-rejected"
)

const (
	// ClusterKeyPrefix 集群限流插件在 Redis 中 key 的统一前缀
	ClusterKeyPrefix = "model-rate-limit"
	// ClusterRateLimitFormat 规则限流模式 redis key 为 ClusterKeyPrefix:限流规则名称:限流key对应的实际值
	ClusterRateLimitFormat = ClusterKeyPrefix + ":%s:%s"
	FixedWindowScript      = `
local ttl = redis.call('ttl', KEYS[1])

if ttl < 0 then
    redis.call('set', KEYS[1], 0, 'EX', ARGV[2])
else
    redis.call('expire', KEYS[1], ARGV[2])
end

local current = tonumber(redis.call('get', KEYS[1]) or "0")
local quota = tonumber(ARGV[3])

if ARGV[1] == 'inc' then
    if current >= quota then
        return {quota, -1}
    else
        current = redis.call('incrby', KEYS[1], 1)
        return {quota, quota - current}
    end

elseif ARGV[1] == 'dec' then
    current = redis.call('decrby', KEYS[1], 1)
    if current < 0 then
        redis.call('set', KEYS[1], 0, 'EX', ARGV[2])
        current = 0
    end
end

return {quota, quota - current}
`

	LimitContextKey = "LimitContext" // 限流上下文信息

	RateLimitLimitHeader     = "X-RateLimit-Limit"     // 限制的总请求数
	RateLimitRemainingHeader = "X-RateLimit-Remaining" // 剩余还可以发送的请求数
	RateLimitResetHeader     = "X-RateLimit-Reset"     // 限流重置时间（触发限流时返回）
)

type LimitContext struct {
	count     int
	remaining int
	reset     int
}

type ConfigModelRateLimit struct {
	RuleName             string          // 限流规则名称
	RuleItems            []LimitRuleItem // 限流规则项
	ShowLimitQuotaHeader bool            // 响应头中是否显示X-RateLimit-Limit和X-RateLimit-Remaining
	RejectedCode         uint32          // 当请求超过阈值被拒绝时,返回的HTTP状态码
	RejectedMsg          string          // 当请求超过阈值被拒绝时,返回的响应体
	RedisClient          wrapper.RedisClient
}

type LimitRuleItem struct {
	Key   string
	Quota int
}

func main() {
	wrapper.SetCtx(
		"model-rate-limit",

		wrapper.ParseConfigBy(parseConfig),
		wrapper.ProcessRequestHeadersBy(onHttpRequestHeaders),
		wrapper.ProcessResponseHeadersBy(onHttpResponseHeaders),
		wrapper.ProcessStreamingResponseBodyBy(onHttpStreamingBody),
		//wrapper.ProcessStreamDoneBy(onHttpStreamDone),
	)
}

func parseConfig(json gjson.Result, cfg *ConfigModelRateLimit, log wrapper.Log) error {
	err := initRedisClusterClient(json, cfg)
	if err != nil {
		return err
	}
	err = parseModelRateLimitConfig(json, cfg)
	if err != nil {
		return err
	}
	log.Infof("model rate limit config: %+v", cfg)

	return nil
}

func onHttpRequestHeaders(ctx wrapper.HttpContext, cfg ConfigModelRateLimit, log wrapper.Log) types.Action {
	ctx.DisableReroute()

	path, method := ctx.Path(), ctx.Method()
	if method != "POST" {
		return types.ActionContinue
	}
	if path != "/v1/chat/completions" && path != "/v1/completion" {
		return types.ActionContinue
	}

	model, err := proxywasm.GetHttpRequestHeader(XAiModel)
	if err != nil {
		log.Warnf("get http request header [x-ai-model] failed: %v", err)
		return types.ActionContinue
	}

	quota, ok := queryQuota(model, cfg.RuleItems)
	if !ok {
		return types.ActionContinue
	}

	ctx.SetContext(XAiModel, model)

	err = inc(ctx, cfg, log, model, quota)
	if err != nil {
		log.Errorf("redis call failed: %v", err)
		return types.ActionContinue
	}

	return types.HeaderStopAllIterationAndWatermark
}

func onHttpResponseHeaders(ctx wrapper.HttpContext, cfg ConfigModelRateLimit, log wrapper.Log) types.Action {
	limitContext, ok := ctx.GetContext(LimitContextKey).(LimitContext)
	if !ok {
		return types.ActionContinue
	}

	if cfg.ShowLimitQuotaHeader {
		_ = proxywasm.ReplaceHttpResponseHeader(RateLimitLimitHeader, strconv.Itoa(limitContext.count))
		_ = proxywasm.ReplaceHttpResponseHeader(RateLimitRemainingHeader, strconv.Itoa(limitContext.remaining))
	}

	return types.ActionContinue
}

func onHttpStreamingBody(ctx wrapper.HttpContext, cfg ConfigModelRateLimit, data []byte, endOfStream bool, log wrapper.Log) []byte {
	if !endOfStream {
		return data
	}

	if !ctx.GetBoolContext(XAiRejected, false) {
		return data
	}

	model := ctx.GetStringContext(XAiModel, "")
	if model == "" {
		return data
	}

	quota, ok := queryQuota(model, cfg.RuleItems)
	if !ok {
		return data
	}
	err := dec(ctx, cfg, log, model, quota, "stream done")
	if err != nil {
		log.Errorf("redis call failed, stream done: %v", err)
		return data
	}

	return data
}

func onHttpStreamDone(ctx wrapper.HttpContext, cfg ConfigModelRateLimit, log wrapper.Log) {
	if !ctx.GetBoolContext(XAiRejected, false) {
		return
	}

	model := ctx.GetStringContext(XAiModel, "")
	if model == "" {
		return
	}

	quota, ok := queryQuota(model, cfg.RuleItems)
	if !ok {
		return
	}
	err := dec(ctx, cfg, log, model, quota, "stream done")
	if err != nil {
		log.Errorf("redis call failed, stream done: %v", err)
		return
	}
}

func initRedisClusterClient(json gjson.Result, config *ConfigModelRateLimit) error {
	redisConfig := json.Get("redis")
	if !redisConfig.Exists() {
		return errors.New("missing redis in config")
	}

	serviceName := redisConfig.Get("service_name").String()
	if serviceName == "" {
		return errors.New("redis service name must not be empty")
	}

	servicePort := int(redisConfig.Get("service_port").Int())
	if servicePort == 0 {
		if strings.HasSuffix(serviceName, ".static") {
			// use default logic port which is 80 for static service
			servicePort = 80
		} else {
			servicePort = 6379
		}
	}

	username := redisConfig.Get("username").String()
	password := redisConfig.Get("password").String()
	timeout := int(redisConfig.Get("timeout").Int())
	if timeout == 0 {
		timeout = 1000
	}

	config.RedisClient = wrapper.NewRedisClusterClient(wrapper.FQDNCluster{
		FQDN: serviceName,
		Port: int64(servicePort),
	})
	database := int(redisConfig.Get("database").Int())
	return config.RedisClient.Init(username, password, int64(timeout), wrapper.WithDataBase(database))
}

func parseModelRateLimitConfig(json gjson.Result, cfg *ConfigModelRateLimit) error {
	ruleName := json.Get("rule_name")
	if !ruleName.Exists() {
		return errors.New("missing rule_name in config")
	}
	cfg.RuleName = ruleName.String()

	showLimitQuotaHeader := json.Get("show_limit_quota_header")
	if showLimitQuotaHeader.Exists() {
		cfg.ShowLimitQuotaHeader = showLimitQuotaHeader.Bool()
	}

	rejectedCode := json.Get("rejected_code")
	if rejectedCode.Exists() {
		cfg.RejectedCode = uint32(rejectedCode.Uint())
	} else {
		cfg.RejectedCode = DefaultRejectedCode
	}

	rejectedMsg := json.Get("rejected_msg")
	if rejectedMsg.Exists() {
		cfg.RejectedMsg = rejectedMsg.String()
	} else {
		cfg.RejectedMsg = DefaultRejectedMsg
	}

	ruleItemsResult := json.Get("rule_items")
	hasRule := ruleItemsResult.Exists()
	if !hasRule {
		return errors.New("missing rule_items in config")
	}

	// 处理条件限流规则
	items := ruleItemsResult.Array()
	if len(items) == 0 {
		return errors.New("config rule_items cannot be empty")
	}

	var ruleItems []LimitRuleItem
	for _, item := range items {
		var ruleItem LimitRuleItem
		ruleItem.Key = item.Get("key").String()
		ruleItem.Quota = int(item.Get("quota").Int())

		if ruleItem.Key == "" || ruleItem.Quota == 0 {
			return errors.New("invalid rule_item")
		}

		ruleItems = append(ruleItems, ruleItem)
	}

	cfg.RuleItems = ruleItems
	return nil
}

func rejected(config ConfigModelRateLimit, context LimitContext, log wrapper.Log) {
	headers := make(map[string][]string)
	headers[RateLimitResetHeader] = []string{strconv.Itoa(context.reset)}
	if config.ShowLimitQuotaHeader {
		headers[RateLimitLimitHeader] = []string{strconv.Itoa(context.count)}
		headers[RateLimitRemainingHeader] = []string{strconv.Itoa(0)}
	}
	er := proxywasm.SendHttpResponseWithDetail(
		config.RejectedCode, "model-rate-limit.rejected", reconvertHeaders(headers), []byte(config.RejectedMsg), -1)
	if er != nil {
		log.Errorf("http rejected failed: %v", er)
	}
}

func queryQuota(model string, rules []LimitRuleItem) (int, bool) {
	for _, rule := range rules {
		if rule.Key == model {
			return rule.Quota, true
		}
	}
	return 0, false
}

func inc(ctx wrapper.HttpContext, cfg ConfigModelRateLimit, log wrapper.Log, model string, quota int) error {
	limitKey := fmt.Sprintf(ClusterRateLimitFormat, cfg.RuleName, model)

	// 执行限流逻辑
	keys := []interface{}{limitKey}
	args := []interface{}{"inc", 3600, quota}
	err := cfg.RedisClient.Eval(FixedWindowScript, 1, keys, args, func(response resp.Value) {
		resultArray := response.Array()
		if len(resultArray) != 2 {
			log.Errorf("redis response parse error, response: %v", response)
			if er := proxywasm.ResumeHttpRequest(); er != nil {
				log.Errorf("http resume 1 error: %v", er)
			}
			return
		}
		context := LimitContext{
			count:     resultArray[0].Integer(),
			remaining: resultArray[1].Integer(),
			reset:     3600,
		}
		if context.remaining < 0 {
			// 触发限流
			ctx.SetContext(XAiRejected, true)
			rejected(cfg, context, log)
		} else {
			ctx.SetContext(LimitContextKey, context)
			if er := proxywasm.ResumeHttpRequest(); er != nil {
				log.Errorf("http resume 2 error: %v", er)
			}
		}
	})
	return err
}

func dec(_ wrapper.HttpContext, cfg ConfigModelRateLimit, log wrapper.Log, model string, quota int, str string) error {

	log.Infof("dec rejected rate limit: %v", str)

	limitKey := fmt.Sprintf(ClusterRateLimitFormat, cfg.RuleName, model)
	keys := []interface{}{limitKey}
	args := []interface{}{"dec", 3600, quota}
	err := cfg.RedisClient.Eval(FixedWindowScript, 1, keys, args, func(_ resp.Value) {})
	return err
}

// reconvertHeaders headers: map[string][]string -> [][2]string
func reconvertHeaders(hs map[string][]string) [][2]string {
	var ret [][2]string
	for k, vs := range hs {
		for _, v := range vs {
			ret = append(ret, [2]string{k, v})
		}
	}
	sort.SliceStable(ret, func(i, j int) bool {
		return ret[i][0] < ret[j][0]
	})
	return ret
}
