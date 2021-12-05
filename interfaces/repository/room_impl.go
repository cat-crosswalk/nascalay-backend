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

	uid := random.UserId()
	room := model.Room{
		Id:       rid,
		Capacity: cr.Capacity,
		HostId:   uid,
		Members: []model.User{
			{
				Id:     uid,
				Name:   cr.Username,
				Avatar: cr.Avatar,
			},
		},
	}
	r.Room[rid] = &room

	return &room, nil
}
