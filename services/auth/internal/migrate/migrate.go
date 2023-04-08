package migrate

import (
	"fmt"
	"gorm.io/gorm"

	"github.com/accentdesign/grpc/services/auth/internal/models"
)

type Migrator struct {
	DB *gorm.DB
}

func (m *Migrator) MigrateDatabase() error {
	fmt.Println("Starting migrations")
	if err := m.migrate(m.DB); err != nil {
		return err
	}
	fmt.Println("Migrations complete")

	return nil
}

func (m *Migrator) MigrateDatabaseDryRun() error {
	fmt.Println("Dry Run: Starting migrations")
	dryRunDB := m.DB.Session(&gorm.Session{DryRun: true, Logger: m.DB.Logger})
	if err := m.migrate(dryRunDB); err != nil {
		return err
	}
	fmt.Println("Dry Run: Migrations complete")

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
