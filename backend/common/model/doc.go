// Package model holds shared domain types used across pay-platform services (core, trade, gateway, …).
//
// Conventions:
//
//   - Table rows / scan targets: structs carry both json (snake_case, omitempty) and gorm column tags
//     where useful for GORM Scan/Take. Secrets use json:"-" where they must not leak in HTTP JSON.
//
//   - Consul KV snapshot JSON blobs and key paths live in package configkv (not here).
//
// New shared types belong here rather than duplicated per service.
package model
