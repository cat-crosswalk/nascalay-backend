package handler

import (
	"github.com/21hack02win/nascalay-backend/model"
	"github.com/21hack02win/nascalay-backend/oapi"
)

func refillRoom(mr *model.Room, userId model.UserId) oapi.Room {
	var r oapi.Room
	r.Capacity = mr.Capacity.Int()
	r.HostId = mr.HostId.UUID()
	r.Members = make([]oapi.User, len(mr.Members))
	r.RoomId = mr.Id.String()
	r.UserId = userId.UUID()

	for i, v := range mr.Members {
		r.Members[i] = refillUser(&v)
	}

	return r
}

func refillUser(mu *model.User) oapi.User {
	var u oapi.User
	u.Avatar = mu.Avatar.Int()
	u.UserId = mu.Id.UUID()
	u.Username = mu.Name.String()

	return u
}
