package qcow2ova

import (
	"crypto/rand"
	"encoding/base64"
)

// Generates the password of length n
func GeneratePassword(n int) (b64Password string, err error) {
	b := make([]byte, n)
	_, err = rand.Read(b)
	if err != nil {
		return
	}
	b64Password = base64.URLEncoding.EncodeToString(b)
	return
}
