package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
)

// New is to create new Handler
func New() http.Handler {
	mux := mux.NewRouter()
	mux.Handle("/", http.FileServer(http.Dir("public/")))

	// OauthGoogle
	mux.HandleFunc("/auth/google/login", oauthGoogleLogin)
	mux.HandleFunc("/auth/google/callback", oauthGoogleCallback)
	mux.HandleFunc("/parse", handleHTML).Methods("POST")

	return mux
}
