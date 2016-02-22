package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	baseUrl      = "http://localhost:8080"
	pastesBucket = "pastes"
)

func stashPaste(db *bolt.DB, pasteStr string) (key string, err error) {
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
	type notFound struct{}
	rawKey, err := hex.DecodeString(key)
	if err != nil {
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
		key, err := stashPaste(h.db, body)
		if err != nil {
			http.Error(w, "db update error", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "%s/%s\n", baseUrl, key)
		return
	}
	fmt.Fprintf(w, "usage message goes here\n")
}

func newHandler(db *bolt.DB) http.Handler {
	h := handler{db}
	mux := http.NewServeMux()
	mux.HandleFunc("/", h.alles)
	return mux
}

func main() {
	var dbPath string
	flag.StringVar(&dbPath, "db", "", "location of database file")
	flag.Parse()
	if dbPath == "" {
		fmt.Fprintf(os.Stderr, "db option is required\n")
		os.Exit(1)
	}
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	err = db.Update(func(tx *bolt.Tx) error {
		// make all of ze buckets
		tx.CreateBucketIfNotExists([]byte(pastesBucket))
		return nil
	})
	if err != nil {
		db.Close() // deferred functions aren't called for fatal logs :|
		log.Fatal(err)
	}

	if db == nil {
		log.Fatal("WHAT THE FUCK")
	}

	http.Handle("/", newHandler(db))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
