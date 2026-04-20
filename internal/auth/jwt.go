package auth

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"

	// новые импорты
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

type Claims struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

var accessSecret = []byte(os.Getenv("JWT_ACCESS_SECRET"))

func GenerateAccessToken(userID int, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(accessSecret)
}
func ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return accessSecret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, err
	}

	return claims, nil
}
func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32) // 256 бит

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
