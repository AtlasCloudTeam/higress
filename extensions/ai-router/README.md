---
title: AI模型路由
keywords: [ higress, AI, router ]
description: AI模型路由配置参考
---

## 介绍

结合higress提供的请求头路由能力，通过解析HTTP请求体字段，配置HTTP Header，实现自定义路由功能；同时也提供了流式请求的body填充选项。

## 运行属性

插件执行阶段：`认证阶段`

插件执行优先级：`410`

## 配置示例

插件默认请求符合openai协议格式，示例如下：

```yaml
routers:
  - request_path: "/v1/chat/completions" # 匹配 Http Path, 不可为空
    request_methods:                     # 匹配 Http Method List (支持使用 * 代表所有Method), 不可为空
      - "POST"
    stream_options: true                # 是否开启 Stream Options 填充, true:填充,  非必填
    route_maps:                         # 路由映射, 不可为空
      - dest_header_key: "x-ai-model"
        from_body_key: "model"
      - dest_header_key: "x-web-search" # 目的请求头的Key, 不可为空
        from_body_key: "web_search"     # 来源请求体的Key, 不可为空
        default_value: "false"          # 默认值
        support_models:                 # 支持的模型列表, 为空时说明支持所有模型, 不为空时则会匹配请求模型, 匹配则使用请求体的值, 不匹配则使用默认值
          - deepseek-ai/DeepSeek-R1
```

上述配置说明:

1. 匹配 `POST` `/v1/chat/completions` 请求, 其他请求不做任何处理

2. `stream_options: true` 说明:
    - 当请求体字段 `"stream": true` 时, 会进行请求体填充
      `"stream_options.include_usage": true, "stream_options.continuous_usage_stats": false`
    - 当请求体字段 `"stream": false` 时, 不做任何处理
    - 当请求体字段 `"stream"` 不存在时, 会进行请求体填充 `"stream": false`
   
3. `route_maps` 说明
    - 读取请求体字段 `"model"`, 并将其值填充到请求头 `"x-ai-model"` 中
    - 读取请求体字段 `"web_search"`, 并将其值填充到请求头 `"x-web-search"` 中
        - 当请求体字段 `"web_search"` 不存在时, 使用默认值配置 `"default_value"` 的值填充到请求头 `"x-web-search"` 中
        - 当请求体字段 `"model"` 不等于 `deepseek-ai/DeepSeek-R1` 时, 也使用默认值配置 `"default_value"` 的值填充到请求头
          `"x-web-search"` 中


## 构建说明

1. 环境准备(以x86为例)

- go version: 1.20.14
- tinygo version: 0.29.0

2. 构建wasm
```go
go mod tidy
tinygo build -o plugin.wasm -scheduler=none -target=wasi -gc=custom -tags="custommalloc nottinygc_finalizer proxy_wasm_version_0_2_100" ./
```

3. 构建镜像并推送
```dockerfile
docker build -f Dockerfile -t {repo}:{tag} .

docker push {repo}:{tag}
```