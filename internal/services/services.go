package services

import (
	"github.com/charmbracelet/log"
	authentication "github.com/mnemosynefs/mnemo/internal/authentication"
	networking "github.com/mnemosynefs/mnemo/internal/networking"
)

type Services struct {
	Database authentication.Database
	Mnemo    *networking.MnemoServer
}

func CreateServices(
	serverAddress string,
	databaseFilename string,
	fileOps ...authentication.FileInterface,
) (*Services, error) {
	mnemo := networking.CreateMnemoServer(serverAddress)

	var database *authentication.AuthDatabase
	var err error
	if len(fileOps) > 0 {
		database, err = authentication.CreateNewDatabase(databaseFilename, fileOps[0])
	} else {
		database, err = authentication.CreateNewDatabase(databaseFilename)
	}

	if err != nil {
		log.Errorf(
			"Failed to create new database at location %v. Program abort recommended.",
			databaseFilename,
		)
		return nil, err
	}

	return &Services{
		Database: database,
		Mnemo:    mnemo,
	}, nil
}
