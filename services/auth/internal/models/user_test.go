package models_test

import (
	"errors"
	"github.com/accentdesign/grpc/services/auth/internal/models"
	"strings"
)

func (suite *TestSuite) TestUserModel_Validate() {
	testCases := []struct {
		desc          string
		user          *models.User
		expectedError error
	}{
		{"missing email", &models.User{}, errors.New("email is required")},
		{"empty email", &models.User{Email: ""}, errors.New("email is required")},
		{"invalid email", &models.User{Email: "invalid"}, errors.New("invalid email format")},
		{"missing first name", &models.User{Email: "test@example.com"}, errors.New("first_name is required")},
		{"empty first name", &models.User{Email: "test@example.com", FirstName: ""}, errors.New("first_name is required")},
		{"missing last name", &models.User{Email: "test@example.com", FirstName: "Some"}, errors.New("last_name is required")},
		{"empty last name", &models.User{Email: "test@example.com", FirstName: "Some", LastName: ""}, errors.New("last_name is required")},
	}

	for _, tc := range testCases {
		suite.Run(tc.desc, func() {
			err := tc.user.Validate()
			suite.EqualError(tc.expectedError, err.Error())
		})
	}

	// Test a valid user
	validUser := &models.User{Email: "test@example.com", HashedPassword: "password", FirstName: "Test", LastName: "User"}
	err := validUser.Validate()
	suite.NoError(err)
}

func (suite *TestSuite) TestUserModel_SetPassword() {
	testCases := []struct {
		desc          string
		password      string
		expectedError error
	}{
		{"empty", "", errors.New("password is required")},
		{"too short", strings.Repeat("x", 5), errors.New("password must be between 6 and 72 characters in length")},
		{"too long", strings.Repeat("x", 73), errors.New("password must be between 6 and 72 characters in length")},
		{"at min length", strings.Repeat("x", 6), nil},
		{"at max length", strings.Repeat("x", 72), nil},
		{"ok", "password", nil},
	}

	for _, tc := range testCases {
		suite.Run(tc.desc, func() {
			user := &models.User{}
			err := user.SetPassword(tc.password)
			if tc.expectedError != nil {
				suite.EqualError(tc.expectedError, err.Error())
			} else {
				suite.NoError(err)
			}
		})
	}
}

func (suite *TestSuite) TestUserModel_VerifyPassword() {
	// Test a valid user
	user := &models.User{}
	err := user.SetPassword("password")
	suite.NoError(err)

	// correct
	verify := user.VerifyPassword("password")
	suite.True(verify)

	// incorrect
	verify = user.VerifyPassword("password1")
	suite.False(verify)
}
