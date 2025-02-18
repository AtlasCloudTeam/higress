package main

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
	"github.com/tidwall/sjson"

	"github.com/alibaba/higress/plugins/wasm-go/pkg/wrapper"
	"github.com/higress-group/proxy-wasm-go-sdk/proxywasm"
	"github.com/higress-group/proxy-wasm-go-sdk/proxywasm/types"
)

const (
	requestFillHeader = "request-fill-header"

	requestPath = "request-path"
)

func main() {
	wrapper.SetCtx(
		"ai-router",

		wrapper.ParseConfigBy(parseConfig),
		wrapper.ProcessRequestHeadersBy(onHttpRequestHeaders),
		wrapper.ProcessRequestBodyBy(onHttpRequestBody),
	)
}

type AIRouterConfig struct {
	routers []RouterConfig `yaml:"routers"`

	requestInfos map[string]requestInfo `yaml:"-"`
}

type requestInfo struct {
	methods       []string
	fillBody      bool
	destHeaderKey string
	fromBodyKey   string
}

type RouterConfig struct {
	destHeaderKey  string   `yaml:"dest_header_key"`
	fromBodyKey    string   `yaml:"from_body_key"`
	requestPath    string   `yaml:"request_path"`
	requestMethods []string `yaml:"request_methods"`
	streamOptions  bool     `yaml:"stream_options"`
}

func parseConfig(configJson gjson.Result, config *AIRouterConfig, log wrapper.Log) (err error) {
	routersInJson := configJson.Get("routers")
	if !routersInJson.Exists() {
		return errors.New("ai router config not exist in yaml")
	}

	config.routers, config.requestInfos, err = parseRouterConfig(routersInJson.Array())
	if err != nil {
		return errors.Wrap(err, "failed to parse router config")
	}

	log.Infof("ai router config is: routers: %+v", config.routers)

	return nil
}

func onHttpRequestHeaders(ctx wrapper.HttpContext, cfg AIRouterConfig, log wrapper.Log) types.Action {
	path, method := ctx.Path(), ctx.Method()

	if filterRequest(cfg, path, method) {
		ctx.DontReadRequestBody()
		return types.ActionContinue
	}

	ctx.SetContext(requestFillHeader, true)
	ctx.SetContext(requestPath, path)

	if cfg.requestInfos[path].fillBody {
		err := proxywasm.RemoveHttpRequestHeader("content-length")
		if err != nil {
			log.Warnf("remove http request header failed: %v", err)
		}
	}

	return types.HeaderStopIteration
}

func onHttpRequestBody(ctx wrapper.HttpContext, cfg AIRouterConfig, body []byte, log wrapper.Log) types.Action {
	if !ctx.GetBoolContext(requestFillHeader, false) {
		return types.ActionContinue
	}
	if len(body) == 0 {
		return types.ActionContinue
	}

	path := ctx.GetStringContext(requestPath, "")
	if path == "" {
		return types.ActionContinue
	}

	info, ok := cfg.requestInfos[path]
	if !ok {
		return types.ActionContinue
	}

	value := gjson.GetBytes(body, info.fromBodyKey).String()
	if value == "" {
		log.Warnf("failed to get body key field [%s] by request body", info.fromBodyKey)
		return types.ActionContinue
	}

	err := proxywasm.ReplaceHttpRequestHeader(info.destHeaderKey, value)
	if err != nil {
		log.Warnf("failed to replace request headers [%s]-[%s], err: %v", info.destHeaderKey, value, err)
		return types.ActionContinue
	}

	if !info.fillBody {
		return types.ActionContinue
	}

	data, err := fillRequestBody(body)
	if err != nil {
		log.Warnf("failed to fill request body: %v", err)
		return types.ActionContinue
	}

	if err = proxywasm.ReplaceHttpRequestBody(pretty.Pretty(data)); err != nil {
		log.Warnf("failed to replace request body: %v", err)
		return types.ActionContinue
	}

	return types.ActionContinue
}

func fillRequestBody(raw []byte) ([]byte, error) {
	body := append([]byte(nil), raw...)

	stream := gjson.GetBytes(body, "stream")
	// 不存在, 设置为 非流式请求
	if !stream.Exists() {
		return sjson.SetBytes(body, "stream", false)
	}
	// stream = false, 直接返回
	if !stream.Bool() {
		return body, nil
	}

	// stream = true, 改写 body
	body, err := sjson.SetBytes(body, "stream_options.include_usage", true)
	if err != nil {
		return body, fmt.Errorf("failed to set stream_options.include_usage: %v", err)
	}
	body, err = sjson.SetBytes(body, "stream_options.continuous_usage_stats", false)
	if err != nil {
		return body, fmt.Errorf("failed to set stream_options.continuous_usage_stats: %v", err)
	}
	return body, nil
}

// filterRequest 过滤不需要处理的Http请求, true:不需要处理, false:需要处理
func filterRequest(cfg AIRouterConfig, path string, method string) bool {
	// path 不存在, 说明不需要处理
	info, ok := cfg.requestInfos[path]
	if !ok {
		return true
	}

	// method 等于 *, 或者匹配, 说明需要处理
	for _, m := range info.methods {
		if m == "*" || m == method {
			return false
		}
	}
	return true
}

func parseRouterConfig(rules []gjson.Result) (res []RouterConfig, reqInfos map[string]requestInfo, err error) {
	reqInfos = make(map[string]requestInfo)
	for _, r := range rules {
		var router RouterConfig

		router.destHeaderKey = strings.ToLower(r.Get("dest_header_key").String())
		if router.destHeaderKey == "" {
			err = errors.Wrapf(err, "dest_header_key is required")
			return
		}

		router.fromBodyKey = r.Get("from_body_key").String()
		if router.fromBodyKey == "" {
			err = errors.Wrapf(err, "from_body_key is required")
			return
		}

		router.requestPath = strings.ToLower(r.Get("request_path").String())
		if router.requestPath == "" {
			err = errors.Wrapf(err, "request_path is required")
			return
		}

		for _, m := range r.Get("request_methods").Array() {
			router.requestMethods = append(router.requestMethods, strings.ToUpper(m.String()))
		}
		if len(router.requestMethods) == 0 {
			err = errors.Wrapf(err, "request_methods is required")
			return
		}

		router.streamOptions = r.Get("stream_options").Bool()

		_, ok := reqInfos[router.requestPath]
		if ok {
			err = errors.Wrapf(err, "request_path is duplicated")
			return
		}
		reqInfos[router.requestPath] = requestInfo{
			methods:       router.requestMethods,
			fillBody:      router.streamOptions,
			destHeaderKey: router.destHeaderKey,
			fromBodyKey:   router.fromBodyKey,
		}

		res = append(res, router)
	}
	return
}
