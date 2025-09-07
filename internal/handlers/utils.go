package handlers

import (
	"time"

	"github.com/bigelle/warehouse/internal/schemas"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(p string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	return string(b), err
}

func IsCorrectPassword(password, hash string) bool {
	return nil == bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func GenerateAccessJWT(id string, role string, secret []byte, expires time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub":  id,
		"role": role,
		"exp":  time.Now().Add(expires).Unix(),
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(secret)
}

func GenerateRefreshJWT(id string, secret []byte, expires time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub": id,
		"exp": time.Now().Add(expires).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(secret)
}

func PgTypeText(txt string) pgtype.Text {
	return pgtype.Text{
		String: txt,
		Valid:  true,
	}
}

func UUIDFromString(str string) (pgtype.UUID, error) {
	u, err := uuid.Parse(str)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return pgtype.UUID{
		Bytes: u,
		Valid: true,
	}, nil
}

func IsAppropriateRole(v any, expected schemas.Role) bool {
	got, ok := v.(schemas.Role)
	return ok && expected <= got
}
