package handler

import (
	"github.com/21hack02win/nascalay-backend/usecases/repository"
	"github.com/21hack02win/nascalay-backend/usecases/service"
	"github.com/21hack02win/nascalay-backend/usecases/service/ws"
)

type handler struct {
	s      service.Service
	r      repository.Repository
	stream ws.Streamer
}

func NewHandler(s service.Service, r repository.Repository, stream ws.Streamer) ServerInterface {
	return &handler{s, r, stream}
}
