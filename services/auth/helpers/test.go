package helpers

import (
	"gorm.io/gorm"

	"github.com/accentdesign/grpc/services/auth/internal/migrate"
	"github.com/accentdesign/grpc/services/auth/internal/models"
)

type TestHelpers struct {
	DB *gorm.DB
}

func (h *TestHelpers) MigrateDatabase() error {
	migrator := migrate.Migrator{DB: h.DB}
	return migrator.MigrateDatabase()
}

func (h *TestHelpers) CleanDatabase() error {
	if err := h.DB.Where("1 = 1").Delete(models.User{}).Error; err != nil {
		return err
	}
	if err := h.DB.Where("1 = 1").Delete(models.UserType{}).Error; err != nil {
		return err
	}
	return nil
}

func (h *TestHelpers) CreateTestUserType() (*models.UserType, error) {
	var userType models.UserType
	if err := h.DB.Where(models.UserType{IsDefault: true}).FirstOrCreate(&userType, models.UserType{Name: "standard"}).Error; err != nil {
		return nil, err
	}
	return &userType, nil
}

func (h *TestHelpers) CreateTestUser() (*models.User, error) {
	userType, err := h.CreateTestUserType()
	if err != nil {
		return nil, err
	}
	user := models.User{
		Email:      "test@example.com",
		FirstName:  "John",
		LastName:   "Doe",
		UserTypeId: userType.ID,
		IsActive:   true,
		IsVerified: false,
	}
	if err := user.SetPassword("password"); err != nil {
		return nil, err
	}
	if err := h.DB.Create(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
