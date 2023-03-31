package db_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"

	"github.com/accentdesign/grpc/db"
)

func TestConnect(t *testing.T) {
	// Create a new SQL mock
	testDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	// Set expectations for the mock (gorm.Open will execute Ping internally)
	mock.ExpectPing()

	// Create a gorm postgres config with the mocked SQL database
	dialector := postgres.New(postgres.Config{Conn: testDB})
	config := gorm.Config{
		Dialector: dialector,
		Plugins: map[string]gorm.Plugin{
			"dbresolver": dbresolver.Register(dbresolver.Config{}),
		},
	}

	// Call Connect with the DSN and config containing the mocked SQL database
	_, err = db.Connect(dialector, config)
	require.NoError(t, err)

	// Ensure all expectations were met
	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}
