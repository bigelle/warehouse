package handlers

import "golang.org/x/crypto/bcrypt"

func HashPassword(p string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	return string(b), err
}

func IsCorrectPassword(password, hash string) bool {
	return nil == bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
