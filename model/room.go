package model

type Room struct {
	Id       RoomId
	Capacity Capacity
	HostId   UserId
	Members  []User
	Game     Game
}

type RoomId string

func (rid RoomId) String() string {
	return string(rid)
}

type Capacity int

func (c Capacity) Int() int {
	return int(c)
}

func (r *Room) AllMembersAreReady() bool {
	for _, m := range r.Members {
		if _, ok := r.Game.Ready[m.Id]; !ok {
			return false
		}
	}

	return true
}
