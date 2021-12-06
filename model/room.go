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
