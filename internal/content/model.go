package content

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"time"
	"cmsapp/internal/user"
)
type EntryRelation struct {
    ID            uint   `gorm:"primaryKey"`
    SourceEntryID uint   `gorm:"index"`
    FieldName     string `gorm:"index"`
    TargetEntryID uint   `gorm:"index"`
}
type ContentModel struct {
	gorm.Model
	Name string `json:"name"`
	Schema datatypes.JSON `json:"schema"`
}

type ContentEntry struct {
	gorm.Model
	ModelID uint `json:"model_id"`
	Data datatypes.JSON `json:"data"`
	Status string `json:"status" gorm:"default:draft"`
	Slug string `gorm:"uniqueIndex" json:"slug"`
	AuthorID  uint `json:"author_id"`
    Author user.User `gorm:"foreignKey:AuthorID"`
	createAt time.Time
	UpdateAt time.Time
}

type AuditLog struct {
	ID        uint           `gorm:"primaryKey"`
	EntryID   uint           `json:"entry_id"`
	UserID    uint           `json:"user_id"`
	Action    string         `json:"action"`
	Changes   datatypes.JSON `json:"changes"`
	CreatedAt time.Time      `json:"created_at"`
}
