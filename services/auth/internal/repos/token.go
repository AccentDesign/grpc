package repos

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/accentdesign/grpc/services/auth/internal/models"
)

type TokenConfig struct {
	BearerDuration time.Duration
	ResetDuration  time.Duration
	VerifyDuration time.Duration
}

type TokenRepository struct {
	DB     *gorm.DB
	Config *TokenConfig
}

func (r *TokenRepository) createToken(token interface{}, userId uuid.UUID, size int, duration time.Duration) error {
	randomBytes := make([]byte, size)
	if _, err := rand.Read(randomBytes); err != nil {
		return err
	}

	tokenStr := base64.URLEncoding.EncodeToString(randomBytes)
	now := time.Now()

	switch t := token.(type) {
	case *models.AccessToken:
		t.Token = tokenStr
		t.UserId = userId
		t.CreatedAt = now
		t.ExpiresAt = now.Add(duration)
	case *models.ResetToken:
		t.Token = tokenStr
		t.UserId = userId
		t.CreatedAt = now
		t.ExpiresAt = now.Add(duration)
	case *models.VerifyToken:
		t.Token = tokenStr
		t.UserId = userId
		t.CreatedAt = now
		t.ExpiresAt = now.Add(duration)
	default:
		return fmt.Errorf("unsupported token type: %T", t)
	}

	if result := r.DB.Create(token); result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *TokenRepository) CreateAccessToken(userId uuid.UUID) (*models.AccessToken, error) {
	accessToken := &models.AccessToken{}
	if err := r.createToken(accessToken, userId, 64, r.Config.BearerDuration); err != nil {
		return nil, err
	}
	return accessToken, nil
}

func (r *TokenRepository) CreateResetToken(userId uuid.UUID) (*models.ResetToken, error) {
	resetToken := &models.ResetToken{}
	if err := r.createToken(resetToken, userId, 64, r.Config.ResetDuration); err != nil {
		return nil, err
	}
	return resetToken, nil
}

func (r *TokenRepository) CreateVerifyToken(userId uuid.UUID) (*models.VerifyToken, error) {
	verifyToken := &models.VerifyToken{}
	if err := r.createToken(verifyToken, userId, 64, r.Config.VerifyDuration); err != nil {
		return nil, err
	}
	return verifyToken, nil
}

func (r *TokenRepository) RevokeBearerToken(token string) error {
	if result := r.DB.Where("token = ?", token).Delete(&models.AccessToken{}); result.Error != nil {
		return result.Error
	}
	return nil
}
