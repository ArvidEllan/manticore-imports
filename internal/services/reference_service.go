package services

import (
	"fmt"
	"math/rand"
	"time"
)

const letters = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func GenerateReference(now time.Time) string {
	rand.Seed(time.Now().UnixNano())
	suffix := make([]byte, 5)
	for i := range suffix {
		suffix[i] = letters[rand.Intn(len(letters))]
	}
	return fmt.Sprintf("MANT-%s-%s", now.Format("20060102"), string(suffix))
}
