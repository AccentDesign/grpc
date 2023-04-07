package models

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/accentdesign/grpc/core/validator"
)

type UserValidateError struct {
	Message string
}

func (e *UserValidateError) Error() string {
	return e.Message
}

type User struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email          string    `gorm:"type:varchar(320);unique;not null"`
	HashedPassword string    `gorm:"type:varchar(1024);not null"`
	FirstName      string    `gorm:"type:varchar(120);not null"`
	LastName       string    `gorm:"type:varchar(120);not null"`
	UserTypeId     uuid.UUID `gorm:"not null"`
	UserType       UserType
	IsActive       bool `gorm:"type:boolean;not null;default:true"`
	IsVerified     bool `gorm:"type:boolean;not null;default:false"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (*User) TableName() string {
	return "auth_users"
}

func (u *User) BeforeUpdate(tx *gorm.DB) (err error) {
	var oldUser User
	if err := tx.Unscoped().Where("id = ?", u.ID).First(&oldUser).Error; err != nil {
		return err
	}

	tx.Statement.Context = context.WithValue(tx.Statement.Context, "oldUser", &oldUser)

	return nil
}

func (u *User) AfterUpdate(tx *gorm.DB) (err error) {
	oldUser, ok := tx.Statement.Context.Value("oldUser").(*User)
	if !ok {
		return fmt.Errorf("could not get old user from context")
	}

	if oldUser.HashedPassword != u.HashedPassword {
		tx.Where("user_id = ?", u.ID).Delete(&ResetToken{})
	}

	if u.IsVerified {
		tx.Where("user_id = ?", u.ID).Delete(&VerifyToken{})
	}

	return nil
}

func (u *User) SetPassword(password string) error {
	v := validator.New()

	if v.IsEmpty(password) {
		return &UserValidateError{"password is required"}
	}
	if !v.IsStringLength(password, 6, 72) {
		return &UserValidateError{"password must be between 6 and 72 characters in length"}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.HashedPassword = string(hashedPassword)

	return nil
}

func (u *User) Validate() error {
	v := validator.New()

	if v.IsEmpty(u.Email) {
		return &UserValidateError{"email is required"}
	}
	if !v.Matches(u.Email, validator.EmailRX) {
		return &UserValidateError{"invalid email format"}
	}
	if v.IsEmpty(u.FirstName) {
		return &UserValidateError{"first_name is required"}
	}
	if v.IsEmpty(u.LastName) {
		return &UserValidateError{"last_name is required"}
	}

	return nil
}

func (u *User) VerifyPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.HashedPassword), []byte(password))
	return err == nil
}
