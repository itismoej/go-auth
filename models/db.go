package models

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
)

func ConnectAndMigrate() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	_ = db.AutoMigrate(&User{}, &Role{})
	roles := []Role{{Name: "Admin"}, {Name: "Service"}, {Name: "User"}}
	for _, role := range roles {
		db.FirstOrCreate(&role)
	}

	var adminRole Role
	db.Take(&adminRole, "Name = ?", "Admin")

	adminUser := &User{
		FirstName: "Admin",
		Email:     "admin@ui.ac.ir",
		Role:      adminRole,
		Username:  os.Getenv("ADMIN_USER"),
		IsActive:  true,
		IsAdmin:   true,
	}
	adminUser.SetNewPassword(os.Getenv("ADMIN_PASS"))
	db.FirstOrCreate(&adminUser)

	return db
}
