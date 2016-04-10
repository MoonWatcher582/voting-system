package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
)

type claConfig struct {
	ClaSecret uint64
	CtfSecret uint64
}

type Cla struct {
	Config           claConfig
	AuthorizedVoters map[string]string // map of voter names to their shared secret with the CLA
	voterNumberMap   map[string]string // map of voter names to their validation numbers
	generator        *rand.Zipf
}

type registration struct {
	Name         string
	SharedSecret string
}

func NewCla(configFileName string) (*Cla, error) {
	var cla Cla

	voterFile, err := os.Open(configFileName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error opening file", configFileName, ":", err)
		return nil, err
	}
	defer voterFile.Close()

	fileBytes, err := ioutil.ReadAll(voterFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading file", configFileName, ":", err)
		return nil, err
	}
	if err = json.Unmarshal(fileBytes, &cla); err != nil {
		fmt.Fprintln(os.Stderr, "Error Unmarshalling data:", err)
		return nil, err
	}

	cla.voterNumberMap = make(map[string]string)
	for key, _ := range cla.AuthorizedVoters {
		cla.voterNumberMap[key] = ""
	}

	s := rand.NewSource(1011001)
	r := rand.New(s)
	cla.generator = rand.NewZipf(r, 145150.7518525715, 145150.7518525715, 0xFFFFFF)

	return &cla, nil
}

func listHandler(w http.ResponseWriter, r *http.Request, cla *Cla) {
	// if request comes from CTF, send full list of validation numbers
}

func generateRandom(cla *Cla) string {
	newNum := false
	var v uint64
	for !newNum {
		v = cla.generator.Uint64()
		if presentInMap(cla, v) {
			newNum = false
		} else {
			newNum = true
		}
	}
	return strconv.FormatUint(v, 10)
}

func presentInMap(cla *Cla, v uint64) bool {
	for _, b := range cla.voterNumberMap {
		n, err := strconv.ParseUint(b, 10, 64)
		if err != nil {
			return false
		}
		if n == v {
			return true
		}
	}
	return false
}

func registrationHandler(w http.ResponseWriter, r *http.Request, cla *Cla) {
	var args registration

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error unpacking user args.", 400)
	}

	err = json.Unmarshal(body, &args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error unmarshalling json!", err)
	}

	if args.Name == "" || args.SharedSecret == "" {
		http.Error(w, "User did not send either their name or shared secret their request.", 400)
		return
	}

	storedSecret := cla.AuthorizedVoters[args.Name]
	if storedSecret != args.SharedSecret {
		http.Error(w, "User is not an authorized voter.", 403)
		return
	}

	if cla.voterNumberMap[args.Name] != "" {
		http.Error(w, "User has already registered.", 400)
		return
	}

	val := generateRandom(cla)
	cla.voterNumberMap[args.Name] = val

	resp := val
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(resp))
}

func main() {
	cla, err := NewCla("cla/config.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: ", err)
		return
	}

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		registrationHandler(w, r, cla)
	})
	http.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		listHandler(w, r, cla)
	})
	fmt.Println("Listening and Serving...")
	http.ListenAndServeTLS(":9889", "certs/localhost.crt", "keys/localhost.key", nil)
}
