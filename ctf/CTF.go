package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
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
	Name      string
	VoteCount int
	VoterIDs  []string // list of validation numbers who have voted for that particular candidate
}

type ListRequest struct {
	SharedSecret string
}

type ListResult struct {
	SharedSecret   string
	ValidationNums []string
}

type PublishResp struct {
	Candidates []*candidate
}

type voteRequest struct {
	Candidate     string
	ValidationNum string
}

type Ctf struct {
	Config         ctfConfig
	candidates     map[string]*candidate // Map of candidate names to candidates
	ValidationNums map[string]bool       // validation values will be true if used, false otherwise
	CandidateNames []string              // required to unmarshal json into candidates
}

var client *http.Client

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
	ctf.candidates = map[string]*candidate{}
	ctf.ValidationNums = map[string]bool{}
	mapCandidateNames(ctf)

	return &ctf, nil
}

func mapCandidateNames(ctf Ctf) Ctf {
	for _, val := range ctf.CandidateNames {
		// populate map by giving candidate name a non-zero value
		ctf.candidates[val] = &candidate{Name: val}
	}
	ctf.CandidateNames = nil
	return ctf
}

func publishHandler(w http.ResponseWriter, r *http.Request, ctf *Ctf) {
	// Confirm that voting has ended
	if len(ctf.ValidationNums) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Voting still ongoing"))
		return
	}

	for _, val := range ctf.ValidationNums {
		if val == false {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Voting still ongoing"))
			return
		}
	}

	// Create JSON response of all candidate objects
	var publishResp PublishResp
	for _, val := range ctf.candidates {
		publishResp.Candidates = append(publishResp.Candidates, val)
	}

	resp, err := json.Marshal(publishResp)
	if err != nil {
		http.Error(w, "Could not marshal candidates!", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func getList(ctf *Ctf) error {
	args := ListRequest{
		SharedSecret: ctf.Config.CtfSecret,
	}
	b, err := json.Marshal(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "CTF listHandler: Failed to marshal arguments:", err)
		return errors.New("Could not marshal args")
	}

	buf := bytes.NewBuffer(b)
	resp, err := client.Post("https://localhost:9889/list", "application/json", buf)
	if err != nil {
		fmt.Fprintln(os.Stderr, "CTF listHandler: Post failed:", err)
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("Voting not done")
	}

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintln(os.Stderr, "CTF listHandler: Error opening response:", err)
		return errors.New("Could not open response")
	}
	defer resp.Body.Close()

	var listResult ListResult
	err = json.Unmarshal(contents, &listResult)
	if err != nil {
		fmt.Fprintln(os.Stderr, "CTF listHandler: Error unmarshalling json:", err)
	}

	if listResult.SharedSecret == "" || ctf.Config.ClaSecret != listResult.SharedSecret {
		return errors.New("CLA did not send shared secret")
	}

	for _, validationNum := range listResult.ValidationNums {
		ctf.ValidationNums[validationNum] = false
	}

	return nil
}

func votingHandler(w http.ResponseWriter, r *http.Request, ctf *Ctf) {
	var args voteRequest

	if len(ctf.ValidationNums) == 0 {
		getList(ctf)
	}

	// check if voting allowed by requesting validation numbers if you don't have them yet
	// get voterID from request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error unpacking user args.", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &args)
	if err != nil {
		http.Error(w, "Failed to unmarshal arguments", http.StatusBadRequest)
		fmt.Fprintln(os.Stderr, "Error unmarshalling json!", err)
		return
	}

	// will throw error either when voter ID has been used or doesn't exist
	if v, ok := ctf.ValidationNums[args.ValidationNum]; v == true || ok == false {
		http.Error(w, "Validation number is invalid", http.StatusForbidden)
		fmt.Fprintln(os.Stderr, "Validation number", args.ValidationNum)
		return
	}
	// test if submitted candidate exists in candidate list
	if _, ok := ctf.candidates[args.Candidate]; ok == false {
		http.Error(w, "Candidate does not exist in our directory. Please try again.", http.StatusBadRequest)
		return
	}
	// logic for voting
	ctf.ValidationNums[args.ValidationNum] = true
	targetCandidate := ctf.candidates[args.Candidate]
	targetCandidate.VoteCount++
	targetCandidate.VoterIDs = append(targetCandidate.VoterIDs, args.ValidationNum)
}

func main() {
	certFile, err := os.Open("certs/ca.crt")
	if err != nil {
		fmt.Fprintln(os.Stderr, "CTF main: Error opening cert:", err)
		return
	}
	defer certFile.Close()

	cert, err := ioutil.ReadAll(certFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "CTF main: Error reading cert:", err)
		return
	}

	cp := x509.NewCertPool()
	if ok := cp.AppendCertsFromPEM(cert); !ok {
		fmt.Fprintln(os.Stderr, "CTF main: Error adding cert:", err)
		return
	}

	cnf := tls.Config{RootCAs: cp}
	transport := http.Transport{TLSClientConfig: &cnf}

	client = &http.Client{Transport: &transport}

	ctf, err := NewCtf("ctf/config.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, "CTF main: Error creating CTF: ", err)
		return
	}

	http.HandleFunc("/vote", func(w http.ResponseWriter, r *http.Request) {
		votingHandler(w, r, ctf)
	})
	http.HandleFunc("/publish", func(w http.ResponseWriter, r *http.Request) {
		publishHandler(w, r, ctf)
	})
	fmt.Println("Listening and Serving...")
	http.ListenAndServeTLS(":9999", "certs/localhost.crt", "keys/localhost.key", nil)
}
