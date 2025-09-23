package main

import (
	"errors"
	"net/http"

	"github.com/charmbracelet/log"
)

// Authentication handlers
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok {
		w.Header().Set("WWW-Authenticate", "Basic")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		log.Info("Incorrect authentication headers")
		return
	}

	session_token, err := Database.LoginUser(username, password)
	if errors.Is(err, USER_DOESNT_EXIST) {
		w.Header().Set("WWW-Authenticate", "Basic")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		log.Infof("Unknown user attempted login: user %v", username)
		return
	} else if errors.Is(err, INVALID_LOGIN) {
		w.Header().Set("WWW-Authenticate", "Basic")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		log.Infof("Incorrect login attempted: user: %v", username)
		return
	} else if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Errorf("Internal login error: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(session_token))
}
