package random

import (
	"crypto/rand"
	"math/big"

	"github.com/21hack02win/nascalay-backend/model"
	"github.com/gofrs/uuid"
)

const (
	letters      = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	roomIdLength = 10
)

func RoomId() model.RoomId {
	ret := make([]byte, roomIdLength)
	for i := 0; i < roomIdLength; i++ {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		ret[i] = letters[num.Int64()]
	}
	return model.RoomId(ret)
}

func UserId() model.UserId {
	return model.UserId(uuid.Must(uuid.NewV4()))
}
