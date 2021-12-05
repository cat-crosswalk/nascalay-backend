package repository

import (
	"github.com/21hack02win/nascalay-backend/model"
	"github.com/21hack02win/nascalay-backend/usecases/repository"
)

type storeRepository struct {
	Room map[string]*model.Room
}

func NewRepository() repository.Repository {
	return &storeRepository{
		Room: make(map[string]*model.Room),
	}
}
