package api

import (
	"github.com/go-playground/validator/v10"
	"job4j.ru/share_trip/internal/service"
)

type Server struct {
	validator   *validator.Validate
	InfoService *service.InfoService
	TripService *service.TripService
}

func NewServer(
	vl *validator.Validate,
	service *service.InfoService,
	tripService *service.TripService,
) *Server {
	return &Server{
		validator:   vl,
		InfoService: service,
		TripService: tripService,
	}
}
