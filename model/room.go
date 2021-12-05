package model

type Room struct {
	Id       RoomId
	Capacity int
	HostId   UserId
	Members  []User
}

type RoomId string

func (rid RoomId) String() string {
	return string(rid)
}
