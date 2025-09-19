package user

import (
	"github.com/gofiber/fiber/v2"
	"cmsapp/internal/auth"
	"cmsapp/internal/db"
	"golang.org/x/crypto/bcrypt"
)

//Struct Login
type LoginRequest struct {
	Email string `json"email"`
	Password string `json:"password"`
}

//Struct Register
type RegisterRequest struct {
	Username string `json:"username"`
	Email string `json:"email"`
	Password string `json:"password"`
}

// Method Register
func RegisterHandler (c *fiber.Ctx) error {
	var body RegisterRequest
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error" : "Invalid Body"})
	}

	if err := CreateUser(body.Username, body.Email, body.Password); err != nil {
		return c.Status(400).JSON(fiber.Map{"error" : err.Error()})
	}

	return c.JSON(fiber.Map{"message" : "User Created"})
}

// Method Login
func LoginHandler(c *fiber.Ctx) error {
	var body LoginRequest
	if err := c.BodyParser(&body); err != nil{
		return c.Status(400).JSON(fiber.Map{"error" : "Invalid body"})
	}

	var u User
	if err := db.DB.Where("email = ?", body.Email).First(&u).Error; err != nil {
		return c.Status(401).JSON(fiber.Map{"error" : "Invalid, Cek Lagi"})
	}

	if u.Status == "suspended" {
		return c.Status(403).JSON(fiber.Map{"error" : "Account Disuspend"})
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(body.Password)); err != nil{
		return c.Status(401).JSON(fiber.Map{"error" : "Invalid, Cek Lagi"})
	}

	token, err := auth.GenerateJWT(u.ID, u.Role, u.Status)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error" : "Token have been failed to generate"})
	}

	return c.JSON(fiber.Map{
		"token" : token,
	})
}

//Handler Protect
func MeHandler(c *fiber.Ctx) error {
	 claims := c.Locals("user").(*auth.Claims)

	return c.JSON(fiber.Map{
		"user_id" : claims.UserID,
		"role" : claims.Role,
		"status" : claims.Status,
	})
}