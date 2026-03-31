package api

import (
	"job4j.ru/share_trip/internal/service"
)

type Server struct {
	InfoService        *service.InfoService
	CommandTripService *service.CommandTripService
	QueryTripService   *service.QueryTripService
}

func NewServer(
	service *service.InfoService,
	commandTripService *service.CommandTripService,
	queryTripService *service.QueryTripService,
) *Server {
	return &Server{
		InfoService:        service,
		CommandTripService: commandTripService,
		QueryTripService:   queryTripService,
	}
}
