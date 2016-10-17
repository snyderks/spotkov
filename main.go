// Attempting to authenticate to spotify

package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"flag"

	"github.com/snyderks/spotkov/markov"
	"github.com/snyderks/spotkov/lastFm"
	"github.com/snyderks/spotkov/spotifyPlaylistGenerator"

	"github.com/zmb3/spotify"
)

const redirectURI = "http://localhost:8080/callback"

var (
	scopes = []string {spotify.ScopeUserReadPrivate,
										 spotify.ScopePlaylistReadPrivate,
										 spotify.ScopePlaylistModifyPrivate,
										 spotify.ScopePlaylistModifyPublic}
	auth  = spotify.NewAuthenticator(redirectURI, scopes...)
	ch    = make(chan *spotify.Client)
	state = "abc123"
)

type flags struct {
	lastFmUserId string
	publicPlaylist bool
	playlistLength int
	startingSong string
	startingArtist string
}

func main() {
	_, keep_going := handleArgs()
	if keep_going == false {
		return
	}
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

	lastFMUserId := "snyderks"

	titles := lastFm.ReadLastFMSongs(lastFMUserId)

	if len(titles) > 0 {
		fmt.Println("Success! I got", len(titles), "titles.")
	}

	chain := markov.BuildChain(titles)
	fmt.Println("Generating song list")
	list := markov.GenerateSongList(20, lastFm.Song{Artist: "Meg Myers", Title: "Lemon Eyes"}, chain)
	fmt.Println("Got song list")
	spotifyPlaylistGenerator.CreatePlaylist(list, client, user.ID)
	fmt.Println("Came back from creating the playlist")
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
// processes the arguments passed and returns whether execution should continue.
func handleArgs() (flags, bool) {
	help := flag.Bool("help", false, "Description of the program and arguments")
	lastFm := flag.String("lastFm", "", "Your Last.FM User ID")
	publicPlaylist := flag.Bool("public", false, "Make the generated playlist public")
	playlistLength := flag.Int("length", 20, "Length of the generated playlist")
	songTitle := flag.String("title", "", "Title of the song to start with")
	songArtist := flag.String("artist", "", "Artist of the song to start with")

	flag.Parse()

	if *help == true {
		fmt.Println("Spotkov is a Markov chain generator that uses your scrobbling history on Last.FM to find a playlist of songs that you've listened to and might like together and adds it to your Spotify profile.")
		fmt.Println("Use it like this:")
		fmt.Println("./spotkov -lastFm=\"your_Last.FM_user_id\"")
		fmt.Println("./spotkov -lastFm=\"your_Last.FM_user_id\" -public")
		fmt.Println("./spotkov -lastFm=\"your_Last.FM_user_id\" -length=45 -title=\"Madness\" -artist=\"Muse\"")
		return flags{}, false
	}
	return flags {
		lastFmUserId: *lastFm,
		publicPlaylist: *publicPlaylist,
		playlistLength: *playlistLength,
		startingSong: *songTitle,
		startingArtist: *songArtist }, true
	
}
