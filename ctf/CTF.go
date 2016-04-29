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
	name		string
	voteCount	[]int
	voterIDs 	[]string
}

type Ctf struct {
	Config					ctfConfig
	candidates 				[]candidate
	validationNumbers 		map[string]bool // validation values will be true if used, false otherwise
	CandidateNames 			[]string // required to unmarshal json into
	voters					[]string
}

// create a new
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
	// set candidate slice to len of
	ctf.candidates = make([]candidate, len(ctf.CandidateNames), cap(ctf.CandidateNames))
	fmt.Println(ctf.CandidateNames)
	// Get validationNumbers from CLA
	mapCandidateNames(ctf)
	return &ctf, nil
}

func mapCandidateNames(ctf Ctf) {
	for key, val := range ctf.CandidateNames {
		// fmt.Println("Key: ", key, "\tVal: ", val)
		ctf.candidates[key].name = val
	}
	// I should destroy ctf.CandidateNames but I'm not sure how
}

func votingHandler(w http.ResponseWriter, r *http.Request, ctf *Ctf) {

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
	http.HandleFunc("/vote", func(w http.ResponseWriter, r *http.Request) {
		votingHandler(w, r, ctf)
	})
	http.ListenAndServeTLS(":9999", "certs/localhost.crt", "keys/localhost.key", nil)
}
