# atlascloud higress 配置

---

## atlascloud 使用的 higress 插件

目前我们使用到的插件列表如下：

| 插件名                                                                                                                      | 类型           | 作用维度                 | 功能说明                                                                                                                                              | 
|--------------------------------------------------------------------------------------------------------------------------|--------------|----------------------|---------------------------------------------------------------------------------------------------------------------------------------------------| 
| [ai-router](https://github.com/AtlasCloudTeam/higress/tree/main/extensions/ai-router)                                    | 自定义插件        | 域名 api.atlascloud.ai | 路由请求，请求体填充等                                                                                                                                       | 
| [ai-search](https://github.com/alibaba/higress/tree/main/plugins/wasm-go/extensions/ai-search)                           | higress 内置插件 | 域名 api.atlascloud.ai | 联网搜索                                                                                                                                              |
| [ai-statistics](https://github.com/alibaba/higress/tree/main/plugins/wasm-go/extensions/ai-statistics)                   | higress 内置插件 | 域名 api.atlascloud.ai | 记录input,output信息，用于计费，[**需要修改higress配置文件**](https://github.com/alibaba/higress/blob/main/plugins/wasm-go/extensions/ai-statistics/README.md#配置示例) |
| [cluster-key-rate-limit](https://github.com/alibaba/higress/tree/main/plugins/wasm-go/extensions/cluster-key-rate-limit) | higress 内置插件 | 域名 api.atlascloud.ai | 请求限流                                                                                                                                              |
| [cors](https://github.com/alibaba/higress/tree/main/plugins/wasm-go/extensions/cors)                                     | higress 内置插件 | 域名 api.atlascloud.ai | 跨域                                                                                                                                                |
| [ext-auth](https://github.com/alibaba/higress/tree/main/plugins/wasm-go/extensions/ext-auth)                             | higress 内置插件 | 域名 api.atlascloud.ai | 请求认证                                                                                                                                              |

资源清单路径为 [manifests](./manifests) 中以`plugin_`开头的文件，新增集群时，可直接使用资源清单部署


## atlascloud 新增模型，注册higress资源清单

目前我们支持三个模型，资源清单为 `mcpbridge_` 开头和 `ingress_` 开头的文件，新增集群时，可[**参考**]()资源清单部署

## 自定义插件ai-router 构建流程

[构建说明文档](https://github.com/AtlasCloudTeam/higress/blob/main/extensions/ai-router/README.md#构建说明)


