// Package adminshell registers the platform admin shell HTTP routes (login, RBAC, users, display, op_logs, jobs).
//
// Product gateways should call [Register] with their middlewares and handler bundle; see pay-platform
// gateway internal/handler/shell_register.go for an example.
package adminshell
