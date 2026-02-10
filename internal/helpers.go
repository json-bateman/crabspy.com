package internal

import (
	"crypto/rand"
	"math/big"
)

const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func GenerateRoomCode(length int) string {
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[int(n.Int64())]
	}
	return string(b)
}
