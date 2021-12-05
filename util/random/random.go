package random

import (
	"crypto/rand"
	"encoding/base64"
	"math/big"

	"github.com/gofrs/uuid"
)

const (
	roomIdLength = 10
	letterlen    = 26*2 + 10 // A-Z, a-z, 0-9
)

func RoomId() string {
	runes := make([]byte, roomIdLength)

	for i := 0; i < roomIdLength; i++ {
		num, _ := rand.Int(rand.Reader, big.NewInt(letterlen))
		runes[i] = byte(num.Int64())
	}

	return base64.RawStdEncoding.EncodeToString(runes)
}

func UserId() uuid.UUID {
	return uuid.Must(uuid.NewV4())
}
