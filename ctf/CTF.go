package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type ctfConfig struct {
	ClaSecret string
	CtfSecret string
}

type candidate struct {
	name      string
	voteCount []int
	voterIDs  []string // list of validation numbers who have voted for that particular candidate
}

type vote struct {
	CandidateName string
}

type listRequest struct {
	SharedSecret string
	ValidationNums []string
}

type Ctf struct {
	Config            ctfConfig
	candidates        []candidate
	ValidationNums map[string]bool // validation values will be true if used, false otherwise
	CandidateNames    []string        // required to unmarshal json into candidates
}

func NewCtf(configFileName string) (*Ctf, error) {
	var ctf Ctf

	candidateFile, err := os.Open(configFileName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error opening file", configFileName, ":", err)
		return nil, err
	}
	defer candidateFile.Close()

	fileBytes, err := ioutil.ReadAll(candidateFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading file", configFileName, ":", err)
		return nil, err
	}
	if err = json.Unmarshal(fileBytes, &ctf); err != nil {
		fmt.Fprintln(os.Stderr, "Error Unmarshalling data:", err)
		return nil, err
	}
	// set candidate slice to len of the candidate name list
	ctf.candidates = make([]candidate, len(ctf.CandidateNames), cap(ctf.CandidateNames))
	mapCandidateNames(ctf)

	return &ctf, nil
}

func mapCandidateNames(ctf Ctf) {
	for key, val := range ctf.CandidateNames {
		// fmt.Println("Key: ", key, "\tVal: ", val)
		ctf.candidates[key].name = val
	}
	ctf.CandidateNames = nil
	return
}

// grabs validation numbers from /list
func listHandler(w http.ResponseWriter, r *http.Request, ctf *Ctf) {
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
	for _, validationNum := range args.ValidationNums {
		ctf.ValidationNums[validationNum] = true
	}
}

func votingHandler(w http.ResponseWriter, r *http.Request, ctf *Ctf) {
	// check if voting allowed by requesting validation numbers if you don't have them yet
	// get voterID from request body
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

	// will throw error either when voter ID has been used or doesn't exist
	if ctf.validationNumbers[voterID] == true {
		http.Error(w, "Voter ID not valid", http.StatusForbidden)
		return
	}
}

func sendOutcome(ctf Ctf) {
	// http POST
}

func main() {

	ctf, err := NewCtf("ctf/config.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: ", err)
		return
	}
	http.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		listHandler(w, r, ctf)
	}

	http.HandleFunc("/vote", func(w http.ResponseWriter, r *http.Request) {
		votingHandler(w, r, ctf)
	})
	http.ListenAndServeTLS(":9999", "certs/localhost.crt", "keys/localhost.key", nil)
}
