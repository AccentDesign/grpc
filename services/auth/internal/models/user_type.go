package models

import (
	"github.com/google/uuid"
)

type UserType struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name      string    `gorm:"type:varchar(120);unique;not null"`
	IsDefault bool      `gorm:"type:boolean;not null;default:false"`
	Scopes    []Scope   `gorm:"many2many:auth_user_type_scopes;"`
}

func (*UserType) TableName() string {
	return "auth_user_types"
}

func (u *UserType) ScopeNames() []string {
	names := make([]string, len(u.Scopes))
	for i, scope := range u.Scopes {
		names[i] = scope.Name
	}
	return names
}
