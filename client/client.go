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
	HOME     = "client/templates/home.html"
	REGISTER = "client/templates/register.html"
	VOTE     = "client/templates/vote.html"
	NO_VOTE    = "client/templates/no_vote.html"
 	PUBLISH    = "client/templates/publish.html"
 	NO_PUBLISH = "client/templates/no_publish.html"
  )
	CSS		 = "client/static/style.css"
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
		renderError(w, r, string(contents), err, REGISTER)
		return
	}

	if resp.StatusCode == http.StatusTeapot {
		m := &Message{Text: "Voting is no longer underway!", Type: "msg"}
		renderTemplate(w, r, m, REGISTER)
		return
	}

	m := &Message{Text: "Success! Validation number is: " + string(contents), Type: "msg"}
	renderTemplate(w, r, m, REGISTER)
	return
}

func votingHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, &Message{Text: "Voting has not yet begun!", Type: "msg"}, VOTE)
	return
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
		fmt.Fprintf(os.Stderr, "Error opening cert", err)
		return
	}
	defer certFile.Close()

	cert, err := ioutil.ReadAll(certFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading cert", err)
		return
	}

	cp := x509.NewCertPool()
	if ok := cp.AppendCertsFromPEM(cert); !ok {
		fmt.Fprintf(os.Stderr, "Error adding cert", err)
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
	http.HandleFunc("/vote", votingHandler)
	// handle css
    http.Handle("/static", http.StripPrefix("/static/", http.FileServer(http.Dir(CSS))))
	fmt.Println("Serving static at ", CSS)
	fmt.Println("Listening and serving...")
	http.ListenAndServe(":8998", nil)
}
