package main

import (
	"log"
	"net/http"
)

func main() {
	port := ":8080"

	fileServer := http.FileServer(http.Dir("."))
	log.Printf("Server started at http://localhost%s\n", port)

	err := http.ListenAndServe(port, fileServer)
	if err != nil {
		log.Fatal("Err: ", err)
	}
}
