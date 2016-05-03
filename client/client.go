package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
)

const (
<<<<<<< HEAD
	HOME     = "client/templates/home.html"
	REGISTER = "client/templates/register.html"
	VOTE     = "client/templates/vote.html"
	CSS		 = "client/static/style.css"
=======
	HOME       = "client/templates/home.html"
	REGISTER   = "client/templates/register.html"
	VOTE       = "client/templates/vote.html"
	NO_VOTE    = "client/templates/no_vote.html"
	PUBLISH    = "client/templates/publish.html"
	NO_PUBLISH = "client/templates/no_publish.html"
>>>>>>> 55d69d8f1182c1023286ade5046d74ae93b97b29
)

type Message struct {
	Text string
	Type string
}

type Registration struct {
	Name         string
	SharedSecret string
}

type Vote struct {
	Candidate     string
	ValidationNum string
}

type Candidate struct {
	Name      string
	VoteCount int
	VoterIDs  []string
}

type PublishResp struct {
	Candidates []*Candidate
}

var client *http.Client

func mainHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, &Message{Text: "Welcome to the election booth!", Type: "msg"}, HOME)
	return
}

func registrationGet(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, &Message{Text: "Register to vote!", Type: "msg"}, REGISTER)
	return
}

func registrationPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error parsing form", err)
		return
	}

	args := Registration{
		Name:         r.Form.Get("name"),
		SharedSecret: r.Form.Get("shared_secret"),
	}
	b, err := json.Marshal(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to marshal arguments", err)
		return
	}

	buf := bytes.NewBuffer(b)
	resp, err := client.Post("https://localhost:9889/register", "application/json", buf)
	if err != nil {
		fmt.Fprintln(os.Stderr, "client registration: Post failed:", err)
		return
	}

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintln(os.Stderr, "client registration: Error opening response:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusBadRequest {
		renderError(w, r, string(contents), err, REGISTER)
		return
	}

	if resp.StatusCode == http.StatusUnauthorized {
		m := &Message{Text: "Registration is no longer underway!", Type: "msg"}
		renderTemplate(w, r, m, REGISTER)
		return
	}

	m := &Message{Text: "Success! Validation number is: " + string(contents), Type: "msg"}
	renderTemplate(w, r, m, REGISTER)
	return
}

func votingPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error parsing form", err)
		return
	}

	args := Vote{
		Candidate:     r.Form.Get("candidate"),
		ValidationNum: r.Form.Get("validation"),
	}
	b, err := json.Marshal(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to marshal arguments", err)
		return
	}

	buf := bytes.NewBuffer(b)
	resp, err := client.Post("https://localhost:9999/vote", "application/json", buf)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Post failed", err)
		return
	}
	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error opening response", err)
		return
	}

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusBadRequest {
		renderError(w, r, string(contents), err, VOTE)
		return
	}

	m := &Message{Text: "Successfully voted for " + args.Candidate + "!", Type: "msg"}
	renderTemplate(w, r, m, VOTE)
	return
}

<<<<<<< HEAD
func staticHandler(w http.ResponseWriter, r *http.Request) {

=======
func votingGet(w http.ResponseWriter, r *http.Request) {
	var msg string
	var page string

	resp, err := client.Get("https://localhost:9889/ready")
	if err != nil {
		fmt.Fprintln(os.Stderr, "client: Get to CLA failed:", err)
		return
	}

	if resp.StatusCode == http.StatusUnauthorized {
		msg = "Voting has not yet begun!"
		page = NO_VOTE
	} else {
		msg = "Vote now!"
		page = VOTE
	}

	renderTemplate(w, r, &Message{Text: msg, Type: "msg"}, page)
	return
}

func publishHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := client.Get("https://localhost:9889/ready")
	if err != nil {
		fmt.Fprintln(os.Stderr, "client: Get to CLA failed:", err)
		return
	}

	if resp.StatusCode == http.StatusUnauthorized {
		renderError(w, r, "Voting has not even begun!", nil, NO_PUBLISH)
		return
	}

	resp, err = client.Get("https://localhost:9999/publish")
	if err != nil {
		fmt.Fprintln(os.Stderr, "client: Get to CTF failed:", err)
		return
	}

	if resp.StatusCode == http.StatusUnauthorized {
		renderError(w, r, "Voting has not yet ended!", nil, NO_PUBLISH)
	} else if resp.StatusCode == http.StatusOK {
		var contents PublishResp
		c, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintln(os.Stderr, "client: Failed to read resp:", err)
		}

		err = json.Unmarshal(c, &contents)
		if err != nil {
			fmt.Fprintln(os.Stderr, "client: json failed to unmarshal:", err)
			return
		}

		candidates := contents.Candidates

		tmpl, err := template.ParseFiles(PUBLISH)
		if err != nil {
			http.NotFound(w, r)
		}
		tmpl.Execute(w, map[string]interface{}{
			"Candidates": candidates,
		})
	}
>>>>>>> 55d69d8f1182c1023286ade5046d74ae93b97b29
}

func renderError(w http.ResponseWriter, r *http.Request, text string, err error, page string) {
	if err != nil {
		fmt.Printf(err.Error())
	}
	msg := &Message{Text: text, Type: "error"}
	renderTemplate(w, r, msg, page)
	return
}

func renderTemplate(w http.ResponseWriter, r *http.Request, msg *Message, page string) {
	tmpl, err := template.ParseFiles(page)
	if err != nil {
		http.NotFound(w, r)
	}
	tmpl.Execute(w, msg)
	return
}

func main() {
	certFile, err := os.Open("certs/ca.crt")
	if err != nil {
		fmt.Fprintln(os.Stderr, "client main: Error opening cert:", err)
		return
	}
	defer certFile.Close()

	cert, err := ioutil.ReadAll(certFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "client main: Error reading cert:", err)
		return
	}

	cp := x509.NewCertPool()
	if ok := cp.AppendCertsFromPEM(cert); !ok {
		fmt.Fprintln(os.Stderr, "client main: Error adding cert:", err)
		return
	}

	cnf := tls.Config{RootCAs: cp}
	transport := http.Transport{TLSClientConfig: &cnf}

	client = &http.Client{Transport: &transport}
	http.HandleFunc("/", mainHandler)
	http.HandleFunc("/registration", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			registrationGet(w, r)
		} else if r.Method == "POST" {
			registrationPost(w, r)
		}
	})
<<<<<<< HEAD
	http.HandleFunc("/vote", votingHandler)
	// handle css
	// Normal resources
    http.HandleFunc("/static", staticHandler)
=======
	http.HandleFunc("/vote", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			votingGet(w, r)
		} else if r.Method == "POST" {
			votingPost(w, r)
		}
	})
	http.HandleFunc("/results", publishHandler)
>>>>>>> 55d69d8f1182c1023286ade5046d74ae93b97b29
	fmt.Println("Listening and serving...")
	http.ListenAndServe(":8998", nil)
}
