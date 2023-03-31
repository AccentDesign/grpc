package models

import (
	"time"

	"github.com/google/uuid"
)

type ResetToken struct {
	Token     string    `gorm:"type:varchar(1024);primary_key"`
	UserId    uuid.UUID `gorm:"not null;index"`
	User      User      `gorm:"constraint:OnDelete:CASCADE"`
	CreatedAt time.Time
	ExpiresAt time.Time
}

func (*ResetToken) TableName() string {
	return "auth_reset_tokens"
}
