package testutils

import (
	"fmt"
	"log"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func SetupDockerDB() (string, *gorm.DB, func()) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "14",
		Env: []string{
			"POSTGRES_USER=test",
			"POSTGRES_PASSWORD=test",
			"POSTGRES_DATABASE=test",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	conn := fmt.Sprintf("postgresql://test:test@localhost:%s/test?sslmode=disable", resource.GetPort("5432/tcp"))
	dialector := postgres.Open(conn)
	config := &gorm.Config{}

	if err := pool.Retry(func() error {
		var err error
		_, err = gorm.Open(dialector, config)
		return err
	}); err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	db, err := gorm.Open(dialector, config)
	if err != nil {
		log.Fatalf("Could not connect to db: %s", err)
	}

	cleanup := func() {
		if err := pool.Purge(resource); err != nil {
			log.Fatalf("Could not purge resource: %s", err)
		}
	}

	return conn, db, cleanup
}
