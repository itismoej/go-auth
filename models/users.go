package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	RoleID         uint
	Role           Role
	FirstName      string `gorm:"size:255"`
	LastName       string `gorm:"size:255"`
	Email          string `gorm:"type:varchar(100);unique_index"`
	Username       string `gorm:"type:varchar(30);unique_index"`
	HashedPassword string `gorm:"type:varchar(255)"`
	Gender         string
	IsActive       bool
}

func (user *User) SetNewPassword(rawPassword string) {
	bcryptPassword, _ := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	user.HashedPassword = string(bcryptPassword)
}

func (user *User) PasswordIsCorrect(rawPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(rawPassword))
	return err == nil
}
