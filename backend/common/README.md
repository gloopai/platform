# `github.com/gloopai/pay/common`

单 module、多子包；服务通过 `replace ../../common` 引用。

| 子包 | 说明 |
|------|------|
| `model/` | 跨服务共享领域类型（表行、Consul KV 快照 DTO 等） |
| `pb/` | `proto/` 生成的 gRPC / protobuf 代码（按服务分子目录） |
| `grpcclient/` | 各下游 RPC 的薄客户端封装 |
| `consulx/` | Consul 客户端、服务注册、全局 KV 前缀与路径 |
| `notify/` | 商户异步通知载荷与 NSQ 封装 |
| `healthx/` | 健康检查（含 gRPC health） |
| `dbdsn/` | MySQL DSN 与时区 |
| `timex/` | 进程时区 |
| `jwtutil/` | JWT 小工具 |
| `signmd5/` | OpenAPI / 商户回调：排序 `k=v` + `key=secret` 后 MD5 hex |

生成代码时只输出到 **`pb/`**，勿再增加与 `pb/*` 重复的生成目录。
