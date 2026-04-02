// Package adminshell documents the admin API split for product repos (pay-platform, ec-platform).
//
// The HTTP route registration for the **platform shell** (login, RBAC, admin users, display, op_logs, jobs)
// lives in internal/handler/routes.go RegisterAdminHandlers.
//
// Product gateways should mirror that split: shell routes + product-only routes (see pay-platform
// gateway internal/handler routes_admin_shell.go and routes_admin_pay.go).
package adminshell
