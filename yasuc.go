// yasuc - yet another sprunge.us clone (command line pastebin)
//
// Copyright (C) Tom Jakubowski <tom@crystae.net>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"github.com/boltdb/bolt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	pastesBucket = "pastes"
	maxPasteSize = 4 * 1024 * 1024
)

const usageText = `
<!doctype html>
<html>
  <head>
  </head>
  <body>
    <pre>
yasuc(1)                          YASUC                          yasuc(1)

NAME
    yasuc - command line pastebin.

SYNOPSIS
    &lt;command&gt; | curl -F 'sprunge=&lt;-' {{.BaseURL}}

DESCRIPTION
    A command line pastebin.  Pastes are immutable and created with simple HTTP
    POST requests. The path of a paste's URL is the SHA-256 digest of the
    paste's contents.

EXAMPLE
    $ echo 'hello world' | curl -F 'sprunge=&lt;-' {{.BaseURL}}
       {{.BaseURL}}/a948904f2f0f479b8f8197694b30184b0d2ed1c1cd2a1ec0fb85d299a192a447
    $ firefox {{.BaseURL}}/a948904f2f0f479b8f8197694b30184b0d2ed1c1cd2a1ec0fb85d299a192a447

COPYRIGHT
    Copyright © Tom Jakubowski.  License AGPLv3: GNU Affero GPL Version 3
    <a href="https://www.gnu.org/licenses/agpl-3.0.en.html">https://www.gnu.org/licenses/agpl-3.0.en.html</a>.

    This is free software; you are free to change and redistribute it.  There
    is NO WARRANTY, to the extend permitted by law.  For copying conditions and
    source, see <a href="https://github.com/tomjakubowski/yasuc">https://github.com/tomjakubowski/yasuc</a>.

SEE ALSO
    <a href="https://github.com/tomjakubowski/yasuc">http://github.com/tomjakubowski/yasuc</a>
    </pre>
  </body>
</html>
`

type pasteTooLarge struct{}

func (e pasteTooLarge) Error() string {
	return fmt.Sprintf("paste too large (maximum size %d bytes)", maxPasteSize)
}

func stashPaste(db *bolt.DB, pasteStr string) (key string, err error) {
	if len(pasteStr) > maxPasteSize {
		err = pasteTooLarge{}
		return
	}
	paste := []byte(pasteStr)
	rawKey := sha256.Sum256(paste)
	// TODO: launch nukes on collision
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(pastesBucket))
		if b == nil {
			return errors.New("bucket not found??")
		}
		err := b.Put(rawKey[:], paste)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return
	}
	key = hex.EncodeToString(rawKey[:])
	return
}

type pasteNotFound struct{}

func (e pasteNotFound) Error() string {
	return "not found"
}

func fetchPaste(db *bolt.DB, key string) (paste string, err error) {
	rawKey, err := hex.DecodeString(key)
	if err != nil { // bad URL
		err = pasteNotFound{}
		return
	}
	var rawPaste []byte
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(pastesBucket))
		if b == nil {
			return errors.New("bucket not found???")
		}
		rawPaste = b.Get(rawKey)
		paste = string(rawPaste)
		return nil
	})
	if err != nil {
		return
	}
	if rawPaste == nil {
		err = pasteNotFound{}
		return
	}
	return
}

type handler struct {
	db *bolt.DB
}

func (h *handler) alles(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		paste, err := fetchPaste(h.db, req.URL.Path[1:])
		if err != nil {
			if _, ok := err.(pasteNotFound); ok {
				http.Error(w, "not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		fmt.Fprintf(w, "%s", paste)
		return
	}

	if req.Method == http.MethodPost {
		body := req.FormValue("sprunge")
		// sprunge.us proceeds if there's no form data in 'sprunge', and just makes
		// an empty paste.
		key, err := stashPaste(h.db, body)
		if err != nil {
			if _, ok := err.(pasteTooLarge); ok {
				http.Error(w, err.Error(), http.StatusNotAcceptable)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		baseURL := fmt.Sprintf("http://%s", req.Host)
		fmt.Fprintf(w, "%s/%s\n", baseURL, key)
		return
	}
	// usage message
	tmpl, err := template.New("usage").Parse(usageText)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	baseURL := fmt.Sprintf("http://%s", req.Host)
	data := struct{ BaseURL string }{baseURL}
	_ = tmpl.Execute(w, data)
}

func newHandler(db *bolt.DB) http.Handler {
	h := handler{db}
	mux := http.NewServeMux()
	mux.HandleFunc("/", h.alles)
	return mux
}

func main() {
	var dbPath, addr string
	var port int
	flag.StringVar(&dbPath, "db", "", "location of database file")
	flag.StringVar(&addr, "addr", "", "bind address")
	flag.IntVar(&port, "port", 9001, "bind port")
	flag.Parse()
	if dbPath == "" {
		fmt.Fprintf(os.Stderr, "db option is required\n")
		os.Exit(1)
	}
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	err = db.Update(func(tx *bolt.Tx) error {
		// make all of ze buckets
		_, err := tx.CreateBucketIfNotExists([]byte(pastesBucket))
		if err != nil {
			_ = db.Close()
			log.Fatal(err)
		}
		return nil
	})
	if err != nil {
		_ = db.Close() // deferred functions aren't called for fatal logs :|
		log.Fatal(err)
	}

	http.Handle("/", newHandler(db))
	sockAddr := fmt.Sprintf("%s:%d", addr, port)
	log.Fatal(http.ListenAndServe(sockAddr, nil))
}
