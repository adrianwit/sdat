package main

import (
	"net/http"
	"fmt"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, Wordld\n")
}

func main() {
	http.HandleFunc("/", helloHandler)
	http.ListenAndServe(":8081", nil)
}
