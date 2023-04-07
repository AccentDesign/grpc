package migrate

import (
	"gorm.io/gorm"

	"github.com/accentdesign/grpc/services/auth/internal/models"
)

type Migrator struct {
	DB *gorm.DB
}

func (m *Migrator) MigrateDatabase() error {
	if err := m.migrate(m.DB); err != nil {
		return err
	}

	return nil
}

func (m *Migrator) migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&models.Scope{},
		&models.UserType{},
		&models.User{},
		&models.AccessToken{},
		&models.ResetToken{},
		&models.VerifyToken{},
	); err != nil {
		return err
	}

	return nil
}
