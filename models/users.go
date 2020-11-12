package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	RoleID         int
	Role           Role
	FirstName      string `gorm:"size:255"`
	LastName       string `gorm:"size:255"`
	Email          string `gorm:"type:varchar(100);unique_index"`
	Username       string `gorm:"type:varchar(30);unique_index"`
	HashedPassword []byte
	Gender         string
	IsActive       bool
}

func (user *User) SetNewPassword(rawPassword string) {
	bcryptPassword, _ := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	user.HashedPassword = bcryptPassword
}

func (user *User) CheckPassword(rawPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(rawPassword))
	return err == nil
}
