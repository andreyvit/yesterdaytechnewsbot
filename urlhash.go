package main

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func CanonicalURL(url string) string {
	url = strings.ToLower(url)
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	return url
}

func HashOfURL(url string) string {
	h := sha256.Sum256([]byte(url))
	return hex.EncodeToString(h[:])
}
