package random

import (
	"crypto/rand"
	"math/big"

	"github.com/gofrs/uuid"
)

const (
	letters      = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	roomIdLength = 10
)

func RoomId() string {
	ret := make([]byte, roomIdLength)
	for i := 0; i < roomIdLength; i++ {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		ret[i] = letters[num.Int64()]
	}
	return string(ret)
}

func UserId() uuid.UUID {
	return uuid.Must(uuid.NewV4())
}
