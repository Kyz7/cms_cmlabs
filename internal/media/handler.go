package media

import (
	"github.com/gofiber/fiber/v2"
	"strconv"
	"fmt"
	"os"
)

func PresignHandler(c *fiber.Ctx) error {
	fileName := c.Query("filename")
	mimeType := c.Query("mime")
	sizeStr := c.Query("size")

	if fileName == "" || mimeType == "" {
		return c.Status(400).JSON(fiber.Map{"error" : "filename and mime required"})
	}

	size, _ := strconv.ParseInt(sizeStr, 10, 64)
	userID := c.Locals("user_id").(float64)

	url, err := GeneratePresignedURL(fileName, mimeType)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error" : err.Error()})
	}

	SaveMedia(fileName, fmt.Sprintf("https://%s.s3.amazonaws.com/public/%s",
	os.Getenv("AWS_BUCKET_NAME"), fileName), mimeType, size, uint(userID))

	return c.JSON(fiber.Map{
		"upload_url" : url,
		"file_url" : fmt.Sprintf("https://%s.s3.%s.amazonaws.com/public/%s",
		os.Getenv("AWS_BUCKET_NAME"),
		os.Getenv("AWS_REGION"),
		fileName),
	})
}