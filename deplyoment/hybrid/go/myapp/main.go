package main

import (
	"net/http"
	"fmt"
	"os"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, Wordld\n")
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.HandleFunc("/", helloHandler)
	http.ListenAndServe(":"+port, nil)
}
