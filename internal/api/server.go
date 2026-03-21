package api

import (
	"job4j.ru/share_trip/internal/domain"
)

type Server struct {
	Service *domain.Service
}

func NewServer(service *domain.Service) *Server {
	return &Server{Service: service}
}
