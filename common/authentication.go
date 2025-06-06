package common

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	"net/http"
	"strings"
)

func GetAuthorizationHeaderValue(r *http.Request) string {
	header := r.Header["Authorization"]
	if len(header) == 0 {
		return ""
	}

	return header[0]
}

func GetBearerToken(r *http.Request) string {
	authorization := GetAuthorizationHeaderValue(r)
	tokens := strings.Split(authorization, " ")
	if len(tokens) > 1 {
		return tokens[1]
	}

	return ""
}

func GetMapClaimsFromJWT(key, bearerToken string, symmetric bool) (result jwt.MapClaims, err error) {
	token, err := jwt.Parse(bearerToken, func(token *jwt.Token) (interface{}, error) {
		if symmetric {
			return []byte(key), nil
		} else {
			verifyKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(key))
			if err != nil {
				fmt.Printf("{\"module\":\"jwt-auth\", \"error\":\"%s\"}\n", err.Error())
				return nil, err
			}

			return verifyKey, nil
		}
	})
	if err != nil {
		return result, err
	}
	if !token.Valid {
		return result, errors.New("invalid token")
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims, nil
	}
	return result, errors.New("claim type is not map")
}

func GetMapClaimsFromJWTWithoutValidation(bearerToken string) jwt.MapClaims {
	token, _ := jwt.Parse(bearerToken, func(token *jwt.Token) (interface{}, error) {
		return nil, nil
	})
	if token != nil && token.Claims != nil {
		if claims, ok := token.Claims.(jwt.MapClaims); ok && len(claims) > 0 {
			return claims
		}
	}
	return nil
}
