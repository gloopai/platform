package adminshell

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest"
)

// RegisterInfraRoutes registers GET /ready and GET /v1/admin/ops/services (behind adminAuth).
// Readiness and ops/service discovery handlers are supplied by the gateway binary (svc wiring stays app-specific).
func RegisterInfraRoutes(server *rest.Server, adminAuth rest.Middleware, ready, opsServices http.HandlerFunc) {
	server.AddRoutes([]rest.Route{
		{Method: http.MethodGet, Path: "/ready", Handler: ready},
	})
	server.AddRoutes(rest.WithMiddlewares(
		[]rest.Middleware{adminAuth},
		[]rest.Route{
			{Method: http.MethodGet, Path: "/v1/admin/ops/services", Handler: opsServices},
		}...,
	))
}
