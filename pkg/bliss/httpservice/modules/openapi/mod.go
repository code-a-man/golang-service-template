package openapi

import (
	"github.com/eser/go-service/pkg/bliss/httpservice"
	"go.uber.org/fx"
)

type ApiIdentity struct {
	name    string
	version string
}

var Module = fx.Module( //nolint:gochecknoglobals
	"openapi",
	fx.Invoke(
		RegisterRoutes,
	),
)

func RegisterRoutes(routes *httpservice.Router) {
	routes.
		Route("GET /openapi.json", func(ctx *httpservice.Context) httpservice.Result {
			spec := &ApiIdentity{
				name:    "golang-service",
				version: "0.0.0",
			}

			result := GenerateOpenApiSpec(spec, routes)

			return ctx.Results.Json(result)
		}).
		HasSummary("OpenAPI Spec").
		HasDescription("OpenAPI Spec Endpoint")
}
