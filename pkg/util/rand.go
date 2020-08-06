package util

import (
	"math/rand"
	"time"
)

var (
	newRand = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func RandInt() int {
	return newRand.Int()
}
