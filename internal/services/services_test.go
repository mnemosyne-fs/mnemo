package services_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/mnemosynefs/mnemo/internal/services"
	"github.com/mnemosynefs/mnemo/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateServices_Success(t *testing.T) {
	tmp := t.TempDir()
	filename := filepath.Join(tmp, "auth.json")

	services, err := services.CreateServices(":8080", filename)
	assert.NoError(t, err)
	assert.NotNil(t, services)
}

func TestCreateServices_OSFileFailure(t *testing.T) {
	createError := errors.New("create error")
	tmp := t.TempDir()
	filename := filepath.Join(tmp, "auth.json")

	mockOps := new(mocks.FileInterface)
	mockOps.On("Create", mock.Anything).Return(nil, createError)

	services, err := services.CreateServices(":8080", filename, mockOps)
	assert.ErrorIs(t, err, createError)
	assert.Nil(t, services)
}
