package repos_test

import (
	"time"

	"github.com/accentdesign/grpc/services/auth/internal/models"
	"github.com/accentdesign/grpc/services/auth/internal/repos"
)

func (suite *TestSuite) TestUserRepository_CreateUser() {
	teardown := suite.Setup()
	defer teardown()

	userType, err := suite.helpers.CreateTestUserType()
	suite.NoError(err)

	repo := repos.UserRepository{DB: suite.db}

	user, err := repo.CreateUser("a@b.com", "password", "Some", "One")
	suite.NoError(err)

	suite.NotEmpty(user.ID)
	suite.Equal(user.Email, "a@b.com")
	suite.Equal(user.FirstName, "Some")
	suite.Equal(user.LastName, "One")
	suite.Equal(user.UserTypeId, userType.ID)
	suite.NotEmpty(user.HashedPassword)
	suite.True(user.VerifyPassword("password"))
	suite.False(user.VerifyPassword("not-password"))
	suite.True(user.IsActive)
	suite.False(user.IsVerified)
	suite.WithinDuration(time.Now().Add(-5*time.Second), user.CreatedAt, 5*time.Second)
	suite.WithinDuration(time.Now().Add(-5*time.Second), user.UpdatedAt, 5*time.Second)
}

func (suite *TestSuite) TestUserRepository_CreateUser_NoUserType() {
	teardown := suite.Setup()
	defer teardown()

	repo := repos.UserRepository{DB: suite.db}

	user, err := repo.CreateUser("a@b.com", "password", "Some", "One")
	suite.Error(err)
	suite.Equal("no default user type exists", err.Error())
	suite.Nil(user)
}

func (suite *TestSuite) TestUserRepository_CreateUser_Validate() {
	teardown := suite.Setup()
	defer teardown()

	_, err := suite.helpers.CreateTestUserType()
	suite.NoError(err)

	repo := repos.UserRepository{DB: suite.db}

	user, err := repo.CreateUser("", "password", "Some", "One")

	suite.Error(err)
	suite.Equal("invalid email format", err.Error())
	suite.Nil(user)
}

func (suite *TestSuite) TestUserRepository_CreateUser_DBError() {
	teardown := suite.Setup()
	defer teardown()

	user, err := suite.helpers.CreateTestUser()
	suite.NoError(err)

	repo := repos.UserRepository{DB: suite.db}

	duplicate, err := repo.CreateUser(user.Email, "password", "Some", "One")

	suite.Error(err)
	suite.Equal("duplicated key not allowed", err.Error())
	suite.Nil(duplicate)
}

func (suite *TestSuite) TestUserRepository_GetUserByEmail() {
	teardown := suite.Setup()
	defer teardown()

	user, err := suite.helpers.CreateTestUser()
	suite.NoError(err)

	repo := repos.UserRepository{DB: suite.db}

	fetchedUser, err := repo.GetUserByEmail(user.Email)
	suite.NoError(err)
	suite.Equal(user.ID, fetchedUser.ID)
}

func (suite *TestSuite) TestUserRepository_GetUserByEmail_Error() {
	teardown := suite.Setup()
	defer teardown()

	repo := repos.UserRepository{DB: suite.db}

	user, err := repo.GetUserByEmail("test@example.com")
	suite.Error(err)
	suite.Equal("user not found: test@example.com", err.Error())
	suite.Nil(user)
}

func (suite *TestSuite) TestUserRepository_GetUserByAccessToken() {
	teardown := suite.Setup()
	defer teardown()

	user, err := suite.helpers.CreateTestUser()
	suite.NoError(err)

	token := &models.AccessToken{UserId: user.ID, Token: "test_token", ExpiresAt: time.Now().Add(24 * time.Hour)}
	tokenErr := suite.db.Create(token).Error
	suite.NoError(tokenErr)

	repo := repos.UserRepository{DB: suite.db}

	fetchedUser, err := repo.GetUserByAccessToken(token.Token)

	suite.NoError(err)
	suite.Equal(user.ID, fetchedUser.ID)
}

func (suite *TestSuite) TestUserRepository_GetUserByAccessToken_Expired() {
	teardown := suite.Setup()
	defer teardown()

	user, err := suite.helpers.CreateTestUser()
	suite.NoError(err)

	token := &models.AccessToken{UserId: user.ID, Token: "test_token", ExpiresAt: time.Now().Add(-1 * time.Second)}
	tokenErr := suite.db.Create(token).Error
	suite.NoError(tokenErr)

	repo := repos.UserRepository{DB: suite.db}

	fetchedUser, err := repo.GetUserByAccessToken(token.Token)

	suite.Error(err)
	suite.Equal("user not found for token: test_token", err.Error())
	suite.Nil(fetchedUser)
}

func (suite *TestSuite) TestUserRepository_GetUserByResetToken() {
	teardown := suite.Setup()
	defer teardown()

	user, err := suite.helpers.CreateTestUser()
	suite.NoError(err)

	token := &models.ResetToken{UserId: user.ID, Token: "test_token", ExpiresAt: time.Now().Add(24 * time.Hour)}
	tokenErr := suite.db.Create(token).Error
	suite.NoError(tokenErr)

	repo := repos.UserRepository{DB: suite.db}

	fetchedUser, err := repo.GetUserByResetToken(token.Token)

	suite.NoError(err)
	suite.Equal(user.ID, fetchedUser.ID)
}

func (suite *TestSuite) TestUserRepository_GetUserByResetToken_Expired() {
	teardown := suite.Setup()
	defer teardown()

	user, err := suite.helpers.CreateTestUser()
	suite.NoError(err)

	token := &models.ResetToken{UserId: user.ID, Token: "test_token", ExpiresAt: time.Now().Add(-1 * time.Second)}
	tokenErr := suite.db.Create(token).Error
	suite.NoError(tokenErr)

	repo := repos.UserRepository{DB: suite.db}

	fetchedUser, err := repo.GetUserByResetToken(token.Token)

	suite.Error(err)
	suite.Equal("user not found for token: test_token", err.Error())
	suite.Nil(fetchedUser)
}

func (suite *TestSuite) TestUserRepository_GetUserByVerifyToken() {
	teardown := suite.Setup()
	defer teardown()

	user, err := suite.helpers.CreateTestUser()
	suite.NoError(err)

	token := &models.VerifyToken{UserId: user.ID, Token: "test_token", ExpiresAt: time.Now().Add(24 * time.Hour)}
	tokenErr := suite.db.Create(token).Error
	suite.NoError(tokenErr)

	repo := repos.UserRepository{DB: suite.db}

	fetchedUser, err := repo.GetUserByVerifyToken(token.Token)

	suite.NoError(err)
	suite.Equal(user.ID, fetchedUser.ID)
}

func (suite *TestSuite) TestUserRepository_GetUserByVerifyToken_Expired() {
	teardown := suite.Setup()
	defer teardown()

	user, err := suite.helpers.CreateTestUser()
	suite.NoError(err)

	token := &models.VerifyToken{UserId: user.ID, Token: "test_token", ExpiresAt: time.Now().Add(-1 * time.Second)}
	tokenErr := suite.db.Create(token).Error
	suite.NoError(tokenErr)

	repo := repos.UserRepository{DB: suite.db}

	fetchedUser, err := repo.GetUserByVerifyToken(token.Token)

	suite.Error(err)
	suite.Equal("user not found for token: test_token", err.Error())
	suite.Nil(fetchedUser)
}

func (suite *TestSuite) TestUserRepository_UpdateUser() {
	teardown := suite.Setup()
	defer teardown()

	user, err := suite.helpers.CreateTestUser()
	suite.NoError(err)

	repo := repos.UserRepository{DB: suite.db}

	user.FirstName = "Bill"

	err = repo.UpdateUser(user)
	suite.NoError(err)

	user.Email = ""

	err = repo.UpdateUser(user)
	suite.Error(err)
	suite.Equal("invalid email format", err.Error())
}
