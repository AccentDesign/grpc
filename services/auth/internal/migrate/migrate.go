package migrate

import (
	"gorm.io/gorm"

	"github.com/accentdesign/grpc/services/auth/internal/models"
)

type Migrator struct {
	DB *gorm.DB
}

func (m *Migrator) SetupRemoveResetTokenTrigger() error {
	triggerFunction := `
CREATE OR REPLACE FUNCTION func_remove_auth_reset_tokens()
RETURNS TRIGGER AS $$
BEGIN
  DELETE FROM auth_reset_tokens
  WHERE user_id = NEW.id;
RETURN NEW;
END;
$$ LANGUAGE plpgsql;
`

	trigger := `
DO
$$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'remove_auth_reset_tokens') THEN
    CREATE TRIGGER remove_auth_reset_tokens
      AFTER UPDATE ON auth_users
      FOR EACH ROW
      WHEN (OLD.hashed_password != NEW.hashed_password)
      EXECUTE FUNCTION func_remove_auth_reset_tokens();
  END IF;
END;
$$
`

	if err := m.DB.Exec(triggerFunction).Error; err != nil {
		return err
	}

	if err := m.DB.Exec(trigger).Error; err != nil {
		return err
	}

	return nil
}

func (m *Migrator) SetupRemoveVerifyTokenTrigger() error {
	triggerFunction := `
CREATE OR REPLACE FUNCTION func_remove_auth_verify_tokens()
RETURNS TRIGGER AS $$
BEGIN
  DELETE FROM auth_verify_tokens
  WHERE user_id = NEW.id;
RETURN NEW;
END;
$$ LANGUAGE plpgsql;
`

	trigger := `
DO
$$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'remove_auth_verify_tokens') THEN
    CREATE TRIGGER remove_auth_verify_tokens
      AFTER UPDATE ON auth_users
      FOR EACH ROW
      WHEN (NEW.is_verified is true)
      EXECUTE FUNCTION func_remove_auth_verify_tokens();
  END IF;
END;
$$
`

	if err := m.DB.Exec(triggerFunction).Error; err != nil {
		return err
	}

	if err := m.DB.Exec(trigger).Error; err != nil {
		return err
	}

	return nil
}

func (m *Migrator) MigrateDatabase() error {
	if err := m.DB.AutoMigrate(
		&models.Scope{},
		&models.UserType{},
		&models.User{},
		&models.AccessToken{},
		&models.ResetToken{},
		&models.VerifyToken{},
	); err != nil {
		return err
	}

	if err := m.SetupRemoveResetTokenTrigger(); err != nil {
		return err
	}

	if err := m.SetupRemoveVerifyTokenTrigger(); err != nil {
		return err
	}

	return nil
}
