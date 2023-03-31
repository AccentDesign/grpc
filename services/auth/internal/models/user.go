package models

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email          string    `gorm:"type:varchar(320);unique;not null"`
	HashedPassword string    `gorm:"type:varchar(1024);not null"`
	FirstName      string    `gorm:"type:varchar(120);not null"`
	LastName       string    `gorm:"type:varchar(120);not null"`
	UserTypeId     uuid.UUID
	UserType       UserType
	IsActive       bool `gorm:"type:boolean;not null;default:true"`
	IsVerified     bool `gorm:"type:boolean;not null;default:false"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (*User) TableName() string {
	return "auth_users"
}

func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.HashedPassword = string(hashedPassword)
	return nil
}

func (u *User) VerifyPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.HashedPassword), []byte(password))
	return err == nil
}
