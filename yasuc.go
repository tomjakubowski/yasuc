package main

import (
	"fmt"
	"log"
	"net/http"
)

var (
	pastes map[string]string
	count  int
)

const baseUrl = "http://localhost:8080"

func handleRoot(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		paste, ok := pastes[req.URL.Path[1:]]
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		fmt.Fprintf(w, "%s", paste)
		return
	}

	if req.Method == http.MethodPost {
		body := req.FormValue("sprunge")
		// sprunge.us proceeds if there's no form data in 'sprunge', and just makes
		// an empty paste.
		pastes[fmt.Sprintf("%d", count)] = body
		fmt.Fprintf(w, "%s/%d\n", baseUrl, count)
		count++
		return
	}
	fmt.Fprintf(w, "usage message goes here\n")
}

func main() {
	pastes = map[string]string{}
	http.HandleFunc("/", handleRoot)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
