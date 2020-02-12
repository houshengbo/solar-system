package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func getSource() string {
	source := os.Getenv("SOURCE")
	if source == "" {
		source = "Sun"
	}
	return source
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Print("Source of energy: received a request")
	source := getSource()
	log.Printf("Hello! This is the source of energy for the %s, providing heat and light.\n", source)
}

func main() {
	log.Print("Source of energy: starting...")
	source := getSource()
	log.Printf("Hello! This is the source of energy for the %s, providing heat and light.\n", source)

	http.HandleFunc("/", handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Source of energy: listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
