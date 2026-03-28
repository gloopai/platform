// Package kvcache mirrors Consul KV in memory for OpenAPI / 收银台热路径（与库表双写）。
// KV 值为整行快照 JSON（见 common/model 中 *KV 类型），内存中反序列化为结构体；Pick* 在热路径上从快照取字段并回退库表。
// 管理台列表/详情仍以数据库为准。
package kvcache
