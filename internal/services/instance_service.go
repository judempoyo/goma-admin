package services

import "github.com/jkaninda/okapi"

type InstanceService struct{}

func (s InstanceService) List(c *okapi.Context) error {
	return c.OK(okapi.M{"Status": "Ok"})
}
func (s InstanceService) Create(c *okapi.Context) error {
	return c.OK(okapi.M{"Status": "Ok"})
}
func (s InstanceService) Get(c *okapi.Context) error {
	return c.OK(okapi.M{"Status": "Ok"})
}
func (s InstanceService) Update(c *okapi.Context) error {
	return c.OK(okapi.M{"Status": "Ok"})
}
func (s InstanceService) Delete(c *okapi.Context) error {
	return c.OK(okapi.M{"Status": "Ok"})
}
