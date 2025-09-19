package user

import(
	"gorm.io/gorm"
)

type User struct {
gorm.Model
Username string `gorm:"uniqeIndex;size:50" json:"username"`
Email string `gorm:"uniqeIndex;size:100" json"email"`
Password string `json:"-"`
Role string `gorm:"size:20;default:viewer" json"role"`
Status   string `gorm:"size:20;default:active" json:"status"`
}