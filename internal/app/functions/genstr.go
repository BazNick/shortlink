package functions

import "math/rand"

const randomStr = "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandSeq(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = randomStr[rand.Intn(len(randomStr))]
	}
	return string(b)
}
