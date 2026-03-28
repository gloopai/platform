// Package model holds shared domain types used across pay-platform services (core, trade, gateway, …).
//
// Conventions:
//
//   - Table rows / scan targets: structs carry both json (snake_case, omitempty) and gorm column tags
//     where useful for GORM Scan/Take. Secrets use json:"-" where they must not leak in HTTP JSON.
//
//   - Consul KV snapshots (*KV suffix): JSON-only blobs under pay/config/global/.../snapshot/;
//     no gorm tags. Written on admin save; mirrored in process memory for hot paths.
//
//   - Naming: *KV = serialized snapshot for Consul; plain names = DB row or admin query shape.
//
// New shared types belong here rather than duplicated per service.
package model
