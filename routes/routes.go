package routes

import (
	"context"
	"net/http"

	"github.com/jkaninda/goma-admin/config"
	"github.com/jkaninda/goma-admin/middlewares"
	"github.com/jkaninda/goma-admin/services"
	"github.com/jkaninda/okapi"
)

type Route struct {
	app    *okapi.Okapi
	config *config.Config
	cxt    context.Context
	group  *okapi.Group
	auth   *middlewares.Auth
}

var (
	commonService     = &services.CommonService{}
	routeService      = &services.RouteService{}
	providerService   = &services.ProviderService{}
	middlewareService = &services.MiddlewareService{}
	authService       *services.AuthService
	adminService      *services.AdminService
)

func NewRoute(ctx context.Context, app *okapi.Okapi, conf *config.Config) *Route {
	authService = services.NewAuthService(conf)
	adminService = services.NewAdminService(conf)
	return &Route{
		app:    app,
		config: conf,
		cxt:    ctx,
		group:  &okapi.Group{Prefix: "api/v1"},
		auth:   middlewares.NewAuth(conf),
	}
}
func (r *Route) RegisterRoutes() {
	r.app.Register(r.home())
	r.app.Register(r.Version())
	r.app.Register(r.Routes()...)
	r.app.Register(r.providerRoutes()...)
	r.app.Register(r.routeMiddlewares()...)
	r.app.Register(r.authRoutes()...)
	r.app.Register(r.adminRoutes()...)
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
	group.Use(r.auth.JWT.Middleware, middlewares.RequireRoles("admin"))
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
	group.Use(r.auth.JWT.Middleware, middlewares.RequireRoles("admin"))
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

func (r *Route) adminRoutes() []okapi.RouteDefinition {
	group := r.group.Group("/admin").WithTags([]string{"adminService"})
	group.Use(r.auth.JWT.Middleware, middlewares.RequireRoles("admin"))

	return []okapi.RouteDefinition{
		{
			Path:    "/users",
			Method:  http.MethodGet,
			Handler: adminService.ListUsers,
			Group:   group,
		},
		{
			Path:    "/users/:id",
			Method:  http.MethodGet,
			Handler: adminService.GetUser,
			Group:   group,
		},
		{
			Path:    "/users/:id/roles",
			Method:  http.MethodPut,
			Handler: adminService.UpdateUserRoles,
			Group:   group,
		},
		{
			Path:    "/roles",
			Method:  http.MethodGet,
			Handler: adminService.ListRoles,
			Group:   group,
		},
		{
			Path:    "/roles",
			Method:  http.MethodPost,
			Handler: adminService.CreateRole,
			Group:   group,
		},
	}
}

func (r *Route) authRoutes() []okapi.RouteDefinition {
	group := r.group.Group("/auth").WithTags([]string{"authService"})
	protected := group.Group("").WithTags([]string{"authService"})
	protected.Use(r.auth.JWT.Middleware)

	return []okapi.RouteDefinition{
		{
			Path:    "/register",
			Method:  http.MethodPost,
			Handler: authService.Register,
			Group:   group,
		},
		{
			Path:    "/login",
			Method:  http.MethodPost,
			Handler: authService.Login,
			Group:   group,
		},
		{
			Path:    "/refresh",
			Method:  http.MethodPost,
			Handler: authService.Refresh,
			Group:   group,
		},
		{
			Path:    "/logout",
			Method:  http.MethodPost,
			Handler: authService.Logout,
			Group:   group,
		},
		{
			Path:    "/me",
			Method:  http.MethodGet,
			Handler: authService.Me,
			Group:   protected,
		},
		{
			Path:    "/oauth/:provider",
			Method:  http.MethodGet,
			Handler: authService.OAuthStart,
			Group:   group,
		},
		{
			Path:    "/oauth/:provider/callback",
			Method:  http.MethodGet,
			Handler: authService.OAuthCallback,
			Group:   group,
		},
	}
}
