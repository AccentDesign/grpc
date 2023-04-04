package models_test

import (
	"github.com/accentdesign/grpc/services/auth/internal/models"
)

func (suite *TestSuite) TestUserType_ScopeNames() {
	userType := models.UserType{
		Name: "user",
		Scopes: []models.Scope{
			{Name: "read"},
			{Name: "write"},
			{Name: "delete"},
		},
	}
	scopeNames := userType.ScopeNames()
	expected := []string{"read", "write", "delete"}
	suite.Equal(expected, scopeNames)

	userType = models.UserType{
		Name: "user",
	}
	scopeNames = userType.ScopeNames()
	suite.Empty(scopeNames)
}
