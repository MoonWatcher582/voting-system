package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

type claConfig struct {
	ClaSecret string
	CtfSecret string
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

type listRequest struct {
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

	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)
	cla.generator = rand.NewZipf(r, 1.14, 2402.72, 5000)

	return &cla, nil
}

func listHandler(w http.ResponseWriter, r *http.Request, cla *Cla) {
	var args listRequest

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error unpacking user args.", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error unmarshalling json!", err)
		return
	}

	if args.SharedSecret == "" || cla.Config.CtfSecret != args.SharedSecret {
		http.Error(w, "Sent shared secret does not belong to the CTF.", http.StatusForbidden)
		return
	}

	if resp, err := votingDone(cla); err == nil {
		err := sendToCtf(w, resp, cla)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to send to CTF")
			return
		}
	} else {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("-1"))
	}
}

func sendToCtf(w http.ResponseWriter, toSend []string, cla *Cla) error {
	w.WriteHeader(http.StatusOK)

	resp := make(map[string]interface{})
	resp["sharedSecret"] = cla.Config.ClaSecret
	resp["validationNums"] = toSend

	respBytes, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Could not marshal validation numbers!", http.StatusNotFound)
		return err
	}
	w.Write(respBytes)
	return nil
}

func registrationHandler(w http.ResponseWriter, r *http.Request, cla *Cla) {
	var args registration

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error unpacking user args.", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error unmarshalling json!", err)
		return
	}

	if args.Name == "" || args.SharedSecret == "" {
		http.Error(w, "User did not send either their name or shared secret their request.", http.StatusBadRequest)
		return
	}

	storedSecret := cla.AuthorizedVoters[args.Name]
	if storedSecret != args.SharedSecret {
		http.Error(w, "User is not an authorized voter.", http.StatusForbidden)
		return
	}

	if _, err := votingDone(cla); err != nil {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("-1"))
	}

	if cla.voterNumberMap[args.Name] != "" {
		http.Error(w, "User has already registered.", http.StatusBadRequest)
		return
	}

	val := generateRandom(cla)
	fmt.Println("Using: ", val)
	cla.voterNumberMap[args.Name] = val

	resp := val
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(resp))
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
	fmt.Println("generated: ", v)
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

func votingDone(cla *Cla) ([]string, error) {
	validationNums := make([]string, len(cla.voterNumberMap))
	i := 0
	for _, b := range cla.voterNumberMap {
		if b == "" {
			return nil, errors.New("Voting is not done.")
		}
		validationNums[i] = b
		i += 1
	}
	return validationNums, nil
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
