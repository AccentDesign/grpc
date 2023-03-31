package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/accentdesign/grpc/db"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/accentdesign/grpc/core/validator"
	"github.com/accentdesign/grpc/services/auth/internal/migrate"
	"github.com/accentdesign/grpc/services/auth/internal/models"
	"github.com/accentdesign/grpc/services/auth/internal/repos"
	pb "github.com/accentdesign/grpc/services/auth/pkg/api/auth"
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

type AuthService struct {
	pb.UnimplementedAuthenticationServer
	UserRepo  *repos.UserRepository
	TokenRepo *repos.TokenRepository
}

func (s *AuthService) userToResponse(user *models.User) *pb.UserResponse {
	return &pb.UserResponse{
		Id:        user.ID.String(),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		UserType: &pb.UserType{
			Name:   user.UserType.Name,
			Scopes: user.UserType.ScopeNames(),
		},
		IsActive:   user.IsActive,
		IsVerified: user.IsVerified,
	}
}

// BearerToken generates a bearer token for a user, based on their provided credentials.
// It takes in a context and a BearerTokenRequest, and returns a BearerTokenResponse and an error.
func (s *AuthService) BearerToken(_ context.Context, in *pb.BearerTokenRequest) (*pb.BearerTokenResponse, error) {
	v := validator.New()

	email := strings.TrimSpace(strings.ToLower(in.GetEmail()))
	if v.IsEmpty(email) {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if !v.Matches(email, validator.EmailRX) {
		return nil, status.Error(codes.InvalidArgument, "invalid email format")
	}

	password := in.GetPassword()
	if v.IsEmpty(password) {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	user, userErr := s.UserRepo.GetUserByEmail(email)
	if userErr != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	if !user.VerifyPassword(password) {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	token, tokenErr := s.TokenRepo.CreateAccessToken(user.ID)
	if tokenErr != nil {
		return nil, status.Error(codes.Unknown, "could not create access token")
	}

	seconds := float64(s.TokenRepo.Config.BearerDuration) / float64(time.Second)

	return &pb.BearerTokenResponse{
		AccessToken: token.Token,
		TokenType:   "bearer",
		Expiry:      int32(seconds),
	}, nil
}

// RevokeBearerToken revokes a user's bearer token.
// It takes in a context and a Token, and returns an Empty response and an error.
func (s *AuthService) RevokeBearerToken(_ context.Context, in *pb.Token) (*pb.Empty, error) {
	v := validator.New()

	token := in.GetToken()
	if v.IsEmpty(token) {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	if err := s.TokenRepo.RevokeBearerToken(token); err != nil {
		return nil, status.Error(codes.Internal, "failed to revoke bearer token")
	}

	return &pb.Empty{}, nil
}

// Register creates a new user account based on the provided registration details.
// It takes in a context and a RegisterRequest, and returns a UserResponse and an error.
func (s *AuthService) Register(_ context.Context, in *pb.RegisterRequest) (*pb.UserResponse, error) {
	v := validator.New()

	email := strings.TrimSpace(strings.ToLower(in.GetEmail()))
	if v.IsEmpty(email) {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if !v.Matches(email, validator.EmailRX) {
		return nil, status.Error(codes.InvalidArgument, "invalid email format")
	}

	password := in.GetPassword()
	if v.IsEmpty(password) {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}
	if !v.IsStringLength(password, 6, 72) {
		return nil, status.Error(codes.InvalidArgument, "password must be between 6 and 72 bytes in length")
	}

	firstName := strings.TrimSpace(in.GetFirstName())
	if v.IsEmpty(firstName) {
		return nil, status.Error(codes.InvalidArgument, "first_name is required")
	}

	lastName := strings.TrimSpace(in.GetLastName())
	if v.IsEmpty(lastName) {
		return nil, status.Error(codes.InvalidArgument, "last_name is required")
	}

	user, userErr := s.UserRepo.CreateUser(email, password, firstName, lastName)
	if userErr != nil {
		if errors.Is(userErr, gorm.ErrDuplicatedKey) {
			return nil, status.Error(codes.AlreadyExists, "a user with this email already exists")
		}
		return nil, status.Error(codes.Internal, userErr.Error())
	}

	return s.userToResponse(user), nil
}

// ResetPassword resets a user's password, given the provided reset password details.
// It takes in a context and a ResetPasswordRequest, and returns an Empty response and an error.
func (s *AuthService) ResetPassword(_ context.Context, in *pb.ResetPasswordRequest) (*pb.Empty, error) {
	v := validator.New()

	token := in.GetToken()
	if v.IsEmpty(token) {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	password := in.GetPassword()
	if v.IsEmpty(password) {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}
	if !v.IsStringLength(password, 6, 72) {
		return nil, status.Error(codes.InvalidArgument, "password must be between 6 and 72 bytes in length")
	}

	user, userErr := s.UserRepo.GetUserByResetToken(token)
	if userErr != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid token")
	}

	if err := user.SetPassword(password); err != nil {
		return nil, status.Error(codes.Internal, "error setting password")
	}

	if err := s.UserRepo.UpdateUser(user); err != nil {
		return nil, status.Error(codes.Internal, "failed to update user")
	}

	return &pb.Empty{}, nil
}

// ResetPasswordToken generates a reset password token for a user based on their email.
// It takes in a context and a ResetPasswordTokenRequest, and returns a TokenWithEmail and an error.
func (s *AuthService) ResetPasswordToken(_ context.Context, in *pb.ResetPasswordTokenRequest) (*pb.TokenWithEmail, error) {
	v := validator.New()

	email := strings.TrimSpace(strings.ToLower(in.GetEmail()))
	if v.IsEmpty(email) {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if !v.Matches(email, validator.EmailRX) {
		return nil, status.Error(codes.InvalidArgument, "invalid email format")
	}

	user, userErr := s.UserRepo.GetUserByEmail(email)
	if userErr != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	token, tokenErr := s.TokenRepo.CreateResetToken(user.ID)
	if tokenErr != nil {
		return nil, status.Error(codes.Unknown, "could not create reset token")
	}

	return &pb.TokenWithEmail{
		Token:     token.Token,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}, nil
}

// User retrieves user details based on the provided token.
// It takes in a context and a Token, and returns a UserResponse and an error.
func (s *AuthService) User(_ context.Context, in *pb.Token) (*pb.UserResponse, error) {
	v := validator.New()

	token := in.GetToken()
	if v.IsEmpty(token) {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	user, userErr := s.UserRepo.GetUserByAccessToken(token)
	if userErr != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	return s.userToResponse(user), nil
}

// VerifyUser verifies a user's account, given the provided token.
// It takes in a context and a Token, and returns a UserResponse and an error.
func (s *AuthService) VerifyUser(_ context.Context, in *pb.Token) (*pb.UserResponse, error) {
	v := validator.New()

	token := in.GetToken()
	if v.IsEmpty(token) {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	user, userErr := s.UserRepo.GetUserByVerifyToken(token)
	if userErr != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid token")
	}

	user.IsVerified = true
	if err := s.UserRepo.UpdateUser(user); err != nil {
		return nil, status.Error(codes.Internal, "failed to update user")
	}

	return s.userToResponse(user), nil
}

// VerifyUserToken generates a user verification token based on their email.
// It takes in a context and a VerifyUserTokenRequest, and returns a TokenWithEmail and an error.
func (s *AuthService) VerifyUserToken(_ context.Context, in *pb.VerifyUserTokenRequest) (*pb.TokenWithEmail, error) {
	v := validator.New()

	email := strings.TrimSpace(strings.ToLower(in.GetEmail()))
	if v.IsEmpty(email) {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if !v.Matches(email, validator.EmailRX) {
		return nil, status.Error(codes.InvalidArgument, "invalid email format")
	}

	user, userErr := s.UserRepo.GetUserByEmail(email)
	if userErr != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	if user.IsVerified {
		return nil, status.Error(codes.AlreadyExists, "user is already verified")
	}

	token, tokenErr := s.TokenRepo.CreateVerifyToken(user.ID)
	if tokenErr != nil {
		return nil, status.Error(codes.Unknown, "could not create verify token")
	}

	return &pb.TokenWithEmail{
		Token:     token.Token,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}, nil
}

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
	database, dbErr := db.Connect(gormDialector, gormConfig)
	if dbErr != nil {
		log.Fatalf("failed to connect to the database: %v", dbErr)
	}

	// migrate tables
	migrator := &migrate.Migrator{DB: database}
	if err := migrator.MigrateDatabase(); err != nil {
		log.Fatalf("error migrating database: %v", err)
	}

	// create the auth service
	authService := &AuthService{
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
	listen, listenErr := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if listenErr != nil {
		log.Fatalf("failed to listen: %v", listenErr)
	}

	// serve
	log.Printf("server listening at %v", listen.Addr())
	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
