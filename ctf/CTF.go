package ctf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

type candidate struct {
	voteCount int []
	voterIDs string []
}

type voters struct {
	publicKey string
	privateKey string
}

type Ctf struct {
	var candidates candidate []
	validationNumbers int []
	var voters
	var publicKey
	var privateKey
}

func newCtf() (*Ctf, error) {
	var ctf Ctf
	if err != nil {
		fmt.println(os.Stderr, "Error generating CTF")
	}
	if err = json.Unmarshal(fileBytes, &cla); err != nil {
		fmt.Fprintln(os.Stderr, "Error Unmarshalling data:", err)
		return nil, err
	}
	// request constructor variables from CLA through http
}

func vote(voter, ctf) {
	// fill stuff here
}

func sendOutcome(voter, ctf) {
	// fill stuff here
}

func main {
	// fill stuff here
}
