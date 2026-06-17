// Package main
package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	port := ":8080"

	fileServer := http.FileServer(http.Dir("."))
	log.Printf("Server started at http://localhost%s\n", port)

	srv := &http.Server{
		Addr:              port,
		Handler:           fileServer,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal("Err: ", err)
	}
}
