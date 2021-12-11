package handler

import (
	"github.com/21hack02win/nascalay-backend/oapi"
	"github.com/21hack02win/nascalay-backend/usecases/repository"
	"github.com/21hack02win/nascalay-backend/usecases/service/ws"
)

type handler struct {
	r      repository.Repository
	stream ws.Streamer
}

func NewHandler(r repository.Repository, stream ws.Streamer) oapi.ServerInterface {
	return &handler{r, stream}
}
