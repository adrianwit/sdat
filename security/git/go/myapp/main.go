package main

import (
	"net/http"
	"fmt"
	"os"
	"github.com/adranwit/secret"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	message := secret.New()
	fmt.Fprintf(w, fmt.Sprintf("Hello, %v\n", message))
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.HandleFunc("/", helloHandler)
	http.ListenAndServe(":"+port, nil)
}
