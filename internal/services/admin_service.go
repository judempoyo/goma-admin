package services

import (
	"github.com/jkaninda/goma-admin/internal/config"
	"github.com/jkaninda/okapi"
)

type AdminService struct {
}

func NewAdminService(conf *config.Config) *AdminService {
	return &AdminService{}
}

func (s *AdminService) ListUsers(c *okapi.Context) error {
	return c.OK(nil)
}

func (s *AdminService) GetUser(c *okapi.Context) error {

	return c.OK(nil)

}
