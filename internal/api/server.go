package api

import (
	"job4j.ru/share_trip/internal/service"
)

type Server struct {
	InfoService        *service.InfoService
	TripService        *service.TripService
	CommandTripService *service.CommandTripService
	QueryTripService   *service.QueryTripService
}

func NewServer(
	service *service.InfoService,
	tripService *service.TripService,
	commandTripService *service.CommandTripService,
	queryTripService *service.QueryTripService,
) *Server {
	return &Server{
		InfoService:        service,
		TripService:        tripService,
		CommandTripService: commandTripService,
		QueryTripService:   queryTripService,
	}
}
