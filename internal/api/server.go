package api

import (
	"job4j.ru/share_trip/internal/service"
)

type Server struct {
	Service *service.Service
}

func NewServer(service *service.Service) *Server {
	return &Server{Service: service}
}
