package main

import (
	_ "embed"
	"encoding/json"
	"os"
	"time"

	"github.com/google/uuid"
)

//go:embed authTemplate.json
var templateDatabase string

// Keep session alive for a week of inactivity
const SESSION_LIFETIME = 604800

type InternalResponseCode int

// Internal response codes
const (
	SUCCESS InternalResponseCode = iota
	USER_EXISTS
	USER_DOESNT_EXIST
	INVALID_LOGIN
)

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
	Filename     string
	Admin        []string                  `json:"admin"`
	Users        map[string]string         `json:"users"`
	Sessions     map[string]Session        `json:"sessions"`
	Shared_files map[string]SharedFile     `json:"shared_files"`
	Permissions  map[string]UserPermission `json:"permissions"`
}

func LoadAuthDatabase(file string) (*AuthDatabase, error) {
	databaseFile, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer databaseFile.Close()

	content, err := os.ReadFile(file)

	if err != nil {
		return nil, err
	}

	var payload AuthDatabase
	err = json.Unmarshal(content, &payload)
	if err != nil {
		return nil, err
	}

	payload.Filename = file

	return &payload, nil
}

func CreateAuthDatabase(file string) (*AuthDatabase, error) {
	f, err := os.Create(file)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	_, err = f.WriteString(templateDatabase)
	if err != nil {
		return nil, err
	}

	return LoadAuthDatabase(file)
}

func (d *AuthDatabase) CheckAuth(username string, password string) bool {
	saved_password, ok := d.Users[username]
	if !ok {
		return false
	}
	return saved_password == password
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

func (d *AuthDatabase) ValidateToken(session_token string, username string) bool {
	// Returns true if the session token is valid
	is_valid := d.Sessions[session_token].Username == username
	is_alive := d.CheckSessionTime(session_token)

	return is_valid && is_alive
}

func (d *AuthDatabase) CheckSessionTime(session_token string) bool {
	current_time := int(time.Now().Unix())
	time_inactive := current_time - d.Sessions[session_token].Last_login
	return time_inactive <= SESSION_LIFETIME
}

func (d *AuthDatabase) UpdateSession(session_token string, recently_accessed bool) {
	is_alive := d.CheckSessionTime(session_token)
	if !is_alive {
		delete(d.Sessions, session_token)
		d.Save()
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
}

func (d *AuthDatabase) CreateUser(username string) InternalResponseCode {
	// Check for an existing user with the same name
	for key, _ := range d.Users {
		if key == username {
			return USER_EXISTS
		}
	}

	// No matching username found, create a new user with the username and password the same
	d.Users[username] = username

	return SUCCESS
}

func (d *AuthDatabase) RemoveUser(username string) InternalResponseCode {
	// Check the user exists
	_, ok := d.Users[username]
	if !ok {
		return USER_DOESNT_EXIST
	}
	delete(d.Users, username)
	return SUCCESS
}

func (d *AuthDatabase) LoginUser(username string, password string) (string, InternalResponseCode, error) {
	if !d.CheckAuth(username, password) {
		return "", INVALID_LOGIN, nil
	}

	session_token, err := d.GetSessionToken(username)

	return session_token, SUCCESS, err
}

func (d *AuthDatabase) Save() error {
	byteValue, err := json.Marshal(d)
	if err != nil {
		return err
	}

	return os.WriteFile(d.Filename, byteValue, 0644)
}
