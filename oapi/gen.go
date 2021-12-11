//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest --config ./.server.yml ../docs/openapi.yml
//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest --config ./.types.yml ../docs/openapi.yml

package oapi

import "github.com/21hack02win/nascalay-backend/model"

// NOTE: UserIdInQueryをuuid.UUID型にするとBindしてくれないためstring型にしてここで詰め替えをする
func (u UserIdInQuery) Refill() (model.UserId, error) {
	return model.UserIdFromString(string(u))
}

func RefillUser(mu *model.User) User {
	return User{
		Avatar: Avatar{
			Color: mu.Avatar.Color.String(),
			Type:  mu.Avatar.Type.Int(),
		},
		Username: mu.Name.String(),
		UserId:   mu.Id.UUID(),
	}
}

func RefillUsers(mus []model.User) []User {
	us := make([]User, len(mus))
	for i, v := range mus {
		us[i] = RefillUser(&v)
	}

	return us
}

func RefillRoom(mr *model.Room, userId model.UserId) Room {
	var r Room
	r.Capacity = mr.Capacity.Int()
	r.HostId = mr.HostId.UUID()
	r.Members = make([]User, len(mr.Members))
	r.RoomId = mr.Id.String()
	r.UserId = userId.UUID()

	for i, v := range mr.Members {
		r.Members[i] = RefillUser(&v)
	}

	return r
}
