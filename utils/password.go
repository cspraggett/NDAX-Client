package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func GetSignature(nonce, api, userId, secret string) string {

	h := hmac.New(sha256.New, []byte(secret))

	passKey := nonce + userId + api
	h.Write([]byte(passKey))

	sha := hex.EncodeToString(h.Sum(nil))

	fmt.Println("Result:", sha)
	return sha
}
