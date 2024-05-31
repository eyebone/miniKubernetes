package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	fileServer := http.FileServer(http.Dir("./"))
	http.Handle("/", fileServer)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // 默认端口
	}

	fmt.Printf("Starting server at port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
