// Package configkv defines Consul KV path helpers and JSON snapshot types for global pay-platform
// configuration under pay/config/global/.... Per-service overlays use ServiceConfigPrefix.
//
// Snapshot structs are transport-only (JSON tags, no GORM); domain table rows remain in package model.
package configkv
