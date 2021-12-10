package model

type Room struct {
	Id       RoomId
	Capacity Capacity
	HostId   UserId
	Members  []User
	Game     *Game
}

type RoomId string

func (rid RoomId) String() string {
	return string(rid)
}

type Capacity int

func (c Capacity) Int() int {
	return int(c)
}

func (r *Room) GameStatusIs(status GameStatus) bool {
	return r.Game.Status == status
}

func (r *Room) ResetGame() {
	r.Game = InitGame()
}
