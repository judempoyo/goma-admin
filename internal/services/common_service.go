package services

import (
	util "github.com/jkaninda/goma-admin/utils"
	"github.com/jkaninda/okapi"
)

type CommonService struct{}

func (cm CommonService) Home(c *okapi.Context) error {
	return c.OK(okapi.M{"message": "Welcome to the Okapi Web Framework!"})
}
func (cm CommonService) Healthz(c *okapi.Context) error {
	return c.OK(okapi.M{"status": "healthy"})
}
func (cm CommonService) Readyz(c *okapi.Context) error {
	return c.OK(okapi.M{"status": "running"})
}

func (cm CommonService) Version(c *okapi.Context) error {
	return c.OK(okapi.M{"version": util.AppVersion})
}
func (cm CommonService) Dashboard(c *okapi.Context) error {
	return c.OK(okapi.M{"message": "Welcome to the Okapi Web Framework!"})
}
