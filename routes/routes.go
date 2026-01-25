package routes

import (
	"context"
	"net/http"

	"github.com/jkaninda/goma-admin/config"
	"github.com/jkaninda/goma-admin/services"
	"github.com/jkaninda/okapi"
)

type Route struct {
	app    *okapi.Okapi
	config *config.Config
	cxt    context.Context
	group  *okapi.Group
}

var (
	commonService     = &services.CommonService{}
	routeService      = &services.RouteService{}
	providerService   = &services.ProviderService{}
	middlewareService = &services.MiddlewareService{}
)

func NewRoute(ctx context.Context, app *okapi.Okapi, conf *config.Config) *Route {
	return &Route{
		app:    app,
		config: conf,
		cxt:    ctx,
		group:  &okapi.Group{Prefix: "api/v1"},
	}
}
func (r *Route) RegisterRoutes() {
	r.app.Register(r.home())
	r.app.Register(r.Version())
	r.app.Register(r.Routes()...)
	r.app.Register(r.providerRoutes()...)
	r.app.Register(r.routeMiddlewares()...)
}

func (r *Route) home() okapi.RouteDefinition {
	return okapi.RouteDefinition{
		Path:    "/",
		Method:  http.MethodGet,
		Handler: commonService.Home,
		Group:   &okapi.Group{Prefix: "/", Tags: []string{"commonService"}},
	}
}
func (r *Route) Version() okapi.RouteDefinition {
	return okapi.RouteDefinition{
		Path:    "/version",
		Method:  http.MethodGet,
		Handler: commonService.Version,
		Group:   &okapi.Group{Prefix: "/", Tags: []string{"commonService"}},
	}
}

func (r *Route) Routes() []okapi.RouteDefinition {
	group := r.group.Group("/routes").WithTags([]string{"routeService"})
	return []okapi.RouteDefinition{
		{
			Path:    "",
			Method:  http.MethodGet,
			Handler: routeService.List,
			Group:   group,
		},
		{
			Path:    "/:id",
			Method:  http.MethodPost,
			Handler: routeService.Create,
			Group:   group,
		},
		{
			Path:    "/:id",
			Method:  http.MethodGet,
			Handler: routeService.Get,
			Group:   group,
		},
		{
			Path:    "/:id",
			Method:  http.MethodPut,
			Handler: routeService.Update,
			Group:   group,
		},
		{
			Path:    "/:id",
			Method:  http.MethodDelete,
			Handler: routeService.Delete,
			Group:   group,
		},
	}
}
func (r *Route) routeMiddlewares() []okapi.RouteDefinition {
	group := r.group.Group("/middlewares").WithTags([]string{"middlewareService"})
	return []okapi.RouteDefinition{
		{
			Path:    "",
			Method:  http.MethodGet,
			Handler: middlewareService.List,
			Group:   group,
		},
		{
			Path:    "/:id",
			Method:  http.MethodPost,
			Handler: middlewareService.Create,
			Group:   group,
		},
		{
			Path:    "/:id",
			Method:  http.MethodGet,
			Handler: middlewareService.Get,
			Group:   group,
		},
		{
			Path:    "/:id",
			Method:  http.MethodPut,
			Handler: middlewareService.Update,
			Group:   group,
		},
		{
			Path:    "/:id",
			Method:  http.MethodDelete,
			Handler: middlewareService.Delete,
			Group:   group,
		},
	}
}
func (r *Route) providerRoutes() []okapi.RouteDefinition {
	group := r.group.Group("/provider").WithTags([]string{"providerService"})
	return []okapi.RouteDefinition{
		{
			Path:    "/:name",
			Method:  http.MethodGet,
			Handler: providerService.Provider,
			Group:   group,
		},
		{
			Path:    "/:name/middlewares",
			Method:  http.MethodGet,
			Handler: providerService.Middlewares,
			Group:   group,
		},
		{
			Path:    "/:name/routes",
			Method:  http.MethodGet,
			Handler: providerService.Routes,
			Group:   group,
		},
		{
			Path:    "/:name/webhook",
			Method:  http.MethodPost,
			Handler: providerService.Webhook,
			Group:   group,
		},
	}
}
