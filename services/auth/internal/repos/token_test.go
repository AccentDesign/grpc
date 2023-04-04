package repos_test

import (
	"time"

	"github.com/accentdesign/grpc/services/auth/internal/models"
	"github.com/accentdesign/grpc/services/auth/internal/repos"
)

func (suite *TestSuite) TestTokenRepository_CreateAccessToken() {
	teardown := suite.Setup()
	defer teardown()

	config := &repos.TokenConfig{
		BearerDuration: time.Hour,
		ResetDuration:  time.Hour,
		VerifyDuration: time.Hour,
	}

	repo := repos.TokenRepository{
		DB:     suite.db,
		Config: config,
	}

	user, err := suite.helpers.CreateTestUser()
	suite.NoError(err)

	token, tokenErr := repo.CreateAccessToken(user.ID)
	suite.NoError(tokenErr)

	var found models.AccessToken
	err = suite.db.Where("token = ?", token.Token).First(&found).Error
	suite.NoError(err)

	suite.Equal(token.Token, found.Token)
	suite.Equal(token.UserId, found.UserId)
	suite.WithinDuration(time.Now().Add(config.BearerDuration), found.ExpiresAt, 10*time.Second)
}

func (suite *TestSuite) TestTokenRepository_CreateResetToken() {
	teardown := suite.Setup()
	defer teardown()

	config := &repos.TokenConfig{
		BearerDuration: time.Hour,
		ResetDuration:  time.Hour,
		VerifyDuration: time.Hour,
	}

	repo := repos.TokenRepository{
		DB:     suite.db,
		Config: config,
	}

	user, err := suite.helpers.CreateTestUser()
	suite.NoError(err)

	token, tokenErr := repo.CreateResetToken(user.ID)
	suite.NoError(tokenErr)

	var found models.ResetToken
	err = suite.db.Where("token = ?", token.Token).First(&found).Error
	suite.NoError(err)

	suite.Equal(token.Token, found.Token)
	suite.Equal(token.UserId, found.UserId)
	suite.WithinDuration(time.Now().Add(config.ResetDuration), found.ExpiresAt, 10*time.Second)
}

func (suite *TestSuite) TestTokenRepository_CreateVerifyToken() {
	teardown := suite.Setup()
	defer teardown()

	config := &repos.TokenConfig{
		BearerDuration: time.Hour,
		ResetDuration:  time.Hour,
		VerifyDuration: time.Hour,
	}

	repo := repos.TokenRepository{
		DB:     suite.db,
		Config: config,
	}

	user, err := suite.helpers.CreateTestUser()
	suite.NoError(err)

	token, tokenErr := repo.CreateVerifyToken(user.ID)
	suite.NoError(tokenErr)

	var found models.VerifyToken
	err = suite.db.Where("token = ?", token.Token).First(&found).Error
	suite.NoError(err)

	suite.Equal(token.Token, found.Token)
	suite.Equal(token.UserId, found.UserId)
	suite.WithinDuration(time.Now().Add(config.VerifyDuration), found.ExpiresAt, 10*time.Second)
}

func (suite *TestSuite) TestTokenRepository_RevokeBearerToken() {
	teardown := suite.Setup()
	defer teardown()

	config := &repos.TokenConfig{
		BearerDuration: time.Hour,
		ResetDuration:  time.Hour,
		VerifyDuration: time.Hour,
	}

	repo := repos.TokenRepository{
		DB:     suite.db,
		Config: config,
	}

	user, err := suite.helpers.CreateTestUser()
	suite.NoError(err)

	token := models.AccessToken{
		Token:     "some-token",
		UserId:    user.ID,
		ExpiresAt: time.Now(),
	}

	tokenErr := suite.db.Create(&token).Error
	suite.NoError(tokenErr)

	revokeErr := repo.RevokeBearerToken(token.Token)
	suite.NoError(revokeErr)

	var found models.AccessToken
	foundErr := suite.db.Where("token = ?", token.Token).First(&found).Error
	suite.Error(foundErr)
}
