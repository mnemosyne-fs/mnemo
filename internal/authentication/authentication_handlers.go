package authentication

import (
	"errors"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/mnemosynefs/mnemo/internal"
)

func (d *AuthDatabase) SessionMiddlewareHandler(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		session_token := r.Header.Get("session_token")

		if session_token == "" {
			r.Header.Set("username", "")
			next.ServeHTTP(w, r)
			log.Info("Admitting guest account")
			return
		}

		username, err := d.GetUserFromToken(session_token)

		if errors.Is(err, internal.ErrInvalidSession) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("invalid token"))
			log.Infof("Login attempt by invalid token: %v", session_token)
			return
		}

		r.Header.Set("username", username)
		w.Write([]byte("success"))

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(f)
}

func (d *AuthDatabase) LoginHandler(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok {
		w.Header().Set("WWW-Authenticate", "Basic")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		log.Info("Incorrect authentication headers")
		return
	}

	session_token, err := d.LoginUser(username, password)
	if errors.Is(err, internal.ErrUserNotExists) {
		w.Header().Set("WWW-Authenticate", "Basic")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		log.Infof("Unknown user attempted login: user %v", username)
		return
	} else if errors.Is(err, internal.ErrInvalidLogin) {
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
