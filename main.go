package main

import (
	// "fmt"

	"fmt"
	"log"
	"net/http"
	"os"
	"simaster-parser/handlers"
)

func main() {
	Port := os.Getenv("PORT")
	if Port == "" {
		Port = "8000"
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":" + Port),
		Handler: handlers.New(),
	}

	log.Printf("Starting HTTP Server. Listening at %q", server.Addr)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("%v", err)
	} else {
		log.Println("Server closed!")
	}
}
