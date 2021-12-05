package handler

import (
	"github.com/21hack02win/nascalay-backend/usecases/repository"
	"github.com/21hack02win/nascalay-backend/usecases/service"
)

type handler struct {
	s service.Service
	r repository.Repository
}

func NewHandler(s service.Service, r repository.Repository) ServerInterface {
	return &handler{s, r}
}
