package repository

import (
	"github.com/21hack02win/nascalay-backend/model"
	"github.com/21hack02win/nascalay-backend/usecases/repository"
)

type storeRepository struct {
	room           map[model.RoomId]*model.Room
	userIdToRoomId map[model.UserId]model.RoomId
}

func NewRepository() repository.Repository {
	return &storeRepository{
		room:           make(map[model.RoomId]*model.Room),
		userIdToRoomId: make(map[model.UserId]model.RoomId),
	}
}
