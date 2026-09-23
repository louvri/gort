// Package common provides framework-agnostic helpers for JWT minting and
// verification, bearer token extraction, and timezone parsing shared by the
// gort echo and gin middleware.
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

// errEmptyHMACKey is returned instead of signing or verifying with an empty
// HMAC key, which would let anyone forge tokens.
var errEmptyHMACKey = errors.New("empty HMAC key")

// GenerateAuthToken mints an HS256-signed JWT for sub, verifiable via
// JWTKeyFunc(jwtKey, true). The variadic data is stored under the "data" claim:
// zero args yield a nil claim (JSON null), and one or more args yield a []any,
// so a single value still arrives as a one-element slice (consumers must
// type-assert to []any and index). jwtLifetimeInMinute sets exp relative to
// iat; values <= 0 mint an already-expired token (intended for tests, not
// production callers). An empty jwtKey is rejected.
func GenerateAuthToken(sub, jwtKey string, jwtLifetimeInMinute int, data ...any) (string, error) {
	if jwtKey == "" {
		return "", errEmptyHMACKey
	}

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

// GetAuthorizationHeaderValue returns the raw Authorization header of r.
func GetAuthorizationHeaderValue(r *http.Request) string {
	return r.Header.Get("Authorization")
}

// GetBearerToken returns the token from a "Bearer <token>" Authorization
// header, or "" when the header is missing or uses another scheme.
func GetBearerToken(r *http.Request) string {
	token, found := strings.CutPrefix(GetAuthorizationHeaderValue(r), "Bearer ")
	if !found {
		return ""
	}
	return token
}

// JWTKeyFunc returns a jwt.Keyfunc that pins the signing algorithm family to
// the key type, preventing algorithm-confusion attacks: symmetric expects
// HMAC and uses key as the shared secret; otherwise it expects RSA and parses
// key as a PEM-encoded public key. An empty HMAC key is rejected, since HMAC
// would otherwise verify tokens anyone can sign with that same empty key.
func JWTKeyFunc(key string, symmetric bool) jwt.Keyfunc {
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

// GetMapClaimsFromJWT verifies bearerToken's signature using
// JWTKeyFunc(key, symmetric), rejects it if its exp or nbf claim (when
// present) is out of range, and returns its claims.
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

// GetMapClaimsFromJWTWithoutValidation decodes bearerToken's claims WITHOUT
// verifying its signature or expiry, so the result must never be used for
// authentication or authorization. It returns nil if the token is malformed
// or has no claims.
func GetMapClaimsFromJWTWithoutValidation(bearerToken string) jwt.MapClaims {
	parser := jwt.NewParser()
	claims := jwt.MapClaims{}
	_, _, err := parser.ParseUnverified(bearerToken, claims)
	if err != nil || len(claims) == 0 {
		return nil
	}
	return claims
}
