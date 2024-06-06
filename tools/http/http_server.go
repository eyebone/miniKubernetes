package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// Data represents the structure of the data that will be received in POST requests.
type Data struct {
	Message string `json:"message"`
}

// handleGet handles GET requests.
func handleGet(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Received a GET request\n")
}

// handlePost handles POST requests.
func handlePost(w http.ResponseWriter, r *http.Request) {
	var data Data
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &data); err != nil {
		http.Error(w, "Unable to parse JSON", http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "Received a POST request with message: %s\n", data.Message)
}

func server() {
	http.HandleFunc("/get", handleGet)
	http.HandleFunc("/post", handlePost)

	port := "8082"
	fmt.Printf("Starting server at port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
