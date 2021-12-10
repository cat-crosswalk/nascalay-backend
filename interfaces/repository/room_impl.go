package repository

import (
	"time"

	"github.com/21hack02win/nascalay-backend/model"
	"github.com/21hack02win/nascalay-backend/usecases/repository"
	"github.com/21hack02win/nascalay-backend/util/random"
)

func (r *storeRepository) JoinRoom(jr *repository.JoinRoomArgs) (*model.Room, model.UserId, error) {
	room, ok := r.room[jr.RoomId]
	if !ok {
		return nil, model.UserId{}, repository.ErrNotFound
	}

	uid := random.UserId()
	r.userIdToRoomId[uid] = jr.RoomId

	room.Members = append(room.Members, model.User{
		Id:     uid,
		Name:   jr.Username,
		Avatar: jr.Avatar,
	})

	return room, uid, nil
}

func (r *storeRepository) CreateRoom(cr *repository.CreateRoomArgs) (*model.Room, error) {
	rid := random.RoomId()
	if _, ok := r.room[rid]; ok {
		return nil, repository.ErrAlreadyExists
	}

	uid := random.UserId()
	r.userIdToRoomId[uid] = rid

	timeLimit := model.TimeLimit(40) // Default time limit is 40 seconds
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
		Game: &model.Game{
			Status:    model.GameStatusRoom,
			Ready:     make(map[model.UserId]struct{}),
			Odais:     make([]*model.Odai, 0, 100),
			TimeLimit: timeLimit,
			Timeout:   0,
			Timer:     model.NewTimer(time.Second * time.Duration(timeLimit)),
			DrawCount: 0,
			ShowCount: 0,
			NextShowPhase: 0,
		},
	}

	// TODO: オプション処理

	r.room[rid] = &room

	return &room, nil
}

func (r *storeRepository) GetRoom(rid model.RoomId) (*model.Room, error) {
	room, ok := r.room[rid]
	if !ok {
		return nil, repository.ErrNotFound
	}

	return room, nil
}

func (r *storeRepository) GetRoomFromUserId(uid model.UserId) (*model.Room, error) {
	rid, ok := r.userIdToRoomId[uid]
	if !ok {
		return nil, repository.ErrNotFound
	}

	room, ok := r.room[rid]
	if !ok {
		return nil, repository.ErrNotFound
	}

	return room, nil
}
