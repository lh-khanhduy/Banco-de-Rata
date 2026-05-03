package utils

import (
	"math/rand"
	"strings"
)

const (
	charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func randBool() bool {
	return rand.Intn(2) == 0
}

func randStringFromSet(set ...string) string {
	n := len(set)
	if n == 0 {
		return ""
	}
	return set[rand.Intn(n)]
}

func randInt(min, max int) int {
	return min + rand.Intn(max-min+1)
}

func randFloat64(min float64, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

// RandomString generates a random string of length n.
func randomString(n int) string {
	var sb strings.Builder
	sb.Grow(n)
	for i := 0; i < n; i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}

	return sb.String()
}

func RandomOwner() string {
	return randomString(8)
}

func RandomMoney() int64 {
	return int64(randInt(0, 1000))
}

func RandomCurrency() string {
	currencies := []string{USD, EUR, CAD}
	n := len(currencies)
	return currencies[rand.Intn(n)]
}
