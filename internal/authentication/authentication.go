package authentication

import (
	_ "embed"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/mnemosynefs/mnemo/internal"
)

//go:generate ifacemaker -f=authentication.go -s=AuthDatabase -i=Database -p=authentication -o=authentication_interface.go
//go:generate mockery --name=Database --dir=. --output=../../mocks --outpkg=mocks --filename=database.go

//go:embed authTemplate.json
var TemplateDatabase string

// Keep session alive for a week of inactivity

type InternalResponseCode int

type Session struct {
	Username   string
	Last_login int
}

type SharedFile struct {
	Files       []string
	Accesses    int
	Time_shared int
	Lifetime    int
}

type UserPermission map[string]int

type AuthDatabase struct {
	Filename     string                    `json:"-"`
	Admin        []string                  `json:"admin"`
	Users        map[string]string         `json:"users"`
	Sessions     map[string]Session        `json:"sessions"`
	Shared_files map[string]SharedFile     `json:"shared_files"`
	Permissions  map[string]UserPermission `json:"permissions"`

	FileOps FileInterface `json:"-"`
}

func (d *AuthDatabase) SetFileOperations(newOps FileInterface) {
	d.FileOps = newOps
}

func CreateNewDatabase(file string, fileOps ...FileInterface) (*AuthDatabase, error) {
	database := new(AuthDatabase)
	if len(fileOps) > 0 {
		database.FileOps = fileOps[0]
	} else {
		database.FileOps = new(FileOperations)
	}

	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		log.Warnf("Database doesn't exist at location %v. Creating new database.", file)
		return database.CreateAuthDatabase(file)
	}

	return database.LoadAuthDatabase(file)
}

func (d *AuthDatabase) LoadAuthDatabase(file string) (*AuthDatabase, error) {
	databaseFile, err := d.FileOps.Open(file)
	if err != nil {
		return nil, err
	}
	defer databaseFile.Close()

	content, err := d.FileOps.Read(file)

	if err != nil {
		return nil, err
	}

	var payload AuthDatabase
	err = json.Unmarshal(content, &payload)
	if err != nil {
		return nil, err
	}

	payload.Filename = file
	payload.FileOps = d.FileOps

	return &payload, nil
}

func (d *AuthDatabase) CreateAuthDatabase(file string) (*AuthDatabase, error) {
	f, err := d.FileOps.Create(file)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	_, err = f.WriteString(TemplateDatabase)
	if err != nil {
		return nil, err
	}

	return d.LoadAuthDatabase(file)
}

func (d *AuthDatabase) CheckAuth(username string, password string) bool {
	saved_password, ok := d.Users[username]
	if !ok {
		return false
	}
	return saved_password == password
}

func (d *AuthDatabase) CheckUserExists(username string) bool {
	_, ok := d.Users[username]
	return ok
}

func (d *AuthDatabase) GenerateNewSessionToken(username string) (string, error) {
	if !d.CheckUserExists(username) {
		return "", internal.ErrUserNotExists
	}
	// No session token found, need to create one
	new_token := d.CreateSessionToken(username)
	new_session := Session{
		Username:   username,
		Last_login: int(time.Now().Unix()),
	}
	d.Sessions[new_token] = new_session

	// Save the new session to the database
	err := d.Save()
	if err != nil {
		return "", err
	}

	return new_token, nil
}

func (d *AuthDatabase) CreateSessionToken(username string) string {
	var id string
	for {
		id = uuid.NewString()
		if _, exists := d.Sessions[id]; !exists {
			break
		}
	}
	return id
}

func (d *AuthDatabase) GetSessionToken(username string) (string, error) {
	// This function assumes that the username exists. If it does not exist, it will create a false
	// session token that could be exploited
	for key, value := range d.Sessions {
		if value.Username == username {
			return key, nil
		}
	}

	return "", internal.ErrUserNotExists
}

func (d *AuthDatabase) ValidateToken(session_token string) bool {
	_, is_valid := d.Sessions[session_token]

	return is_valid && d.CheckSessionTime(session_token)
}

func (d *AuthDatabase) CheckSessionTime(session_token string) bool {
	current_time := int(time.Now().Unix())
	time_inactive := current_time - d.Sessions[session_token].Last_login
	return time_inactive <= internal.SESSION_LIFETIME
}

func (d *AuthDatabase) UpdateSession(session_token string, recently_accessed bool) error {
	is_alive := d.CheckSessionTime(session_token)
	if !is_alive {
		delete(d.Sessions, session_token)
		d.Save()
		return internal.ErrInvalidSession
	}

	if recently_accessed {
		username := d.Sessions[session_token].Username
		last_login := int(time.Now().Unix())
		new_session := Session{
			username,
			last_login,
		}
		d.Sessions[session_token] = new_session
		d.Save()
	}

	return nil
}

func (d *AuthDatabase) CreateUser(username string) error {
	// Check for an existing user with the same name
	for key := range d.Users {
		if key == username {
			return internal.ErrUserExists
		}
	}

	// No matching username found, create a new user with the username and password the same
	d.Users[username] = username

	return nil
}

func (d *AuthDatabase) RemoveUser(username string) error {
	// Check the user exists
	_, ok := d.Users[username]
	if !ok {
		return internal.ErrUserNotExists
	}
	delete(d.Users, username)

	for session_token, value := range d.Sessions {
		if value.Username == username {
			delete(d.Sessions, session_token)
			break
		}
	}

	return nil
}

func (d *AuthDatabase) LoginUser(username string, password string) (string, error) {
	if _, exists := d.Users[username]; !exists {
		return "", internal.ErrUserNotExists
	}
	if !d.CheckAuth(username, password) {
		return "", internal.ErrInvalidLogin
	}

	session_token, err := d.GenerateNewSessionToken(username)

	return session_token, err
}

func (d *AuthDatabase) GetUserFromToken(session_token string) (string, error) {
	is_valid := d.ValidateToken(session_token)
	if !is_valid {
		return "", internal.ErrInvalidSession
	}

	username := d.Sessions[session_token].Username

	return username, nil
}

func (d *AuthDatabase) Save() error {
	byteValue, err := json.Marshal(d)
	if err != nil {
		return err
	}

	return d.FileOps.Write(d.Filename, byteValue, 0644)
}
