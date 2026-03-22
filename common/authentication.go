package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
)

func GetAuthorizationHeaderValue(r *http.Request) string {
	return r.Header.Get("Authorization")
}

func GetBearerToken(r *http.Request) string {
	token, found := strings.CutPrefix(GetAuthorizationHeaderValue(r), "Bearer ")
	if !found {
		return ""
	}
	return token
}

func logAuthError(msg string) {
	escaped, _ := json.Marshal(msg)
	fmt.Printf("{\"module\":\"jwt-auth\", \"error\":%s}\n", escaped)
}

func JWTKeyFunc(key string, symmetric bool) jwt.Keyfunc {
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

func GetMapClaimsFromJWT(key, bearerToken string, symmetric bool) (jwt.MapClaims, error) {
	token, err := jwt.Parse(bearerToken, JWTKeyFunc(key, symmetric))
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims, nil
	}
	return nil, errors.New("claim type is not map")
}

func GetMapClaimsFromJWTWithoutValidation(bearerToken string) jwt.MapClaims {
	parser := &jwt.Parser{}
	claims := jwt.MapClaims{}
	_, _, err := parser.ParseUnverified(bearerToken, claims)
	if err != nil || len(claims) == 0 {
		return nil
	}
	return claims
}
