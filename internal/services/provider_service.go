package services

import "github.com/jkaninda/okapi"

type ProviderService struct{}

func (s ProviderService) Provider(c *okapi.Context) error {
	return c.OK(okapi.M{"Status": "Ok"})
}
func (s ProviderService) Routes(c *okapi.Context) error {
	return c.OK(okapi.M{"Status": "Ok"})
}
func (s ProviderService) Middlewares(c *okapi.Context) error {
	return c.OK(okapi.M{"Status": "Ok"})
}
func (s ProviderService) Webhook(c *okapi.Context) error {
	return c.OK(okapi.M{"Status": "Ok"})
}
