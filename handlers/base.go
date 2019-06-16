package handlers

import (
	"encoding/gob"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

var (
	key   = []byte("secretwoy")
	store = sessions.NewFilesystemStore("", key)
)

// New is to create new Handler
func New() http.Handler {
	gob.Register(&MK{})
	gob.Register(&[]MK{})
	mux := mux.NewRouter()
	mux.Handle("/", http.FileServer(http.Dir("public/")))

	// OauthGoogle
	mux.HandleFunc("/auth/google/login", oauthGoogleLogin)
	mux.HandleFunc("/auth/google/callback", oauthGoogleCallback)
	mux.HandleFunc("/parse", handleHTML).Methods("POST")

	return mux
}
