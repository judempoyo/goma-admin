package services

import (
	"github.com/jkaninda/goma-admin/internal/config"
	"github.com/jkaninda/goma-admin/internal/dto"
	"github.com/jkaninda/okapi"
)

type AuthService struct {
}

func NewAuthService(conf *config.Config) *AuthService {
	return &AuthService{}
}

func (s *AuthService) Login(c *okapi.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.AbortBadRequest("Invalid request", err)
	}

	return c.OK(nil)

}

func (s *AuthService) Logout(c *okapi.Context) error {

	return c.OK(okapi.M{"status": "ok"})
}
