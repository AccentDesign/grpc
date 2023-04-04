package repos

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/accentdesign/grpc/services/auth/internal/models"
)

type UserRepository struct {
	DB *gorm.DB
}

func (r *UserRepository) getDefaultUserType() (*models.UserType, error) {
	var userType models.UserType
	if err := r.DB.Where("is_default is true").First(&userType).Error; err != nil {
		return nil, fmt.Errorf("error fetching user type: %v", err)
	}

	return &userType, nil
}

func (r *UserRepository) getUserByToken(token string, tokenTable interface{}) (*models.User, error) {
	var user models.User
	now := time.Now()

	var joinTable string
	switch tokenTable.(type) {
	case *models.AccessToken:
		t := tokenTable.(*models.AccessToken)
		joinTable = t.TableName()
	case *models.ResetToken:
		t := tokenTable.(*models.ResetToken)
		joinTable = t.TableName()
	case *models.VerifyToken:
		t := tokenTable.(*models.VerifyToken)
		joinTable = t.TableName()
	default:
		return nil, fmt.Errorf("unsupported token type: %T", tokenTable)
	}

	result := r.DB.Joins(fmt.Sprintf("JOIN %s t ON %s.ID = t.user_id", joinTable, user.TableName())).
		Where("t.token = ? AND t.expires_at >= ?", token, now).
		Preload("UserType").Preload("UserType.Scopes").First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found for token: %s", token)
		}
		return nil, fmt.Errorf("error fetching user: %v", result.Error)
	}

	return &user, nil
}

func (r *UserRepository) CreateUser(email string, password string, firstName string, lastName string) (*models.User, error) {
	userType, userTypeErr := r.getDefaultUserType()
	if userTypeErr != nil {
		return nil, errors.New("no default user type exists")
	}

	user := models.User{
		Email:      email,
		FirstName:  firstName,
		LastName:   lastName,
		UserTypeId: userType.ID,
		CreatedAt:  time.Time{},
	}

	if err := user.Validate(); err != nil {
		return nil, err
	}

	if err := user.SetPassword(password); err != nil {
		return nil, err
	}

	if err := r.DB.Create(&user).Error; err != nil {
		return nil, err
	}

	var fetchUser models.User
	if err := r.DB.Preload("UserType").Preload("UserType.Scopes").First(&fetchUser, "id = ?", user.ID).Error; err != nil {
		return nil, err
	}

	return &fetchUser, nil
}

func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	result := r.DB.Where("email = ?", email).First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found: %s", email)
		}
		return nil, fmt.Errorf("error fetching user: %v", result.Error)
	}

	return &user, nil
}

func (r *UserRepository) GetUserByAccessToken(token string) (*models.User, error) {
	return r.getUserByToken(token, &models.AccessToken{})
}

func (r *UserRepository) GetUserByResetToken(token string) (*models.User, error) {
	return r.getUserByToken(token, &models.ResetToken{})
}

func (r *UserRepository) GetUserByVerifyToken(token string) (*models.User, error) {
	return r.getUserByToken(token, &models.VerifyToken{})
}

func (r *UserRepository) UpdateUser(user *models.User) error {
	if err := user.Validate(); err != nil {
		return err
	}

	if result := r.DB.Save(user); result.Error != nil {
		return result.Error
	}
	return nil
}
