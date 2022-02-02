package application

import (
	"math/rand"
	"time"
)

func GetRandomMinutes() int {
	rand.Seed(time.Now().Unix())

	rangeLower := 1
	rangeUpper := 59
	randomNum := rangeLower + rand.Intn(rangeUpper-rangeLower+1)
	return randomNum
}
