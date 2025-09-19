package auth

import (
	"github.com/gofiber/fiber/v2"
)

func RoleRequired(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := c.Locals("user").(*Claims)
		for _, role := range roles {
			if claims.Role == role {
				return c.Next()
			}
		}
		return c.Status(403).JSON(fiber.Map{"error" : "Forbidden"})
	}
}