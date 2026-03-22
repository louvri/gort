package gin

import (
	"encoding/json"
	"fmt"
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

func logAuthError(msg string) {
	escaped, _ := json.Marshal(msg)
	fmt.Printf("{\"module\":\"jwt-auth\", \"error\":%s}\n", escaped)
}

func jwtKeyFunc(key string, symmetric bool) jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
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
			logAuthError(err.Error())
			return nil, err
		}
		return verifyKey, nil
	}
}
