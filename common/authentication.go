package common

import (
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
