package functions

import (
	"crypto/rand"
	"encoding/hex"
)


func RandSeq(n int) string {
	b := make([]byte, n)
    _, err := rand.Read(b)
    if err != nil {
        panic(err)
    }

    return hex.EncodeToString(b)[:n]
}
