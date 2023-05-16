package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"github.com/accentdesign/grpc/services/auth/internal/models"
	"github.com/accentdesign/grpc/services/auth/internal/repos"
	pb "github.com/accentdesign/grpc/services/auth/pkg/api/auth"
)

var (
	ErrEmailAlreadyExists  = status.Error(codes.AlreadyExists, "a user with this email already exists")
	ErrEmailInvalid        = status.Error(codes.InvalidArgument, "invalid email format")
	ErrInvalidCredentials  = status.Error(codes.InvalidArgument, "invalid credentials")
	ErrPasswordRequired    = status.Error(codes.InvalidArgument, "password is required")
	ErrTokenInvalid        = status.Error(codes.InvalidArgument, "invalid token")
	ErrTokenRequired       = status.Error(codes.InvalidArgument, "token is required")
	ErrUserAlreadyVerified = status.Error(codes.FailedPrecondition, "user is already verified")
	ErrUserNotFound        = status.Error(codes.NotFound, "user not found")
)

func ErrInternal(err error) error {
	return status.Error(codes.Internal, err.Error())
}

func ErrInvalidArgument(err error) error {
	return status.Error(codes.InvalidArgument, err.Error())
}

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
	email := strings.TrimSpace(strings.ToLower(in.GetEmail()))
	if !govalidator.IsEmail(email) {
		return nil, ErrEmailInvalid
	}

	password := in.GetPassword()
	if govalidator.IsNull(password) {
		return nil, ErrPasswordRequired
	}

	user, err := s.UserRepo.GetUserByEmail(email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !user.VerifyPassword(password) {
		return nil, ErrInvalidCredentials
	}

	token, err := s.TokenRepo.CreateAccessToken(user.ID)
	if err != nil {
		return nil, ErrInternal(err)
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
	token := in.GetToken()
	if govalidator.IsNull(token) {
		return nil, ErrTokenRequired
	}

	if err := s.TokenRepo.RevokeBearerToken(token); err != nil {
		return nil, ErrInternal(err)
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
			return nil, ErrInvalidArgument(err)
		case errors.Is(err, gorm.ErrDuplicatedKey):
			return nil, ErrEmailAlreadyExists
		default:
			return nil, ErrInternal(err)
		}
	}

	return s.userToResponse(user), nil
}

// ResetPassword resets a user's password, given the provided reset password details.
// It takes in a context and a ResetPasswordRequest, and returns an Empty response and an error.
func (s *AuthService) ResetPassword(_ context.Context, in *pb.ResetPasswordRequest) (*pb.Empty, error) {
	token := in.GetToken()
	if govalidator.IsNull(token) {
		return nil, ErrTokenRequired
	}

	password := in.GetPassword()

	user, err := s.UserRepo.GetUserByResetToken(token)
	if err != nil {
		return nil, ErrTokenInvalid
	}

	if err := user.SetPassword(password); err != nil {
		var ve *models.UserValidateError
		switch {
		case errors.As(err, &ve):
			return nil, ErrInvalidArgument(err)
		default:
			return nil, ErrInternal(err)
		}
	}

	if err := s.UserRepo.UpdateUser(user); err != nil {
		return nil, ErrInternal(err)
	}

	return &pb.Empty{}, nil
}

// ResetPasswordToken generates a reset password token for a user based on their email.
// It takes in a context and a ResetPasswordTokenRequest, and returns a TokenWithEmail and an error.
func (s *AuthService) ResetPasswordToken(_ context.Context, in *pb.ResetPasswordTokenRequest) (*pb.TokenWithEmail, error) {
	email := strings.TrimSpace(strings.ToLower(in.GetEmail()))
	if !govalidator.IsEmail(email) {
		return nil, ErrEmailInvalid
	}

	user, err := s.UserRepo.GetUserByEmail(email)
	if err != nil {
		return nil, ErrUserNotFound
	}

	token, err := s.TokenRepo.CreateResetToken(user.ID)
	if err != nil {
		return nil, ErrInternal(err)
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
	token := in.GetToken()
	if govalidator.IsNull(token) {
		return nil, ErrTokenRequired
	}

	user, err := s.UserRepo.GetUserByAccessToken(token)
	if err != nil {
		return nil, ErrTokenInvalid
	}

	return s.userToResponse(user), nil
}

// UpdateUser updates a user based on the provided bearer token.
// It takes in a context and a UpdateUserRequest, and returns a UserResponse and an error.
func (s *AuthService) UpdateUser(_ context.Context, in *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	token := in.GetToken()
	if govalidator.IsNull(token) {
		return nil, ErrTokenRequired
	}

	user, err := s.UserRepo.GetUserByAccessToken(token)
	if err != nil {
		return nil, ErrTokenInvalid
	}

	email := strings.TrimSpace(strings.ToLower(in.GetEmail()))
	firstName := strings.TrimSpace(in.GetFirstName())
	lastName := strings.TrimSpace(in.GetLastName())
	password := in.GetPassword()

	if email != "" {
		user.Email = email
	}
	if firstName != "" {
		user.FirstName = firstName
	}
	if lastName != "" {
		user.LastName = lastName
	}

	if err := user.Validate(); err != nil {
		return nil, ErrInvalidArgument(err)
	}

	if password != "" {
		if err := user.SetPassword(password); err != nil {
			var ve *models.UserValidateError
			switch {
			case errors.As(err, &ve):
				return nil, ErrInvalidArgument(err)
			default:
				return nil, ErrInternal(err)
			}
		}
	}

	if err := s.UserRepo.UpdateUser(user); err != nil {
		switch {
		case errors.Is(err, gorm.ErrDuplicatedKey):
			return nil, ErrEmailAlreadyExists
		default:
			return nil, ErrInternal(err)
		}
	}

	return s.userToResponse(user), nil
}

// VerifyUser verifies a user's account, given the provided token.
// It takes in a context and a Token, and returns a UserResponse and an error.
func (s *AuthService) VerifyUser(_ context.Context, in *pb.Token) (*pb.UserResponse, error) {
	token := in.GetToken()
	if govalidator.IsNull(token) {
		return nil, ErrTokenRequired
	}

	user, err := s.UserRepo.GetUserByVerifyToken(token)
	if err != nil {
		return nil, ErrTokenInvalid
	}

	user.IsVerified = true
	if err := s.UserRepo.UpdateUser(user); err != nil {
		return nil, ErrInternal(err)
	}

	return s.userToResponse(user), nil
}

// VerifyUserToken generates a user verification token based on their email.
// It takes in a context and a VerifyUserTokenRequest, and returns a TokenWithEmail and an error.
func (s *AuthService) VerifyUserToken(_ context.Context, in *pb.VerifyUserTokenRequest) (*pb.TokenWithEmail, error) {
	email := strings.TrimSpace(strings.ToLower(in.GetEmail()))
	if !govalidator.IsEmail(email) {
		return nil, ErrEmailInvalid
	}

	user, err := s.UserRepo.GetUserByEmail(email)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if user.IsVerified {
		return nil, ErrUserAlreadyVerified
	}

	token, err := s.TokenRepo.CreateVerifyToken(user.ID)
	if err != nil {
		return nil, ErrInternal(err)
	}

	return &pb.TokenWithEmail{
		Token:     token.Token,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}, nil
}
