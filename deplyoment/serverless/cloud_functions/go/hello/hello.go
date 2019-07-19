package hello

import (
	"net/http"
	"fmt"
)

// HelloWorld prints "Hello, World!"
func HelloWorldFn(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintf(w, "Hello, World!")
}
