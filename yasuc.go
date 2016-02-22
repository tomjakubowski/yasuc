package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
)

var pastes map[string]string

const baseUrl = "http://localhost:8080"

func stashPaste(paste string) string {
	rawKey := sha256.Sum256([]byte(paste))
	key := hex.EncodeToString(rawKey[:])
	// TODO: launch nukes on collision
	pastes[key] = paste
	return key
}

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
		key := stashPaste(body)
		fmt.Fprintf(w, "%s/%s\n", baseUrl, key)
		return
	}
	fmt.Fprintf(w, "usage message goes here\n")
}

func main() {
	pastes = map[string]string{}
	http.HandleFunc("/", handleRoot)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
