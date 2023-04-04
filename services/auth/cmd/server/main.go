package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/accentdesign/grpc/services/auth/internal/migrate"
	"github.com/accentdesign/grpc/services/auth/internal/repos"
	pb "github.com/accentdesign/grpc/services/auth/pkg/api/auth"
	"github.com/accentdesign/grpc/services/auth/service"
)

var (
	helpFlag         = flag.Bool("help", false, "Display help information")
	enableReflection = flag.Bool("reflection", false, "Enable reflection")
	port             = flag.Int("port", 50051, "The server port")
	bearerDuration   = flag.Duration("bearer-duration", 3600*time.Second, "Bearer token duration")
	resetDuration    = flag.Duration("reset-duration", 3600*time.Second, "Reset token duration")
	verifyDuration   = flag.Duration("verify-duration", 3600*time.Second, "Verify token duration")
	gormDialector    = postgres.New(postgres.Config{DSN: os.Getenv("DB_DNS")})
	gormConfig       = gorm.Config{}
)

func displayHelp() {
	flag.PrintDefaults()
	fmt.Println("Environment variables:")
	fmt.Println("  DB_DNS - Database dns e.g. postgresql://user:pass@host:5432/db?sslmode=disable")
}

func main() {
	// get runtime flags
	flag.Parse()

	// display help
	if *helpFlag {
		displayHelp()
		os.Exit(0)
	}

	// connect to the database
	database, err := gorm.Open(gormDialector, &gormConfig)
	if err != nil {
		log.Fatalf("failed to connect to the database: %v", err)
	}

	// migrate tables
	migrator := &migrate.Migrator{DB: database}
	if err := migrator.MigrateDatabase(); err != nil {
		log.Fatalf("error migrating database: %v", err)
	}

	// create the auth service
	authService := &service.AuthService{
		UserRepo: &repos.UserRepository{DB: database},
		TokenRepo: &repos.TokenRepository{
			DB: database,
			Config: &repos.TokenConfig{
				BearerDuration: *bearerDuration,
				ResetDuration:  *resetDuration,
				VerifyDuration: *verifyDuration,
			},
		},
	}

	// create the server
	grpcServer := grpc.NewServer()

	// register the services
	pb.RegisterAuthenticationServer(grpcServer, authService)

	// enable reflection
	if *enableReflection {
		reflection.Register(grpcServer)
		log.Print("reflection enabled")
	}

	// configure the listen address
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// serve
	log.Printf("server listening at %v", listen.Addr())
	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
