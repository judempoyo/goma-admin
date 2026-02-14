package routes

import (
	"github.com/jkaninda/okapi"
)

func (r *Router) authRoutes() []okapi.RouteDefinition {
	group := r.group.Group("/auth").WithTags([]string{"authService"})
	group.Use(r.auth.JWT.Middleware)

	return []okapi.RouteDefinition{}
}
