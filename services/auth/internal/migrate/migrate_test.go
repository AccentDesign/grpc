package migrate_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/accentdesign/grpc/services/auth/helpers"
	"github.com/accentdesign/grpc/services/auth/internal/migrate"
	"github.com/accentdesign/grpc/testutils"
)

type TestSuite struct {
	suite.Suite
	helpers *helpers.TestHelpers
	db      *gorm.DB
	cleanup func()
}

func (suite *TestSuite) SetupSuite() {
	_, suite.db, suite.cleanup = testutils.SetupDockerDB()
	suite.helpers = &helpers.TestHelpers{DB: suite.db}
}

func (suite *TestSuite) TearDownSuite() {
	suite.cleanup()
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (suite *TestSuite) TestMigrate_MigrateDatabase() {
	migrator := &migrate.Migrator{DB: suite.db}

	// test dry run
	err := migrator.MigrateDatabaseDryRun()
	suite.NoError(err)

	var count int64
	err = suite.db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name like 'auth_%'").Scan(&count).Error
	suite.NoError(err)
	suite.Equal(int64(0), count)

	// test real migration
	err = migrator.MigrateDatabase()
	suite.NoError(err)

	for _, table := range []string{
		"auth_scopes",
		"auth_user_types",
		"auth_user_type_scopes",
		"auth_users",
		"auth_access_tokens",
		"auth_reset_tokens",
		"auth_verify_tokens",
	} {
		var count int64
		err := suite.db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = ?", table).Scan(&count).Error
		suite.NoError(err)
		suite.Equal(int64(1), count)
	}
}
