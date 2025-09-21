package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DatabaseTestSuite struct {
	suite.Suite
	Database *AuthDatabase
	filename string
}

func (suite *DatabaseTestSuite) SetupTest() {
	var err error
	suite.filename = "authTemplate.json"
	suite.Database, err = LoadAuthDatabase(suite.filename)

	assert.Nil(suite.T(), err)
}

func (suite *DatabaseTestSuite) TestLoad() {
	templateAdmins := []string{"admin"}
	templatePermissions := make(map[string]UserPermission)
	templateSessions := make(map[string]Session)
	templateSharedFiles := make(map[string]SharedFile)
	templateUsers := make(map[string]string)

	templatePermissions["/"] = make(UserPermission)
	templatePermissions["/"]["admin"] = 3
	templateUsers["admin"] = "admin"

	// Check for correct initialisation
	assert.Equal(suite.T(), templateAdmins, suite.Database.Admin)
	assert.Equal(suite.T(), suite.filename, suite.Database.Filename)
	assert.Equal(suite.T(), templatePermissions, suite.Database.Permissions)
	assert.Equal(suite.T(), templateSessions, suite.Database.Sessions)
	assert.Equal(suite.T(), templateSharedFiles, suite.Database.Shared_files)
	assert.Equal(suite.T(), templateUsers, suite.Database.Users)

	// Try to load a database that doesn't exist
	failingDatabase, err := LoadAuthDatabase("failingDatabase.json")
	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), failingDatabase)
}

func (suite *DatabaseTestSuite) TestCreation() {
	createdDatabase, err := CreateAuthDatabase("testingDatabase.json")
	createdDatabase.Filename = "authTemplate.json"

	assert.Equal(suite.T(), nil, err)
	assert.Equal(suite.T(), suite.Database, createdDatabase)

	os.Remove("testingDatabase.json")
}

func (suite *DatabaseTestSuite) TestAccountOperations() {
	// Test creation
	username := "test2"
	password := "password"
	return_code := suite.Database.CreateUser(username)
	assert.Contains(suite.T(), suite.Database.Users, username)
	assert.Equal(suite.T(), SUCCESS, return_code)

	// Test creation of an existing user
	return_code = suite.Database.CreateUser(username)
	assert.Equal(suite.T(), USER_EXISTS, return_code)

	// Test logging in with incorrect password
	session_token, return_code, error := suite.Database.LoginUser(username, password)
	assert.Equal(suite.T(), "", session_token)
	assert.Equal(suite.T(), INVALID_LOGIN, return_code)
	assert.Nil(suite.T(), error)

	// Test logging in
	session_token, return_code, error = suite.Database.LoginUser(username, username)
	assert.Contains(suite.T(), suite.Database.Users, username)
	assert.Contains(suite.T(), suite.Database.Sessions, session_token)
	assert.Equal(suite.T(), SUCCESS, return_code)
	assert.Nil(suite.T(), error)

	// Test session validation
	assert.True(suite.T(), suite.Database.ValidateToken(session_token, username))
	assert.False(suite.T(), suite.Database.ValidateToken(session_token, password))
	new_token, error := suite.Database.GetSessionToken(username)
	assert.Nil(suite.T(), error)
	assert.Equal(suite.T(), new_token, session_token)

	// Test authentication checking
	assert.True(suite.T(), suite.Database.CheckAuth(username, username))
	assert.False(suite.T(), suite.Database.CheckAuth(password, username))

	// Test session updating
	return_code = suite.Database.UpdateSession(session_token, true)
	assert.Equal(suite.T(), SUCCESS, return_code)
	return_code = suite.Database.UpdateSession(session_token, false)
	assert.Equal(suite.T(), SUCCESS, return_code)

	// Test expired session
	suite.Database.CreateUser(password)
	new_token, _ = suite.Database.GetSessionToken(password)
	new_session := Session{
		username,
		0,
	}
	suite.Database.Sessions[new_token] = new_session
	return_code = suite.Database.UpdateSession(new_token, false)
	assert.Equal(suite.T(), INVALID_SESSION, return_code)
	assert.NotContains(suite.T(), suite.Database.Sessions, new_token)
	suite.Database.RemoveUser(password)

	// Test removal
	return_code = suite.Database.RemoveUser(username)
	assert.NotContains(suite.T(), suite.Database.Users, username)
	assert.NotContains(suite.T(), suite.Database.Sessions, session_token)
	assert.Equal(suite.T(), SUCCESS, return_code)
	suite.Database.Save()

	// Test logging into account that doesn't exist
	session_token, return_code, error = suite.Database.LoginUser(username, username)
	assert.Equal(suite.T(), "", session_token)
	assert.Equal(suite.T(), USER_DOESNT_EXIST, return_code)
	assert.Nil(suite.T(), error)

	// Test removal of account that doesn't exist
	return_code = suite.Database.RemoveUser(username)
	assert.Equal(suite.T(), USER_DOESNT_EXIST, return_code)
}

func TestDatabaseTestSuite(t *testing.T) {
	suite.Run(t, new(DatabaseTestSuite))
}
