package auth

import (
    "github.com/golang-jwt/jwt/v5"
	"time"
    "os"
)

type Claims struct {
    UserID uint   `json:"user_id"`
    Role   string `json:"role"`
    Status string  `json:"status"`
    jwt.RegisteredClaims
}

func GenerateJWT(userID uint, role string,status string) (string, error) {
    claims := Claims{
        UserID: userID,
        Role:   role,
        Status: status,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
        },
    }

    jwtSecret := []byte(os.Getenv("JWT_SECRET"))
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}
