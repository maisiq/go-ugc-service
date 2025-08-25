package cache

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
)

func HashArgs(args ...string) string {
	h := sha1.New()
	for _, arg := range args {
		h.Write([]byte(arg))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func BuildKey(parts ...string) string {
	return fmt.Sprintf("cache:%s", join(parts, ":"))
}

func join(parts []string, sep string) string {
	result := ""
	for i, p := range parts {
		if len(p) > 0 {
			if i > 0 {
				result += sep
			}
			result += p
		}
	}
	return result
}
