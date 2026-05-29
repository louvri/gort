package common

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateAuthToken mints an HS256-signed JWT for sub, verifiable via
// JWTKeyFunc(jwtKey, true). The variadic data is stored under the "data" claim:
// zero args yield a nil claim (JSON null), and one or more args yield a []any,
// so a single value still arrives as a one-element slice (consumers must
// type-assert to []any and index). jwtLifetimeInMinute sets exp relative to
// iat; values <= 0 mint an already-expired token (intended for tests, not
// production callers).
func GenerateAuthToken(sub, jwtKey string, jwtLifetimeInMinute int, data ...any) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  sub,
		"exp":  now.Add(time.Duration(jwtLifetimeInMinute) * time.Minute).Unix(),
		"iat":  now.Unix(),
		"data": data,
	})

	tokenString, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		slog.Error("failed to sign token", "module", "jwt-auth", "error", err)
		return "", err
	}

	return tokenString, nil
}

// GenerateAuthTokenAsym mints an RS256-signed JWT for sub using the PEM-encoded
// RSA private key, verifiable via JWTKeyFunc(publicKeyPEM, false). The "data"
// claim and jwtLifetimeInMinute behave as documented on GenerateAuthToken.
func GenerateAuthTokenAsym(sub, jwtPrivKey string, jwtLifetimeInMinute int, data ...any) (string, error) {
	signKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(jwtPrivKey))
	if err != nil {
		slog.Error("failed to parse RSA private key", "module", "jwt-auth", "error", err)
		return "", err
	}

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub":  sub,
		"exp":  now.Add(time.Duration(jwtLifetimeInMinute) * time.Minute).Unix(),
		"iat":  now.Unix(),
		"data": data,
	})

	tokenString, err := token.SignedString(signKey)
	if err != nil {
		slog.Error("failed to sign token", "module", "jwt-auth", "error", err)
		return "", err
	}

	return tokenString, nil
}

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

func JWTKeyFunc(key string, symmetric bool) jwt.Keyfunc {
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
	parser := jwt.NewParser()
	claims := jwt.MapClaims{}
	_, _, err := parser.ParseUnverified(bearerToken, claims)
	if err != nil || len(claims) == 0 {
		return nil
	}
	return claims
}
