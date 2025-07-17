package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/alibaba/higress/plugins/wasm-go/pkg/wrapper"
	"github.com/higress-group/proxy-wasm-go-sdk/proxywasm"
	"github.com/higress-group/proxy-wasm-go-sdk/proxywasm/types"
)

const (
	defaultBodySize    = 32 * 1024
	defaultSupportSize = 32 * 1024 * 1024

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
	routers   []RouterConfig `yaml:"routers"`
	modelMaps []ModelMap     `yaml:"model_map"`

	requestInfos map[string]requestInfo `yaml:"-"`
}

type requestInfo struct {
	methods    []string
	routerMaps []RouterMap
	fillBody   bool
}

type RouterConfig struct {
	requestPath    string      `yaml:"request_path"`
	requestMethods []string    `yaml:"request_methods"`
	routerMaps     []RouterMap `yaml:"route_maps"`
	streamOptions  bool        `yaml:"stream_options"`
}

type ModelMap struct {
	Raw string `yaml:"raw"`
	Dst string `yaml:"dst"`
}

type RouterMap struct {
	destHeaderKey string   `yaml:"dest_header_key"`
	fromBodyKey   string   `yaml:"from_body_key"`
	defaultValue  string   `yaml:"default_value"`
	supportModels []string `yaml:"support_models"`
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

	mapInJson := configJson.Get("model_map")
	if mapInJson.Exists() {
		config.modelMaps = parseModelMapConfig(mapInJson.Array())
	}

	log.Infof("ai router config is: routers: %+v, modelmap: %+v", config.routers, config.modelMaps)

	return nil
}

func onHttpRequestHeaders(ctx wrapper.HttpContext, cfg AIRouterConfig, log wrapper.Log) types.Action {
	path, method := ctx.Path(), ctx.Method()
	if filterRequest(cfg, path, method) {
		ctx.DontReadRequestBody()
		return types.ActionContinue
	}

	length, err := getHttpHeaderContentLength()
	if err != nil {
		log.Errorf("failed to get http header content-length: %v", err)
		ctx.DontReadRequestBody()
		return types.ActionContinue
	}
	if length >= defaultBodySize {
		ctx.SetRequestBodyBufferLimit(defaultSupportSize)
	}

	if cfg.requestInfos[path].fillBody {
		err = proxywasm.RemoveHttpRequestHeader("content-length")
		if err != nil {
			log.Warnf("remove http request header failed: %v", err)
			ctx.DontReadRequestBody()
			return types.ActionContinue
		}
	}

	ctx.SetContext(requestFillHeader, true)
	ctx.SetContext(requestPath, path)

	return types.HeaderStopIteration
}

func onHttpRequestBody(ctx wrapper.HttpContext, cfg AIRouterConfig, body []byte, log wrapper.Log) types.Action {
	if !ctx.GetBoolContext(requestFillHeader, false) {
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

	model := gjson.GetBytes(body, "model").String()
	if model == "" {
		log.Warn("failed to get model field by request body")
		return types.ActionContinue
	}

	webSearch := gjson.GetBytes(body, "web_search").Bool()

	for _, rm := range info.routerMaps {
		var value string
		if rm.fromBodyKey == "model" {
			value = model
		} else {
			value = gjson.GetBytes(body, rm.fromBodyKey).String()
			value = supportWebSearch(rm.supportModels, model, value, rm.defaultValue)
		}

		err := proxywasm.ReplaceHttpRequestHeader(rm.destHeaderKey, value)
		if err != nil {
			log.Warnf("failed to replace request headers [%s]-[%s], err: %v", rm.destHeaderKey, value, err)
			return types.ActionContinue
		}
	}

	if !info.fillBody {
		return types.ActionContinue
	}

	dstModel := getDstModel(model, cfg.modelMaps)

	data, err := fillRequestBody(body, dstModel, webSearch)
	if err != nil {
		log.Warnf("failed to fill request body: %v", err)
		return types.ActionContinue
	}

	if err = proxywasm.ReplaceHttpRequestBody(data); err != nil {
		log.Warnf("failed to replace request body: %v", err)
		return types.ActionContinue
	}

	return types.ActionContinue
}

func fillRequestBody(raw []byte, dstModel string, webSearch bool) ([]byte, error) {
	body := append([]byte(nil), raw...)
	var err error

	if webSearch {
		body, err = sjson.SetBytes(body, "web_search_options.search_context_size", "high")
		if err != nil {
			return body, fmt.Errorf("failed to set web_search_options.search_context_size: %v", err)
		}
	}

	if dstModel != "" {
		body, err = sjson.SetBytes(body, "model", dstModel)
		if err != nil {
			return body, fmt.Errorf("failed to set model: %v", err)
		}
	}

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
	body, err = sjson.SetBytes(body, "stream_options.include_usage", true)
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

func getDstModel(model string, maps []ModelMap) string {
	if len(maps) == 0 {
		return ""
	}
	for _, m := range maps {
		if m.Raw == model {
			return m.Dst
		}
	}
	return ""
}

func getHttpHeaderContentLength() (int64, error) {
	contentLength, err := proxywasm.GetHttpRequestHeader("content-length")
	if err != nil {
		return 0, err
	}
	contentLength = strings.TrimSpace(contentLength)
	if contentLength == "" {
		return 0, errors.New("content-length header is empty")
	}
	contentLengthInt, err := strconv.ParseInt(contentLength, 10, 64)
	if err != nil {
		return 0, err
	}

	if contentLengthInt <= 0 {
		return 0, errors.New("content-length header is zero")
	}

	return contentLengthInt, nil
}

func parseModelMapConfig(rules []gjson.Result) (res []ModelMap) {
	for _, rule := range rules {
		var modelMap ModelMap

		modelMap.Raw = rule.Get("raw").String()
		modelMap.Dst = rule.Get("dst").String()
		res = append(res, modelMap)
	}
	return res
}

func parseRouterConfig(rules []gjson.Result) (res []RouterConfig, reqInfos map[string]requestInfo, err error) {
	reqInfos = make(map[string]requestInfo)
	for _, r := range rules {
		var router RouterConfig

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

		for _, routeMap := range r.Get("route_maps").Array() {
			var models []string
			for _, m := range routeMap.Get("support_models").Array() {
				models = append(models, m.String())
			}

			kv := RouterMap{
				destHeaderKey: strings.ToLower(routeMap.Get("dest_header_key").String()),
				fromBodyKey:   strings.ToLower(routeMap.Get("from_body_key").String()),
				defaultValue:  strings.ToLower(routeMap.Get("default_value").String()),
				supportModels: models,
			}

			if kv.fromBodyKey == "" || kv.destHeaderKey == "" {
				err = errors.Wrapf(err, "dest_header_key or from_body_key is required")
				return
			}
			router.routerMaps = append(router.routerMaps, kv)
		}
		if len(router.routerMaps) == 0 {
			err = errors.Wrapf(err, "route_maps is required")
			return
		}

		router.streamOptions = r.Get("stream_options").Bool()

		_, ok := reqInfos[router.requestPath]
		if ok {
			err = errors.Wrapf(err, "request_path is duplicated")
			return
		}

		reqInfos[router.requestPath] = requestInfo{
			methods:    router.requestMethods,
			fillBody:   router.streamOptions,
			routerMaps: router.routerMaps,
		}

		res = append(res, router)
	}
	return
}

func supportWebSearch(supportModels []string, requestModel string, requestValue, defaultValue string) string {
	if requestValue == "" {
		return defaultValue
	}

	if len(supportModels) <= 0 {
		return requestValue
	}

	for _, model := range supportModels {
		if model == requestModel {
			return requestValue
		}
	}
	return defaultValue
}
