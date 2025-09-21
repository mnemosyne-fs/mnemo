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

	assert.Equal(suite.T(), nil, err)
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
}

func (suite *DatabaseTestSuite) TestCreation() {
	createdDatabase, err := CreateAuthDatabase("testingDatabase.json")
	createdDatabase.Filename = "authTemplate.json"

	assert.Equal(suite.T(), nil, err)
	assert.Equal(suite.T(), suite.Database, createdDatabase)

	os.Remove("testingDatabase.json")
}

func (suite *DatabaseTestSuite) TestAccountOperations() {
	testingDatabase, err := CreateAuthDatabase("testingDatabase.json")
	assert.Equal(suite.T(), nil, err)

	newUser := make(map[string]string)
	newUser["test"] = "test"

	testingDatabase.CreateUser("test")
	assert.Contains(suite.T(), testingDatabase.Users, "test")

	testingDatabase.RemoveUser("test")
	assert.NotContains(suite.T(), testingDatabase.Users, "test")
}

func TestDatabaseTestSuite(t *testing.T) {
	suite.Run(t, new(DatabaseTestSuite))
}
