package services

import "github.com/jkaninda/okapi"

type MiddlewareService struct{}

func (m MiddlewareService) List(c okapi.C) error {
	return c.OK(okapi.M{"Status": "Ok"})
}
func (m MiddlewareService) Create(c okapi.C) error {
	return c.OK(okapi.M{"Status": "Ok"})
}
func (m MiddlewareService) Get(c okapi.C) error {
	return c.OK(okapi.M{"Status": "Ok"})
}
func (m MiddlewareService) Update(c okapi.C) error {
	return c.OK(okapi.M{"Status": "Ok"})
}
func (m MiddlewareService) Delete(c okapi.C) error {
	return c.OK(okapi.M{"Status": "Ok"})
}
