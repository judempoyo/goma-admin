package services

import "github.com/jkaninda/okapi"

type RouteService struct{}

func (r RouteService) List(c okapi.C) error {
	return c.OK(okapi.M{"Status": "Ok"})
}
func (r RouteService) Create(c okapi.C) error {
	return c.Created(okapi.M{"Status": "Created"})
}
func (r RouteService) Get(c okapi.C) error {
	return c.OK(okapi.M{"Status": "Ok"})
}
func (r RouteService) Update(c okapi.C) error {
	return c.OK(okapi.M{"Status": "Ok"})
}
func (r RouteService) Delete(c okapi.C) error {
	return c.OK(okapi.M{"Status": "Ok"})
}
