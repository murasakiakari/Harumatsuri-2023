package utility

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

var SecretKey = GetRandomKey()

func SignClaims(claims jwt.Claims) (tokenString string, err error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token.SignedString([]byte(SecretKey))
}

// This function return error when the token is invalid
func ValidateTokenWithKey(tokenString string, t jwt.Claims, key string) (claims jwt.Claims, err error) {
	token, err := jwt.ParseWithClaims(tokenString, t, func(t *jwt.Token) (interface{}, error) {
		return []byte(key), nil
	})

	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("token not valid")
	}
	return token.Claims, nil
}

func ValidateToken(tokenString string, t jwt.Claims) (claims jwt.Claims, err error) {
	return ValidateTokenWithKey(tokenString, t, SecretKey)
}
