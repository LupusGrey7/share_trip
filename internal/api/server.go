package api

import (
	"job4j.ru/share_trip/internal/service"
)

type Server struct {
	InfoService *service.InfoService
	TripHandler *Handler
}

func NewServer(
	service *service.InfoService,
	tpHandler *Handler,
) *Server {
	return &Server{
		InfoService: service,
		TripHandler: tpHandler,
	}
}
