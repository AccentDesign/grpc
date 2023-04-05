package service_test

import (
	"context"
	"gorm.io/gorm"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/accentdesign/grpc/services/auth/helpers"
	"github.com/accentdesign/grpc/services/auth/internal/models"
	"github.com/accentdesign/grpc/services/auth/internal/repos"
	pb "github.com/accentdesign/grpc/services/auth/pkg/api/auth"
	"github.com/accentdesign/grpc/services/auth/service"
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

func (suite *TestSuite) TestAuthService_BearerToken() {
	teardown := suite.Setup()
	defer teardown()

	user, err := suite.helpers.CreateTestUser()
	suite.NoError(err)

	// create the repos
	userRepo := &repos.UserRepository{DB: suite.db}
	tokenRepo := &repos.TokenRepository{
		DB: suite.db,
		Config: &repos.TokenConfig{
			BearerDuration: 3600 * time.Second,
			ResetDuration:  3600 * time.Second,
			VerifyDuration: 3600 * time.Second,
		},
	}

	// create the auth service
	authService := &service.AuthService{
		UserRepo:  userRepo,
		TokenRepo: tokenRepo,
	}

	ctx := context.Background()

	testCases := []struct {
		desc          string
		request       *pb.BearerTokenRequest
		expectedError error
	}{
		{"missing email", &pb.BearerTokenRequest{}, status.Error(codes.InvalidArgument, "email is required")},
		{"invalid email", &pb.BearerTokenRequest{Email: "invalid"}, status.Error(codes.InvalidArgument, "invalid email format")},
		{"missing password", &pb.BearerTokenRequest{Email: user.Email}, status.Error(codes.InvalidArgument, "password is required")},
		{"invalid password", &pb.BearerTokenRequest{Email: user.Email, Password: "invalid"}, status.Error(codes.InvalidArgument, "invalid credentials")},
		{"invalid email and password", &pb.BearerTokenRequest{Email: "invalid@example.com", Password: "invalid"}, status.Error(codes.InvalidArgument, "invalid credentials")},
	}

	for _, tc := range testCases {
		suite.Run(tc.desc, func() {
			resp, err := authService.BearerToken(ctx, tc.request)
			suite.EqualError(err, tc.expectedError.Error())
			suite.Nil(resp)
		})
	}

	// test with valid credentials
	resp, err := authService.BearerToken(ctx, &pb.BearerTokenRequest{Email: user.Email, Password: "password"})
	suite.NoError(err)

	var token models.AccessToken
	err = suite.db.Where("user_id = ?", user.ID).First(&token).Error
	suite.NoError(err)

	suite.Equal(&pb.BearerTokenResponse{
		AccessToken: token.Token,
		TokenType:   "bearer",
		Expiry:      3600,
	}, resp)
}

func (suite *TestSuite) TestAuthService_RevokeBearerToken() {
	teardown := suite.Setup()
	defer teardown()

	user, err := suite.helpers.CreateTestUser()
	suite.NoError(err)

	// create the repos
	userRepo := &repos.UserRepository{DB: suite.db}
	tokenRepo := &repos.TokenRepository{
		DB: suite.db,
		Config: &repos.TokenConfig{
			BearerDuration: 3600 * time.Second,
			ResetDuration:  3600 * time.Second,
			VerifyDuration: 3600 * time.Second,
		},
	}

	// create the auth service
	authService := &service.AuthService{
		UserRepo:  userRepo,
		TokenRepo: tokenRepo,
	}

	ctx := context.Background()

	testCases := []struct {
		desc          string
		request       *pb.Token
		expectedError error
	}{
		{"missing token", &pb.Token{}, status.Error(codes.InvalidArgument, "token is required")},
		{"invalid token", &pb.Token{Token: "123"}, nil},
	}

	for _, tc := range testCases {
		suite.Run(tc.desc, func() {
			resp, err := authService.RevokeBearerToken(ctx, tc.request)
			if tc.expectedError != nil {
				suite.EqualError(err, tc.expectedError.Error())
				suite.Nil(resp)
			} else {
				suite.NoError(err)
				suite.Equal(&pb.Empty{}, resp)
			}
		})
	}

	// Test with valid token
	token, tokenErr := tokenRepo.CreateAccessToken(user.ID)
	suite.NoError(tokenErr)

	resp, err := authService.RevokeBearerToken(ctx, &pb.Token{Token: token.Token})
	suite.NoError(err)
	suite.Equal(&pb.Empty{}, resp)

	err = suite.db.First(&token).Error
	suite.Error(err)
	suite.Equal(err.Error(), "record not found")
}

func (suite *TestSuite) TestAuthService_Register() {
	teardown := suite.Setup()
	defer teardown()

	// create the repos
	userRepo := &repos.UserRepository{DB: suite.db}
	tokenRepo := &repos.TokenRepository{
		DB: suite.db,
		Config: &repos.TokenConfig{
			BearerDuration: 3600 * time.Second,
			ResetDuration:  3600 * time.Second,
			VerifyDuration: 3600 * time.Second,
		},
	}

	// create the auth service
	authService := &service.AuthService{
		UserRepo:  userRepo,
		TokenRepo: tokenRepo,
	}

	ctx := context.Background()

	// test no default user type exists
	resp, err := authService.Register(ctx, &pb.RegisterRequest{})
	suite.EqualError(err, status.Error(codes.Internal, "no default user type exists").Error())
	suite.Nil(resp)

	_, err = suite.helpers.CreateTestUser()
	suite.NoError(err)

	testCases := []struct {
		desc          string
		request       *pb.RegisterRequest
		expectedError error
	}{
		{"missing email", &pb.RegisterRequest{}, status.Error(codes.InvalidArgument, "email is required")},
		{"invalid email", &pb.RegisterRequest{Email: "invalid"}, status.Error(codes.InvalidArgument, "invalid email format")},
		{"missing first name", &pb.RegisterRequest{Email: "test@test.com"}, status.Error(codes.InvalidArgument, "first_name is required")},
		{"missing last name", &pb.RegisterRequest{Email: "test@test.com", FirstName: "Some"}, status.Error(codes.InvalidArgument, "last_name is required")},
		{"missing password", &pb.RegisterRequest{Email: "test@test.com", FirstName: "Some", LastName: "One"}, status.Error(codes.InvalidArgument, "password is required")},
		{"invalid password", &pb.RegisterRequest{Email: "test@test.com", FirstName: "Some", LastName: "One", Password: "123"}, status.Error(codes.InvalidArgument, "password must be between 6 and 72 characters in length")},
	}

	_, err = suite.helpers.CreateTestUserType()
	suite.NoError(err)

	for _, tc := range testCases {
		suite.Run(tc.desc, func() {
			resp, err := authService.Register(ctx, tc.request)
			suite.EqualError(err, tc.expectedError.Error())
			suite.Nil(resp)
		})
	}

	// Test valid registration
	successTestCases := []struct {
		desc    string
		request *pb.RegisterRequest
		delete  bool
	}{
		{"good response", &pb.RegisterRequest{Email: "test@test.com", Password: "password", FirstName: "Some", LastName: "One"}, true},
		{"lowercase email", &pb.RegisterRequest{Email: "TesT@teSt.com", Password: "password", FirstName: "Some", LastName: "One"}, true},
		{"trim all", &pb.RegisterRequest{Email: " test@test.com   ", Password: "password", FirstName: " Some ", LastName: " One "}, false},
	}

	for _, tc := range successTestCases {
		suite.Run(tc.desc, func() {
			resp, err := authService.Register(ctx, tc.request)
			suite.NoError(err)

			var fetchUser models.User
			err = suite.db.Preload("UserType").Preload("UserType.Scopes").First(&fetchUser, "email = 'test@test.com'").Error
			suite.NoError(err)

			suite.Equal(fetchUser.Email, "test@test.com")
			suite.Equal(fetchUser.FirstName, "Some")
			suite.Equal(fetchUser.LastName, "One")

			suite.Equal(&pb.UserResponse{
				Id:        fetchUser.ID.String(),
				FirstName: fetchUser.FirstName,
				LastName:  fetchUser.LastName,
				Email:     fetchUser.Email,
				UserType: &pb.UserType{
					Name:   fetchUser.UserType.Name,
					Scopes: fetchUser.UserType.ScopeNames(),
				},
				IsActive:   fetchUser.IsActive,
				IsVerified: fetchUser.IsVerified,
			}, resp)

			if tc.delete {
				err = suite.db.Delete(&fetchUser).Error
				suite.NoError(err)
			}
		})
	}

	// Test valid registration, but user already registered
	resp, err = authService.Register(ctx, &pb.RegisterRequest{Email: "test@test.com", Password: "password", FirstName: "Some", LastName: "One"})
	suite.EqualError(err, status.Error(codes.AlreadyExists, "a user with this email already exists").Error())
	suite.Nil(resp)

	// Test valid registration, email manipulation safety
	resp, err = authService.Register(ctx, &pb.RegisterRequest{Email: " TEST@TEST.COM ", Password: "password", FirstName: "Some", LastName: "One"})
	suite.EqualError(err, status.Error(codes.AlreadyExists, "a user with this email already exists").Error())
	suite.Nil(resp)
}

func (suite *TestSuite) TestAuthService_ResetPassword() {
	teardown := suite.Setup()
	defer teardown()

	// create the repos
	userRepo := &repos.UserRepository{DB: suite.db}
	tokenRepo := &repos.TokenRepository{
		DB: suite.db,
		Config: &repos.TokenConfig{
			BearerDuration: 3600 * time.Second,
			ResetDuration:  3600 * time.Second,
			VerifyDuration: 3600 * time.Second,
		},
	}

	// create the auth service
	authService := &service.AuthService{
		UserRepo:  userRepo,
		TokenRepo: tokenRepo,
	}

	user, err := suite.helpers.CreateTestUser()
	suite.NoError(err)

	ctx := context.Background()

	token := &models.ResetToken{UserId: user.ID, Token: "valid-token", ExpiresAt: time.Now().Add(1 * time.Minute)}
	tokenErr := suite.db.Create(token).Error
	suite.NoError(tokenErr)

	testCases := []struct {
		desc          string
		request       *pb.ResetPasswordRequest
		expectedError error
	}{
		{"missing token", &pb.ResetPasswordRequest{}, status.Error(codes.InvalidArgument, "token is required")},
		{"invalid token", &pb.ResetPasswordRequest{Token: "invalid-token"}, status.Error(codes.InvalidArgument, "invalid token")},
		{"missing password", &pb.ResetPasswordRequest{Token: token.Token}, status.Error(codes.InvalidArgument, "password is required")},
		{"invalid password", &pb.ResetPasswordRequest{Token: token.Token, Password: "short"}, status.Error(codes.InvalidArgument, "password must be between 6 and 72 characters in length")},
	}

	for _, tc := range testCases {
		suite.Run(tc.desc, func() {
			resp, err := authService.ResetPassword(ctx, tc.request)
			suite.EqualError(err, tc.expectedError.Error())
			suite.Nil(resp)
		})
	}

	// Test reset password
	resp, err := authService.ResetPassword(ctx, &pb.ResetPasswordRequest{Token: token.Token, Password: "password"})
	suite.NoError(err)
	suite.Equal(&pb.Empty{}, resp)

	// Test cannot reuse token
	resp, err = authService.ResetPassword(ctx, &pb.ResetPasswordRequest{Token: token.Token, Password: "password"})
	suite.EqualError(err, status.Error(codes.InvalidArgument, "invalid token").Error())
	suite.Nil(resp)
}

func (suite *TestSuite) TestAuthService_ResetPasswordToken() {
	teardown := suite.Setup()
	defer teardown()

	// create the repos
	userRepo := &repos.UserRepository{DB: suite.db}
	tokenRepo := &repos.TokenRepository{
		DB: suite.db,
		Config: &repos.TokenConfig{
			BearerDuration: 3600 * time.Second,
			ResetDuration:  3600 * time.Second,
			VerifyDuration: 3600 * time.Second,
		},
	}

	// create the auth service
	authService := &service.AuthService{
		UserRepo:  userRepo,
		TokenRepo: tokenRepo,
	}

	user, err := suite.helpers.CreateTestUser()
	suite.NoError(err)

	ctx := context.Background()

	testCases := []struct {
		desc          string
		request       *pb.ResetPasswordTokenRequest
		expectedError error
	}{
		{"missing email", &pb.ResetPasswordTokenRequest{}, status.Error(codes.InvalidArgument, "email is required")},
		{"invalid email", &pb.ResetPasswordTokenRequest{Email: "invalid"}, status.Error(codes.InvalidArgument, "invalid email format")},
		{"unknown email", &pb.ResetPasswordTokenRequest{Email: "invalid@mail.com"}, status.Error(codes.NotFound, "user not found")},
	}

	for _, tc := range testCases {
		suite.Run(tc.desc, func() {
			resp, err := authService.ResetPasswordToken(ctx, tc.request)
			suite.EqualError(err, tc.expectedError.Error())
			suite.Nil(resp)
		})
	}

	// Test valid email
	resp, err := authService.ResetPasswordToken(ctx, &pb.ResetPasswordTokenRequest{Email: user.Email})
	suite.NoError(err)

	var token models.ResetToken
	err = suite.db.Where("user_id = ?", user.ID).First(&token).Error
	suite.NoError(err)

	suite.Equal(&pb.TokenWithEmail{
		Token:     token.Token,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}, resp)
}

func (suite *TestSuite) TestAuthService_User() {
	teardown := suite.Setup()
	defer teardown()

	// create the repos
	userRepo := &repos.UserRepository{DB: suite.db}
	tokenRepo := &repos.TokenRepository{
		DB: suite.db,
		Config: &repos.TokenConfig{
			BearerDuration: 3600 * time.Second,
			ResetDuration:  3600 * time.Second,
			VerifyDuration: 3600 * time.Second,
		},
	}

	// create the auth service
	authService := &service.AuthService{
		UserRepo:  userRepo,
		TokenRepo: tokenRepo,
	}

	user, err := suite.helpers.CreateTestUser()
	suite.NoError(err)

	ctx := context.Background()

	testCases := []struct {
		desc          string
		token         *models.AccessToken
		request       *pb.Token
		expectedError error
	}{
		{"missing token", nil, &pb.Token{}, status.Error(codes.InvalidArgument, "token is required")},
		{"invalid token", nil, &pb.Token{Token: "invalid"}, status.Error(codes.InvalidArgument, "invalid token")},
		{"expired token", &models.AccessToken{UserId: user.ID, Token: "expired-token", ExpiresAt: time.Now().Add(-1 * time.Second)}, &pb.Token{Token: "expired-token"}, status.Error(codes.InvalidArgument, "invalid token")},
	}

	for _, tc := range testCases {
		suite.Run(tc.desc, func() {
			if tc.token != nil {
				tokenErr := suite.db.Create(tc.token).Error
				suite.NoError(tokenErr)
			}

			resp, err := authService.User(ctx, tc.request)
			suite.EqualError(err, tc.expectedError.Error())
			suite.Nil(resp)
		})
	}

	// Test valid token
	token := &models.AccessToken{UserId: user.ID, Token: "valid-token", ExpiresAt: time.Now().Add(10 * time.Second)}
	tokenErr := suite.db.Create(token).Error
	suite.NoError(tokenErr)

	resp, err := authService.User(ctx, &pb.Token{Token: token.Token})
	suite.NoError(err)

	err = suite.db.Preload("UserType").Preload("UserType.Scopes").First(&user, "id = ?", user.ID).Error
	suite.NoError(err)
	suite.Equal(&pb.UserResponse{
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
	}, resp)
}

func (suite *TestSuite) TestAuthService_UpdateUser() {
	teardown := suite.Setup()
	defer teardown()

	// create the repos
	userRepo := &repos.UserRepository{DB: suite.db}
	tokenRepo := &repos.TokenRepository{
		DB: suite.db,
		Config: &repos.TokenConfig{
			BearerDuration: 3600 * time.Second,
			ResetDuration:  3600 * time.Second,
			VerifyDuration: 3600 * time.Second,
		},
	}

	// create the auth service
	authService := &service.AuthService{
		UserRepo:  userRepo,
		TokenRepo: tokenRepo,
	}

	ctx := context.Background()

	user, err := suite.helpers.CreateTestUser()
	suite.NoError(err)

	// Fail cases
	failTestCases := []struct {
		desc          string
		token         *models.AccessToken
		request       *pb.UpdateUserRequest
		expectedError error
	}{
		{"missing token", nil, &pb.UpdateUserRequest{}, status.Error(codes.InvalidArgument, "token is required")},
		{"invalid token", nil, &pb.UpdateUserRequest{Token: "123"}, status.Error(codes.InvalidArgument, "invalid token")},
		{"expired token", &models.AccessToken{UserId: user.ID, Token: "expired-token", ExpiresAt: time.Now().Add(-1 * time.Second)}, &pb.UpdateUserRequest{Token: "expired-token"}, status.Error(codes.InvalidArgument, "invalid token")},
		{"short password", &models.AccessToken{UserId: user.ID, Token: "valid-one", ExpiresAt: time.Now().Add(1 * time.Minute)}, &pb.UpdateUserRequest{Token: "valid-one", Password: "short"}, status.Error(codes.InvalidArgument, "password must be between 6 and 72 characters in length")},
		{"invalid email", &models.AccessToken{UserId: user.ID, Token: "valid-two", ExpiresAt: time.Now().Add(1 * time.Minute)}, &pb.UpdateUserRequest{Token: "valid-two", Email: "invalid"}, status.Error(codes.InvalidArgument, "invalid email format")},
	}

	for _, tc := range failTestCases {
		if tc.token != nil {
			tokenErr := suite.db.Create(tc.token).Error
			suite.NoError(tokenErr)
		}

		suite.Run(tc.desc, func() {
			resp, err := authService.UpdateUser(ctx, tc.request)
			suite.EqualError(err, tc.expectedError.Error())
			suite.Nil(resp)
		})
	}

	// Success cases

	token := &models.AccessToken{UserId: user.ID, Token: "another-token", ExpiresAt: time.Now().Add(10 * time.Minute)}
	err = suite.db.Create(token).Error
	suite.NoError(err)

	successTestCases := []struct {
		desc    string
		request *pb.UpdateUserRequest
	}{
		{"only email", &pb.UpdateUserRequest{Token: "another-token", Email: "some@one.com"}},
		{"only first name", &pb.UpdateUserRequest{Token: "another-token", FirstName: "Some"}},
		{"only last name", &pb.UpdateUserRequest{Token: "another-token", LastName: "One"}},
		{"only password", &pb.UpdateUserRequest{Token: "another-token", Password: "changed"}},
		{"first and last name", &pb.UpdateUserRequest{Token: "another-token", FirstName: "Someone", LastName: "Else"}},
		{"email and password", &pb.UpdateUserRequest{Token: "another-token", Email: "some@another.com", Password: "again?"}},
	}

	for _, tc := range successTestCases {

		suite.Run(tc.desc, func() {
			var originalUser models.User
			err = suite.db.First(&originalUser, "id = ?", user.ID).Error

			resp, err := authService.UpdateUser(ctx, tc.request)
			suite.NoError(err)

			var fetchedUser models.User
			err = suite.db.Preload("UserType").Preload("UserType.Scopes").First(&fetchedUser, "id = ?", user.ID).Error

			if tc.request.Email != "" {
				suite.Equal(tc.request.Email, fetchedUser.Email)
			} else {
				suite.Equal(originalUser.Email, fetchedUser.Email)
			}
			if tc.request.FirstName != "" {
				suite.Equal(tc.request.FirstName, fetchedUser.FirstName)
			} else {
				suite.Equal(originalUser.FirstName, fetchedUser.FirstName)
			}
			if tc.request.LastName != "" {
				suite.Equal(tc.request.LastName, fetchedUser.LastName)
			} else {
				suite.Equal(originalUser.LastName, fetchedUser.LastName)
			}
			if tc.request.Password != "" {
				suite.True(fetchedUser.VerifyPassword(tc.request.Password))
			}

			suite.Equal(&pb.UserResponse{
				Id:        fetchedUser.ID.String(),
				FirstName: fetchedUser.FirstName,
				LastName:  fetchedUser.LastName,
				Email:     fetchedUser.Email,
				UserType: &pb.UserType{
					Name:   fetchedUser.UserType.Name,
					Scopes: fetchedUser.UserType.ScopeNames(),
				},
				IsActive:   fetchedUser.IsActive,
				IsVerified: fetchedUser.IsVerified,
			}, resp)
		})
	}
}

func (suite *TestSuite) TestAuthService_VerifyUser() {
	teardown := suite.Setup()
	defer teardown()

	// create the repos
	userRepo := &repos.UserRepository{DB: suite.db}
	tokenRepo := &repos.TokenRepository{
		DB: suite.db,
		Config: &repos.TokenConfig{
			BearerDuration: 3600 * time.Second,
			ResetDuration:  3600 * time.Second,
			VerifyDuration: 3600 * time.Second,
		},
	}

	// create the auth service
	authService := &service.AuthService{
		UserRepo:  userRepo,
		TokenRepo: tokenRepo,
	}

	user, err := suite.helpers.CreateTestUser()
	suite.NoError(err)
	suite.False(user.IsVerified)

	ctx := context.Background()

	testCases := []struct {
		desc          string
		token         *models.VerifyToken
		request       *pb.Token
		expectedError error
	}{
		{"missing token", nil, &pb.Token{}, status.Error(codes.InvalidArgument, "token is required")},
		{"invalid token", nil, &pb.Token{Token: "invalid"}, status.Error(codes.InvalidArgument, "invalid token")},
		{"expired token", &models.VerifyToken{UserId: user.ID, Token: "expired-token", ExpiresAt: time.Now().Add(-1 * time.Second)}, &pb.Token{Token: "expired-token"}, status.Error(codes.InvalidArgument, "invalid token")},
	}

	for _, tc := range testCases {
		suite.Run(tc.desc, func() {
			if tc.token != nil {
				tokenErr := suite.db.Create(tc.token).Error
				suite.NoError(tokenErr)
			}

			resp, err := authService.VerifyUser(ctx, tc.request)
			suite.EqualError(err, tc.expectedError.Error())
			suite.Nil(resp)
		})
	}

	// Test valid token
	token := &models.VerifyToken{UserId: user.ID, Token: "valid-token", ExpiresAt: time.Now().Add(10 * time.Second)}
	err = suite.db.Create(token).Error
	suite.NoError(err)

	resp, err := authService.VerifyUser(ctx, &pb.Token{Token: token.Token})
	suite.NoError(err)

	err = suite.db.Preload("UserType").Preload("UserType.Scopes").First(&user, "id = ?", user.ID).Error
	suite.NoError(err)
	suite.True(user.IsVerified)

	suite.Equal(&pb.UserResponse{
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
	}, resp)
}

func (suite *TestSuite) TestAuthService_VerifyUserToken() {
	teardown := suite.Setup()
	defer teardown()

	// create the repos
	userRepo := &repos.UserRepository{DB: suite.db}
	tokenRepo := &repos.TokenRepository{
		DB: suite.db,
		Config: &repos.TokenConfig{
			BearerDuration: 3600 * time.Second,
			ResetDuration:  3600 * time.Second,
			VerifyDuration: 3600 * time.Second,
		},
	}

	// create the auth service
	authService := &service.AuthService{
		UserRepo:  userRepo,
		TokenRepo: tokenRepo,
	}

	user, err := suite.helpers.CreateTestUser()
	suite.NoError(err)
	suite.False(user.IsVerified)

	ctx := context.Background()

	testCases := []struct {
		desc          string
		request       *pb.VerifyUserTokenRequest
		expectedError error
	}{
		{"missing email", &pb.VerifyUserTokenRequest{}, status.Error(codes.InvalidArgument, "email is required")},
		{"invalid email", &pb.VerifyUserTokenRequest{Email: "invalid"}, status.Error(codes.InvalidArgument, "invalid email format")},
		{"unknown email", &pb.VerifyUserTokenRequest{Email: "unknown@mail.com"}, status.Error(codes.NotFound, "user not found")},
	}

	for _, tc := range testCases {
		suite.Run(tc.desc, func() {
			resp, err := authService.VerifyUserToken(ctx, tc.request)
			suite.EqualError(err, tc.expectedError.Error())
			suite.Nil(resp)
		})
	}

	// Test valid email
	resp, err := authService.VerifyUserToken(ctx, &pb.VerifyUserTokenRequest{Email: user.Email})
	suite.NoError(err)

	var token models.VerifyToken
	err = suite.db.Where("user_id = ?", user.ID).First(&token).Error
	suite.NoError(err)

	suite.Equal(&pb.TokenWithEmail{
		Token:     token.Token,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}, resp)

	// Test already verified
	user.IsVerified = true
	err = suite.db.Save(&user).Error
	suite.NoError(err)

	resp, err = authService.VerifyUserToken(ctx, &pb.VerifyUserTokenRequest{Email: user.Email})
	suite.EqualError(err, status.Error(codes.FailedPrecondition, "user is already verified").Error())
	suite.Nil(resp)
}
