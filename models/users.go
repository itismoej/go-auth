package models

import (
	"github.com/mjafari98/go-auth/pb"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID             uint64 `gorm:"primarykey"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
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

func (user *User) FromProtoBuf(pbUser *pb.User) {
	user.ID = pbUser.ID
	user.FirstName = pbUser.FirstName
	user.LastName = pbUser.LastName
	user.Email = pbUser.Email
	user.Username = pbUser.Username
	user.Gender = pbUser.Gender
}

func (user *User) ConvertToProtoBuf() *pb.User {
	return &pb.User{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Username:  user.Username,
		Gender:    user.Gender,
	}
}
