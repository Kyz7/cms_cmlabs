package auth

import(
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"strings"
	"os"
)

func AuthRequired(c *fiber.Ctx) error {
    var jwtSecret = []byte(os.Getenv("JWT_SECRET"))
    authHeader := c.Get("Authorization")
    if authHeader == "" {
        return c.Status(401).JSON(fiber.Map{"error": "Missing token"})
    }

    tokenString := strings.TrimPrefix(authHeader, "Bearer ")
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return jwtSecret, nil
    })

    if err != nil || !token.Valid {
        return c.Status(401).JSON(fiber.Map{"error": "Invalid token"})
    }

    claims := token.Claims.(*Claims)
    c.Locals("user", claims)

    return c.Next()
}