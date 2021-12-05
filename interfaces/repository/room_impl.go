package repository

import (
	"github.com/21hack02win/nascalay-backend/model"
	"github.com/21hack02win/nascalay-backend/usecases/repository"
	"github.com/21hack02win/nascalay-backend/util/random"
)

func (r *storeRepository) CreateRoom(cr *repository.CreateRoomArgs) (*model.Room, error) {
	rid := random.RoomId()
	if _, ok := r.Room[rid]; ok {
		return nil, errAlreadyExists
	}

	room := model.Room{
		Id:       rid,
		Capacity: cr.Capacity,
		User: model.User{
			Id:     random.UserId(),
			Name:   cr.Username,
			Avatar: cr.Avatar,
		},
	}
	r.Room[rid] = &room

	return &room, nil
}
