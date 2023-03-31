package models

import (
	"github.com/google/uuid"
)

type Scope struct {
	ID   uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name string    `gorm:"type:varchar(120);unique;not null"`
}

func (*Scope) TableName() string {
	return "auth_scopes"
}
