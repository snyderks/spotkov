// Attempting to authenticate to spotify

package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"

	"github.com/snyderks/spotkov/lastFm"
	"github.com/snyderks/spotkov/markov"

	"github.com/zmb3/spotify"
)

const redirectURI = "http://localhost:8080/callback"

var (
	auth  = spotify.NewAuthenticator(redirectURI, spotify.ScopeUserReadPrivate)
	ch    = make(chan *spotify.Client)
	state = "abc123"
)

func main() {
	// start a local HTTP server
	http.HandleFunc("/callback", completeAuth) // paths ending in /callback call this!
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})
	go http.ListenAndServe(":8080", nil)

	url := auth.AuthURL(state)
	fmt.Println("Please log in to spotify by visiting the following page: ", url)
	//fmt.Println("Type yes to open this page in your browser, enter to continue")
	//var openResp string
	/*_, err := fmt.Scanf("%s\n", &openResp) // need the reference otherwise a copy is made
	  if err == nil && strings.EqualFold(openResp, "yes") {
	    fmt.Println("Opening in browser...")
	    open(url)
	  }*/ // this doesn't work right now; only first param is actually sent in url

	// this assigns from the channel when it sees that the channel's been assigned to!
	client := <-ch

	// try and make a call that would fail if user wasn't logged in
	user, err := client.CurrentUser()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("You are logged in as:", user.ID)

	user_id := "snyderks"

	titles := lastFm.ReadLastFMSongs(user_id)

	if len(titles) > 0 {
		fmt.Println("Success! I got", len(titles), "titles.")
	}

	markov.BuildChain(titles)
}

// internal startup functions

func completeAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.Token(state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state { // spotify returns the state key
		http.NotFound(w, r) // passed to make sure the call wasn't intercepted
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}

	// get a client
	client := auth.NewClient(tok)
	fmt.Fprintf(w, "Login Completed!")
	ch <- &client // throw the channel a *reference* to the client (it wants a pointer)
}

// open opens the specified URL in the default browser of the user.
// from http://stackoverflow.com/questions/39320371/how-start-web-server-to-open-page-in-browser-in-golang
func open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}
