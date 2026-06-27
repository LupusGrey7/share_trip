package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/prometheus/client_golang/prometheus"
	"job4j.ru/share_trip/internal/service"
)

type Server struct {
	registry    *prometheus.Registry
	validator   *validator.Validate
	InfoService *service.InfoService
	TripService *service.TripService
}

func NewServer(
	m *prometheus.Registry,
	vl *validator.Validate,
	service *service.InfoService,
	tripService *service.TripService,
) *Server {
	return &Server{
		registry:    m,
		validator:   vl,
		InfoService: service,
		TripService: tripService,
	}
}
