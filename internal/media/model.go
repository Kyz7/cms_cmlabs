package media

import (
	"gorm.io/gorm"
)

type MediaAsset struct {
	gorm.Model
	FileName string `json:"file_name"`
	URL string `json:"url"`
	MimeType string `json:"mime_type"`
	Size int64 `json:"size`
	UserID uint `json:"user_id"`
}