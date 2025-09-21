package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/charmbracelet/log"
)

func main() {
	log.SetDefault(log.NewWithOptions(os.Stdout, log.Options{
		ReportTimestamp: true,
		ReportCaller:    true,
	}))

	ParseFlags()
	fs, err := NewAtlas(Flags.root)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("root=%v", Flags.root)

	authDatabase, err := LoadAuthDatabase("./auth.json")
	if errors.Is(err, os.ErrNotExist) {
		log.Warn("database doesn't exist at ./auth.json. Creating new database")
		CreateAuthDatabase("./auth.json")
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /f/{path...}", func(w http.ResponseWriter, r *http.Request) {
		path := CurrPath(r.PathValue("path"))

		if !fs.Exists(path) {
			log.Warn("File does not exist", "path", path)
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "resource not found")
			return
		}

		http.ServeFile(w, r, path.Resolve(fs))
		log.Info("File Serve Success", "path", path)
	})

	mux.HandleFunc("HEAD /f/{path...}", func(w http.ResponseWriter, r *http.Request) {
		path := CurrPath(r.PathValue("path"))

		info, err := path.Stat(fs)
		if err != nil {
			log.Info("", "path", path, "exists", 0)
			w.Header().Add("Exists", "0")
			return
		}

		log.Info(path.path, "exists", 1)
		w.Header().Add("Exists", "1")

		if info.IsDir() {
			w.Header().Add("Type", "directory")
		} else {
			w.Header().Add("Type", "file")
			w.Header().Add("Size", fmt.Sprint(info.Size()))
		}
	})

	mux.HandleFunc("POST /f/{path...}", func(w http.ResponseWriter, r *http.Request) {
		path := CurrPath(r.PathValue("path"))

		r.ParseMultipartForm(1 << 20)
		file, handler, err := r.FormFile("f")
		if err != nil {
			log.Warn(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err)
			return
		}
		defer file.Close()

		if len(path.path) == 0 || path.path[len(path.path)-1] == '/' {
			path.path += handler.Filename
		}

		err = fs.Upload(file, path)
		if err != nil {
			log.Warn(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err)
			return
		}

		log.Info("Upload Success", "path", path, "size", handler.Size)
	})

	mux.HandleFunc("DELETE /f/{path...}", func(w http.ResponseWriter, r *http.Request) {
		path := CurrPath(r.PathValue("path"))
		if err := fs.Delete(path); err != nil {
			log.Warn(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err)
			return
		}
		log.Info("Delete Success", "path", path)
	})

	mux.HandleFunc("POST /u/login", func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			log.Info("login failed")
			w.WriteHeader(401)
			return
		}

		if !authDatabase.CheckAuth(username, password) {
			log.Warnf("invalid login %s:%s", username, password)
		}
	})

	server := http.Server{
		Addr:    Flags.address,
		Handler: mux,
	}

	log.Infof("starting on %v", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Errorf("%v", err)
	}
}
