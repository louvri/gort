package echo

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// errEmptyHMACKey rejects verification with an empty HMAC key, which would
// accept tokens anyone can sign with that same empty key.
var errEmptyHMACKey = errors.New("empty HMAC key")

func getBearerToken(r *http.Request) string {
	token, found := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !found {
		return ""
	}
	return token
}

func jwtKeyFunc(key string, symmetric bool) jwt.Keyfunc {
	return func(token *jwt.Token) (any, error) {
		if symmetric {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			if key == "" {
				return nil, errEmptyHMACKey
			}
			return []byte(key), nil
		}
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		verifyKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(key))
		if err != nil {
			slog.Error("failed to parse RSA public key", "module", "jwt-auth", "error", err)
			return nil, err
		}
		return verifyKey, nil
	}
}

// matchesServerKey reports whether value equals key in constant time. An empty
// key never matches, so an unset key cannot authorize a missing header.
func matchesServerKey(value, key string) bool {
	return key != "" && subtle.ConstantTimeCompare([]byte(value), []byte(key)) == 1
}
