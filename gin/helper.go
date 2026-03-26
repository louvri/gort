package gin

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
)

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
