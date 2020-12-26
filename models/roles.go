package models

import (
	"gorm.io/gorm"
	"time"
)

type Role struct {
	ID        uint64 `gorm:"primarykey"`
	Name      string `gorm:"size:20;unique"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
