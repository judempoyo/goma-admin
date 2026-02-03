package middlewares

import (
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jkaninda/goma-admin/config"
	"github.com/jkaninda/okapi"
)

const (
	authClaimsKey = "auth_claims"
)

type Auth struct {
	JWT *okapi.JWTAuth
}

func NewAuth(conf *config.Config) *Auth {
	jwtAuth := &okapi.JWTAuth{
		SigningSecret: []byte(conf.JWT.Secret),
		TokenLookup:   "header:Authorization",
		Audience:      conf.JWT.Audience,
		Issuer:        conf.JWT.Issuer,
		ContextKey:    authClaimsKey,
		ForwardClaims: map[string]string{
			"user_id": "sub",
			"email":   "email",
			"roles":   "roles",
		},
	}
	return &Auth{JWT: jwtAuth}
}

func RequireRoles(roles ...string) okapi.Middleware {
	required := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		role = strings.ToLower(strings.TrimSpace(role))
		if role == "" {
			continue
		}
		required[role] = struct{}{}
	}
	return func(next okapi.HandlerFunc) okapi.HandlerFunc {
		return func(c *okapi.Context) error {
			claimsValue, ok := c.Get(authClaimsKey)
			if !ok {
				return c.AbortForbidden("Missing authentication claims")
			}
			mapClaims, ok := claimsValue.(jwt.MapClaims)
			if !ok {
				return c.AbortForbidden("Invalid authentication claims")
			}
			rolesClaim := extractRoles(mapClaims)
			if !hasAnyRole(rolesClaim, required) {
				return c.AbortForbidden("Insufficient permissions")
			}
			return next(c)
		}
	}
}

func extractRoles(claims jwt.MapClaims) []string {
	var roles []string
	if raw, ok := claims["roles"]; ok {
		switch value := raw.(type) {
		case []interface{}:
			for _, v := range value {
				roles = append(roles, strings.ToLower(strings.TrimSpace(toString(v))))
			}
		case []string:
			for _, v := range value {
				roles = append(roles, strings.ToLower(strings.TrimSpace(v)))
			}
		case string:
			for _, v := range strings.Split(value, ",") {
				roles = append(roles, strings.ToLower(strings.TrimSpace(v)))
			}
		default:
			roles = append(roles, strings.ToLower(strings.TrimSpace(toString(value))))
		}
	}
	if raw, ok := claims["role"]; ok {
		roles = append(roles, strings.ToLower(strings.TrimSpace(toString(raw))))
	}
	return roles
}

func toString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return ""
	}
}

func hasAnyRole(userRoles []string, required map[string]struct{}) bool {
	if len(required) == 0 {
		return true
	}
	for _, role := range userRoles {
		if role == "" {
			continue
		}
		if _, ok := required[role]; ok {
			return true
		}
	}
	return false
}
