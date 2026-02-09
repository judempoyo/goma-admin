package middlewares

import (
	"github.com/jkaninda/goma-admin/internal/config"
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
		ForwardClaims: map[string]string{
			"user_id": "sub",
			"email":   "email",
			"role":    "role",
		},
		ClaimsExpression: "Equals(`role`, `admin`)",
	}
	return &Auth{JWT: jwtAuth}
}
