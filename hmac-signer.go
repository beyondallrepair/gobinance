package gobinance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// HMACSigner provides a method for signing data using a secret key.
type HMACSigner struct {
	Secret string
}

// Sign generates an HMAC SHA256 signature of the `input` parameter given the `Secret` key in the HMACSigner
func (h *HMACSigner) Sign(input string) string {
	hash := hmac.New(sha256.New, []byte(h.Secret))
	hash.Write([]byte(input))
	return hex.EncodeToString(hash.Sum(nil))
}
