package media

import (
	"cmsapp/internal/db"
	"context"
	"os"
	"time"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
		"github.com/gofiber/fiber/v2"
)

var s3Client *s3.Client

func InitS3() error {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(os.Getenv("AWS_REGION")),
	)
	if err != nil {
		panic("failed to load AWS config: " + err.Error())
	}
	s3Client = s3.NewFromConfig(cfg)
	return nil
}

func GeneratePresignedURL(fileName, mimeType string) (string, error) {
	presigner := s3.NewPresignClient(s3Client)
	key := fmt.Sprintf("public/%s", fileName)

	req, err := presigner.PresignPutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(os.Getenv("AWS_BUCKET_NAME")),
		Key: aws.String(key),
		ContentType: aws.String(mimeType),
	}, s3.WithPresignExpires(15*time.Minute))

	if err != nil {
		return "", err
	}

	return req.URL, nil
}

func SaveMedia(fileName, url, mimeType string, size int64, userID uint) error {
	asset := MediaAsset{
		FileName: fileName,
		URL: url,
		MimeType: mimeType,
		Size: size,
		UserID: userID,
	}
	return db.DB.Create(&asset).Error
}

func GetMedia(c *fiber.Ctx) error {
	var assets []MediaAsset
	if err := db.DB.Find(&assets).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error" : "Failed to fetch media"})
	}
	return c.JSON(assets)
}

func DeleteMedia(c *fiber.Ctx) error {
    id := c.Params("id")

    var asset MediaAsset
    if err := db.DB.First(&asset, id).Error; err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Media not found"})
    }

    _, err := s3Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
        Bucket: aws.String(os.Getenv("AWS_BUCKET_NAME")),
        Key:    aws.String(asset.FileName),
    })
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "failed to delete file from S3"})
    }

    if err := db.DB.Delete(&asset).Error; err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "failed to delete record"})
    }

    return c.JSON(fiber.Map{"message": "media deleted"})
}