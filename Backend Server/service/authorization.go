package service

import (
	"backendserver/database"
	"backendserver/utility"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	_CADS_SERVER  = "CADS Server"
	_AUTH_SUBJECT = "Authorization"
)

func CreateAccount(account *database.Account) (err error) {
	account.Password, err = utility.HashPassword(account.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err = account.Create(); err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}
	return nil
}

func Login(account *database.Account) (token string, err error) {
	enterPassword := account.Password
	err = account.Read()
	if err != nil {
		return "", fmt.Errorf("failed to get account: %w", err)
	}

	if err := utility.VerifyPassword(enterPassword, account.Password); err != nil {
		return "", fmt.Errorf("failed to verify password: %w", err)
	}

	currentTime := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer:    _CADS_SERVER,
		Subject:   _AUTH_SUBJECT,
		Audience:  jwt.ClaimStrings{account.Username},
		ExpiresAt: jwt.NewNumericDate(currentTime.Add(time.Hour * 10)),
		IssuedAt:  jwt.NewNumericDate(currentTime),
		ID:        account.Permission,
	}

	token, err = utility.SignClaims(claims)
	if err != nil {
		return "", fmt.Errorf("failed to create token: %w", err)
	}

	return token, nil
}

func ValidateToken(tokenString string) (claims *jwt.RegisteredClaims, err error) {
	claims_, err := utility.ValidateToken(tokenString, &jwt.RegisteredClaims{})
	if err != nil {
		return nil, err
	}

	claims, ok := claims_.(*jwt.RegisteredClaims)
	if !ok {
		return nil, fmt.Errorf("failed to cast interface")
	}

	if !(claims.Issuer == _CADS_SERVER && claims.Subject == _AUTH_SUBJECT) {
		return nil, fmt.Errorf("failed to validate token: token has invalid issuer or subject")
	}

	return claims, nil
}
