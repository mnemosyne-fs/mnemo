package authentication_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mnemosynefs/mnemo/internal"
	"github.com/mnemosynefs/mnemo/internal/authentication"
	"github.com/mnemosynefs/mnemo/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

//
// Database Testing

func TestCreateNewDatabase_CreateNew(t *testing.T) {
	tmp := t.TempDir()
	filename := filepath.Join(tmp, "./testing.json")
	database, err := authentication.CreateNewDatabase(filename)
	assert.NoError(t, err)

	target := &authentication.AuthDatabase{
		Filename:     filename,
		Admin:        []string{"admin"},
		Users:        map[string]string{"admin": "admin"},
		Sessions:     map[string]authentication.Session{},
		Shared_files: map[string]authentication.SharedFile{},
		Permissions:  map[string]authentication.UserPermission{"/": {"admin": 3}},
		// Functions cannot be compared. These are TestCreateNewDatabase_FileFunctions
	}

	diff := cmp.Diff(
		database,
		target,
		cmp.AllowUnexported(
			authentication.AuthDatabase{},
		), // This is needed because functions are not exported
		cmp.FilterPath(func(p cmp.Path) bool {
			// Ignore functions
			last := p.Last().String()
			return last == ".FileOps"
		}, cmp.Ignore()),
	)

	if diff != "" {
		t.Errorf("AuthDatabase mismatch. (-target +got):\n%s", diff)
	}
}

func TestCreateNewDatabase_Load(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "testing.json")

	content := []byte(authentication.TemplateDatabase)
	os.WriteFile(file, content, 0644)

	database, err := authentication.CreateNewDatabase(file)
	require.NoError(t, err)

	var target authentication.AuthDatabase
	err = json.Unmarshal(content, &target)
	require.NoError(t, err)

	target.Filename = file
	target.FileOps = new(authentication.FileOperations)

	diff := cmp.Diff(database, &target,
		cmp.AllowUnexported(authentication.AuthDatabase{}),
		cmp.FilterPath(func(p cmp.Path) bool {
			last := p.Last().String()
			return last == ".openFile" ||
				last == ".readFile" ||
				last == ".createFile" ||
				last == ".writeFile"
		}, cmp.Ignore()))

	if diff != "" {
		t.Errorf("AuthDatabase mismatch. (-target +got):\n%s", diff)
	}
}

func TestCreateNewDatabase_OSFileFunctions(t *testing.T) {
	tmp := t.TempDir()
	database, err := authentication.CreateNewDatabase(filepath.Join(tmp, "database.json"))
	require.NoError(t, err)

	file, err := database.FileOps.Create(filepath.Join(tmp, "test.txt"))
	require.NoError(t, err)
	require.NotNil(t, file)

	data := []byte("hello, world!")
	err = database.FileOps.Write(file.Name(), data, 0644)
	require.NoError(t, err)

	read, err := database.FileOps.Read(file.Name())
	require.NoError(t, err)
	require.Equal(t, data, read)

	_ = file.Close()

	file, err = database.FileOps.Open(filepath.Join(tmp, "test.txt"))
	require.NoError(t, err)
	require.NotNil(t, file)

	database.SetFileOperations(new(mocks.FileInterface))
}

func TestLoadAuthDatabase_OSFileFailure(t *testing.T) {
	tmp := t.TempDir()
	filename := filepath.Join(tmp, "auth1.json")
	filename2 := filepath.Join(tmp, "auth2.json")

	openError := errors.New("open error")
	readError := errors.New("read error")
	createError := errors.New("create error")

	mockOps := new(mocks.FileInterface)

	// Create fails
	mockOps.On("Create", filename).Return(nil, createError)
	database, err := authentication.CreateNewDatabase(filename, mockOps)
	assert.ErrorIs(t, err, createError)
	assert.Nil(t, database)

	// Open fails
	mockOps.ExpectedCalls = nil
	mockOps.On("Create", filename).Return(os.Create(filename2))
	mockOps.On("Open", filename).Return(nil, openError)

	database, err = authentication.CreateNewDatabase(filename, mockOps)
	assert.NotErrorIs(t, err, createError)
	assert.ErrorIs(t, err, openError)
	assert.Nil(t, database)

	// Read fails
	mockOps.ExpectedCalls = nil
	mockOps.On("Create", filename).Return(os.Create(filename2))
	mockOps.On("Open", filename).Return(os.Open(filename2))
	mockOps.On("Read", filename).Return(nil, readError)

	database, err = authentication.CreateNewDatabase(filename, mockOps)
	assert.NotErrorIs(t, err, openError)
	assert.ErrorIs(t, err, readError)
	assert.Nil(t, database)
}

func TestCreateUser_NewUser(t *testing.T) {
	tmp := t.TempDir()
	database, err := authentication.CreateNewDatabase(filepath.Join(tmp, "auth.json"))
	require.NoError(t, err)

	err = database.CreateUser("test")
	assert.NoError(t, err)
}

func TestCreateUser_ExistingUser(t *testing.T) {
	tmp := t.TempDir()
	database, err := authentication.CreateNewDatabase(filepath.Join(tmp, "auth.json"))
	require.NoError(t, err)

	err = database.CreateUser("test")
	assert.NoError(t, err)

	err = database.CreateUser("test")
	assert.ErrorIs(t, err, internal.ErrUserExists)
}

func TestCheckUserExists(t *testing.T) {
	tmp := t.TempDir()
	database, err := authentication.CreateNewDatabase(filepath.Join(tmp, "auth.json"))
	require.NoError(t, err)

	exists := database.CheckUserExists("test")
	assert.False(t, exists)

	err = database.CreateUser("test")
	assert.NoError(t, err)

	exists = database.CheckUserExists("test")
	assert.True(t, exists)
}

func TestLoginUser_UserExists(t *testing.T) {
	tmp := t.TempDir()
	database, err := authentication.CreateNewDatabase(filepath.Join(tmp, "auth.json"))
	require.Nil(t, err)

	err = database.CreateUser("test")
	assert.NoError(t, err)

	session_token, err := database.LoginUser("test", "test")
	assert.NoError(t, err)
	assert.NotEqual(t, "", session_token)
}

func TestLoginUser_UserNotExists(t *testing.T) {
	tmp := t.TempDir()
	database, err := authentication.CreateNewDatabase(filepath.Join(tmp, "auth.json"))
	require.Nil(t, err)

	session_token, err := database.LoginUser("test", "test")
	assert.ErrorIs(t, err, internal.ErrUserNotExists)
	assert.Equal(t, "", session_token)
}

func TestLoginUser_InvalidLogin(t *testing.T) {
	tmp := t.TempDir()
	database, err := authentication.CreateNewDatabase(filepath.Join(tmp, "auth.json"))
	require.Nil(t, err)

	err = database.CreateUser("test")
	assert.NoError(t, err)

	session_token, err := database.LoginUser("test", "wrong-password")
	assert.ErrorIs(t, err, internal.ErrInvalidLogin)
	assert.Equal(t, "", session_token)
}

func TestCheckAuth_InvalidAuthentication(t *testing.T) {
	tmp := t.TempDir()
	database, err := authentication.CreateNewDatabase(filepath.Join(tmp, "auth.json"))
	require.Nil(t, err)

	is_valid := database.CheckAuth("test", "test")
	assert.False(t, is_valid)
}

func TestGenerateNewSessionToken_UserNotExists(t *testing.T) {
	tmp := t.TempDir()
	database, err := authentication.CreateNewDatabase(filepath.Join(tmp, "auth.json"))
	require.NoError(t, err)

	session_token, err := database.GenerateNewSessionToken("test")
	assert.ErrorIs(t, err, internal.ErrUserNotExists)
	assert.Equal(t, "", session_token)
}

func TestGenerateNewSessionToken_SaveError(t *testing.T) {
	tmp := t.TempDir()

	database, err := authentication.CreateNewDatabase(filepath.Join(tmp, "auth.json"))
	require.Nil(t, err)

	writeError := errors.New("write error")

	mockOps := new(mocks.FileInterface)
	mockOps.On("Write", mock.Anything, mock.Anything, mock.Anything).Return(writeError)
	database.SetFileOperations(mockOps)

	err = database.CreateUser("test")
	assert.NoError(t, err)

	session_token, err := database.GenerateNewSessionToken("test")
	assert.ErrorIs(t, err, writeError)
	assert.Equal(t, "", session_token)
}

func TestGetSessionToken_UserExists(t *testing.T) {
	tmp := t.TempDir()
	filename := filepath.Join(tmp, "auth.json")

	database, err := authentication.CreateNewDatabase(filename)
	require.NoError(t, err)

	err = database.CreateUser("test")
	assert.NoError(t, err)

	session_token, err := database.LoginUser("test", "test")
	assert.NoError(t, err)
	assert.NotEqual(t, "", session_token)

	session_token, err = database.GetSessionToken("test")
	assert.NoError(t, err)
	assert.NotEqual(t, "", session_token)
}

func TestGetSessionToken_UserNotExists(t *testing.T) {
	tmp := t.TempDir()
	filename := filepath.Join(tmp, "auth.json")

	database, err := authentication.CreateNewDatabase(filename)
	require.NoError(t, err)

	session_token, err := database.GetSessionToken("test")
	assert.ErrorIs(t, err, internal.ErrUserNotExists)
	assert.Equal(t, "", session_token)
}

func TestValidateToken_UserExists(t *testing.T) {
	tmp := t.TempDir()
	filename := filepath.Join(tmp, "auth.json")

	database, err := authentication.CreateNewDatabase(filename)
	require.NoError(t, err)

	err = database.CreateUser("test")
	assert.NoError(t, err)
	session_token, err := database.LoginUser("test", "test")
	assert.NoError(t, err)
	assert.NotEqual(t, "", session_token)

	is_valid := database.ValidateToken(session_token)
	assert.True(t, is_valid)
}

func TestUpdateSession_UserExists(t *testing.T) {
	tmp := t.TempDir()
	filename := filepath.Join(tmp, "auth.json")

	database, err := authentication.CreateNewDatabase(filename)
	require.NoError(t, err)

	err = database.CreateUser("test")
	assert.NoError(t, err)
	session_token, err := database.LoginUser("test", "test")
	assert.NoError(t, err)
	assert.NotEqual(t, "", session_token)

	// Recently accessed
	err = database.UpdateSession(session_token, true)
	assert.NoError(t, err)

	// Not recently accessed
	err = database.UpdateSession(session_token, false)
	assert.NoError(t, err)
}

func TestUpdateSession_ExpiredSession(t *testing.T) {
	tmp := t.TempDir()
	filename := filepath.Join(tmp, "auth.json")

	database, err := authentication.CreateNewDatabase(filename)
	require.NoError(t, err)

	err = database.CreateUser("test")
	assert.NoError(t, err)
	session_token, err := database.LoginUser("test", "test")
	assert.NoError(t, err)

	newSession := authentication.Session{
		Last_login: 0,
		Username:   "test",
	}
	database.Sessions[session_token] = newSession

	err = database.UpdateSession(session_token, false)
	assert.ErrorIs(t, err, internal.ErrInvalidSession)
}

func TestGetUserFromToken_ValidToken(t *testing.T) {
	tmp := t.TempDir()
	filename := filepath.Join(tmp, "auth.json")

	database, err := authentication.CreateNewDatabase(filename)
	require.NoError(t, err)

	err = database.CreateUser("test")
	assert.NoError(t, err)
	session_token, err := database.LoginUser("test", "test")
	assert.NoError(t, err)

	username, err := database.GetUserFromToken(session_token)
	assert.NoError(t, err)
	assert.Equal(t, "test", username)
}

func TestGetUserFromToken_InvalidToken(t *testing.T) {
	tmp := t.TempDir()
	filename := filepath.Join(tmp, "auth.json")

	database, err := authentication.CreateNewDatabase(filename)
	require.NoError(t, err)

	username, err := database.GetUserFromToken("")
	assert.ErrorIs(t, err, internal.ErrInvalidSession)
	assert.Equal(t, "", username)
}

func TestRemoveUser_UserExists(t *testing.T) {
	tmp := t.TempDir()
	database, err := authentication.CreateNewDatabase(filepath.Join(tmp, "auth.json"))
	require.NoError(t, err)
	err = database.CreateUser("test")
	assert.NoError(t, err)
	_, err = database.LoginUser("test", "test")
	assert.NoError(t, err)

	err = database.RemoveUser("test")
	assert.NoError(t, err)

	exists := database.CheckUserExists("test")
	assert.False(t, exists)
}

func TestRemoveUser_UserNotExists(t *testing.T) {
	tmp := t.TempDir()
	database, err := authentication.CreateNewDatabase(filepath.Join(tmp, "auth.json"))
	require.NoError(t, err)

	err = database.RemoveUser("test")
	assert.ErrorIs(t, err, internal.ErrUserNotExists)
}

//
// HTTP Handler/Middleware Testing
//

func TestSessionMiddlewareHandler_ValidToken(t *testing.T) {
	tmp := t.TempDir()
	filename := filepath.Join(tmp, "auth.json")

	database, err := authentication.CreateNewDatabase(filename)
	require.NoError(t, err)

	admin_token, err := database.LoginUser("admin", "admin")
	assert.NoError(t, err)

	dummy_handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "admin", r.Header.Get("username"))
		w.WriteHeader(http.StatusOK)
	})

	middleware := database.SessionMiddlewareHandler(dummy_handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("session_token", admin_token)

	rec := httptest.NewRecorder()
	middleware.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "success", rec.Body.String())
}

func TestSessionMiddlewareHandler_InvalidTokne(t *testing.T) {
	tmp := t.TempDir()
	filename := filepath.Join(tmp, "auth.json")

	database, err := authentication.CreateNewDatabase(filename)
	assert.NoError(t, err)

	dummy_handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("this should have not been called at all")
	})

	middleware := database.SessionMiddlewareHandler(dummy_handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("session_token", "badtoken")

	rec := httptest.NewRecorder()
	middleware.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid token")
}

func TestSessionMiddlewareHandler_NoToken(t *testing.T) {
	tmp := t.TempDir()
	filename := filepath.Join(tmp, "auth.json")

	database, err := authentication.CreateNewDatabase(filename)
	assert.NoError(t, err)

	dummy_handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "", r.Header.Get("username"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("guest login"))
	})

	middleware := database.SessionMiddlewareHandler(dummy_handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("session_token", "")

	rec := httptest.NewRecorder()
	middleware.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "guest login")
}

func TestLoginHandler_MissingAuthHeader(t *testing.T) {
	tmp := t.TempDir()
	filename := filepath.Join(tmp, "auth.json")

	database, err := authentication.CreateNewDatabase(filename)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(database.LoginHandler)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "Unauthorized")
	assert.Equal(t, "Basic", rec.Header().Get("WWW-Authenticate"))
}

func TestLoginHandler_UserNotExists(t *testing.T) {
	tmp := t.TempDir()
	filename := filepath.Join(tmp, "auth.json")

	database, err := authentication.CreateNewDatabase(filename)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.SetBasicAuth("ghost", "wrongpassword")
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(database.LoginHandler)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "Unauthorized")
	require.Equal(t, "Basic", rec.Header().Get("WWW-Authenticate"))
}

func TestLoginHandler_InvalidCredentials(t *testing.T) {
	tmp := t.TempDir()
	filename := filepath.Join(tmp, "auth.json")

	database, err := authentication.CreateNewDatabase(filename)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.SetBasicAuth("admin", "wrongpassword")
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(database.LoginHandler)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "Unauthorized")
	assert.Equal(t, "Basic", rec.Header().Get("WWW-Authenticate"))
}

func TestLoginHandler_ValidLogin(t *testing.T) {
	tmp := t.TempDir()
	filename := filepath.Join(tmp, "auth.json")

	database, err := authentication.CreateNewDatabase(filename)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.SetBasicAuth("admin", "admin")
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(database.LoginHandler)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
