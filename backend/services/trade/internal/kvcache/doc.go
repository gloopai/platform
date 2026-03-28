// Package kvcache mirrors Consul KV in memory for OpenAPI / 收银台热路径（与库表双写）。
//
// 按实体拆文件扩展即可，例如 channel_config.go（通道配置 JSON）、后续 merchant_*.go、product_*.go 等。
// 管理台列表/详情仍以数据库为准；仅下单链路使用此处缓存。
package kvcache
