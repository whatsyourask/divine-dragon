package util

import (
	"math/rand"
	"time"
)

func RandInt() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(10-4) + 4
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
