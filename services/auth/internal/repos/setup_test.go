package repos_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/accentdesign/grpc/services/auth/helpers"
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
	err := suite.helpers.MigrateDatabase()
	suite.NoError(err)
}

func (suite *TestSuite) TearDownSuite() {
	suite.cleanup()
}

func (suite *TestSuite) Setup() func() {
	return func() {
		err := suite.helpers.CleanDatabase()
		suite.NoError(err)
	}
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
