package utility

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/sha3"
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword), err
}

func VerifyPassword(password, expectedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(expectedPassword), []byte(password))
}

func GetRandomKey() string {
	temp := make([]byte, 1024)
	rand.Read(temp)
	return HashSha3(temp)
}

func HashSha3(data []byte) string {
	hash := sha3.New512()
	hash.Write(data)
	return fmt.Sprintf("%x", hash.Sum(nil))
}
