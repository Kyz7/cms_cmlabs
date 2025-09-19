package user

import(
	"cmsapp/internal/db"
	"golang.org/x/crypto/bcrypt"
)

func CreateUser(username, email, password string) error {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), 12)

	u:= User{
		Username : username,
		Email : email,
		Password : string(hashed),
		Role : "viewer",
		Status : "active",
	}

	return db.DB.Create(&u).Error
}