package main

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type claConfig struct {
	Public  uint64
	Private uint64
}

type Cla struct {
	Config            claConfig
	AuthorizedVoters  map[string]rsa.PublicKey // map of voter names to their public keys
	validationNumbers []uint64
	voterNumberMap    map[string]uint64 // map of voter names to their validation numbers
}

func NewCla(configFileName string) *Cla {
	var cla Cla

	voterFile, err := os.Open(configFileName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error opening file", configFileName, ":", err)
	}
	defer voterFile.Close()

	fileBytes, err := ioutil.ReadAll(voterFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading file", configFileName, ":", err)
	}
	if err = json.Unmarshal(fileBytes, &cla); err != nil {
		fmt.Fprintln(os.Stderr, "Error Unmarshalling data:", err)
	}

	cla.voterNumberMap = make(map[string]uint64)
	for key, _ := range cla.AuthorizedVoters {
		cla.voterNumberMap[key] = 0
	}
	cla.validationNumbers = make([]uint64, len(cla.voterNumberMap))

	return &cla
}

func listHandler(w http.ResponseWriter, r *http.Request, cla *Cla) {
	// if request comes from CTF, send full list of validation numbers
}

func registrationHandler(w http.ResponseWriter, r *http.Request, cla *Cla) {
	name := r.URL.Query().Get("name")
	sig := r.URL.Query().Get("sig")
	if name == "" || sig == "" {
		http.Error(w, "User did not specify their name, or did not sign their request.", 400)
		return
	}

	pk := cla.AuthorizedVoters[name]
	hashed := sha256.Sum256([]byte(name))
	err := rsa.VerifyPKCS1v15(&pk, crypto.SHA256, hashed[:], []byte(sig))
	if err != nil {
		http.Error(w, "User is not an authorized voter.", 403)
		return
	}

	// generate a validation number and return it
}

func main() {
	cla := NewCla("cla/config.json")
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		registrationHandler(w, r, cla)
	})
	http.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		listHandler(w, r, cla)
	})
	fmt.Println("Listening and Serving...")
	http.ListenAndServeTLS(":9889", "certs/localhost.crt", "keys/localhost.key", nil)
}
