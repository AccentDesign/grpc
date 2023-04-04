package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"github.com/accentdesign/grpc/core/validator"
	"github.com/accentdesign/grpc/services/auth/internal/models"
	"github.com/accentdesign/grpc/services/auth/internal/repos"
	pb "github.com/accentdesign/grpc/services/auth/pkg/api/auth"
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
		return nil, status.Error(codes.InvalidArgument, "invalid credentials")
	}

	if !user.VerifyPassword(password) {
		return nil, status.Error(codes.InvalidArgument, "invalid credentials")
	}

	token, tokenErr := s.TokenRepo.CreateAccessToken(user.ID)
	if tokenErr != nil {
		return nil, status.Error(codes.Internal, tokenErr.Error())
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
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Empty{}, nil
}

// Register creates a new user account based on the provided registration details.
// It takes in a context and a RegisterRequest, and returns a UserResponse and an error.
func (s *AuthService) Register(_ context.Context, in *pb.RegisterRequest) (*pb.UserResponse, error) {
	user, err := s.UserRepo.CreateUser(
		strings.TrimSpace(strings.ToLower(in.GetEmail())),
		in.GetPassword(),
		strings.TrimSpace(in.GetFirstName()),
		strings.TrimSpace(in.GetLastName()),
	)

	if err != nil {
		var ve *models.UserValidateError
		switch {
		case errors.As(err, &ve):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, gorm.ErrDuplicatedKey):
			return nil, status.Error(codes.AlreadyExists, "a user with this email already exists")
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
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

	user, userErr := s.UserRepo.GetUserByResetToken(token)
	if userErr != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid token")
	}

	if err := user.SetPassword(password); err != nil {
		var ve *models.UserValidateError
		switch {
		case errors.As(err, &ve):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	if err := s.UserRepo.UpdateUser(user); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
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
		return nil, status.Error(codes.Internal, tokenErr.Error())
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
		return nil, status.Error(codes.InvalidArgument, "invalid token")
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
		return nil, status.Error(codes.Internal, err.Error())
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
		return nil, status.Error(codes.FailedPrecondition, "user is already verified")
	}

	token, tokenErr := s.TokenRepo.CreateVerifyToken(user.ID)
	if tokenErr != nil {
		return nil, status.Error(codes.Internal, tokenErr.Error())
	}

	return &pb.TokenWithEmail{
		Token:     token.Token,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}, nil
}
