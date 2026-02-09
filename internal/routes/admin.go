package routes

import (
	"net/http"

	"github.com/jkaninda/okapi"
)

func (r *Router) adminRoutes() []okapi.RouteDefinition {
	group := r.group.Group("/admin").WithTags([]string{"adminService"})
	group.Use(r.auth.JWT.Middleware)

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
	}
}
