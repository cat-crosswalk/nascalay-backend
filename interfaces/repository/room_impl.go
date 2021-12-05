package repository

import (
	"github.com/21hack02win/nascalay-backend/model"
	"github.com/21hack02win/nascalay-backend/usecases/repository"
	"github.com/21hack02win/nascalay-backend/util/random"
)

func (r *storeRepository) JoinRoom(jr *repository.JoinRoomArgs) (*model.Room, model.UserId, error) {
	room, ok := r.Room[jr.RoomId]
	if !ok {
		return nil, model.UserId{}, repository.ErrNotFound
	}

	uid := random.UserId()
	room.Members = append(room.Members, model.User{
		Id:     uid,
		Name:   jr.Username,
		Avatar: jr.Avatar,
	})

	return room, uid, nil
}

func (r *storeRepository) CreateRoom(cr *repository.CreateRoomArgs) (*model.Room, error) {
	rid := random.RoomId()
	if _, ok := r.Room[rid]; ok {
		return nil, repository.ErrAlreadyExists
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

func (r *storeRepository) GetRoom(rid string) (*model.Room, error) {
	room, ok := r.Room[model.RoomId(rid)]
	if !ok {
		return nil, repository.ErrNotFound
	}

	return room, nil
}
