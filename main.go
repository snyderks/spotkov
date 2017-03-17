// Attempting to authenticate to spotify

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/snyderks/spotkov/lastFm"
	"github.com/snyderks/spotkov/markov"
	"github.com/snyderks/spotkov/spotifyPlaylistGenerator"

	"github.com/atotto/clipboard"
	"github.com/zmb3/spotify"
)

const redirectURI = "http://localhost:8080/callback"

var (
	scopes = []string{spotify.ScopeUserReadPrivate,
		spotify.ScopePlaylistReadPrivate,
		spotify.ScopePlaylistModifyPrivate,
		spotify.ScopePlaylistModifyPublic}
	auth  = spotify.NewAuthenticator(redirectURI, scopes...)
	ch    = make(chan *spotify.Client)
	state = "abc123"
)

type flags struct {
	lastFmUserId   string
	publicPlaylist bool
	playlistLength int
	song           string
	artist         string
}

func main() {
	args, keep_going := handleArgs()
	if keep_going == false {
		return
	}
	// start a local HTTP server
	http.HandleFunc("/callback", completeAuth) // paths ending in /callback call this!
	/*http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})*/
	go http.ListenAndServe(":8080", nil)

	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page\n(copied to your clipboard):\n\n", url)
	clipboard.WriteAll(url)
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

	titles, _ := lastFm.ReadLastFMSongs(args.lastFmUserId)

	if len(titles) > 0 {
		fmt.Println("Success! I got", len(titles), "titles from your Last.FM profile.")
	} else {
		panic("No titles were returned from Last.FM. Cannot continue.")
	}

	chain := markov.BuildChain(titles)
	if args.song == "" && args.artist == "" {
		reader := bufio.NewReader(os.Stdin)
		lastSong := titles[0]

		fmt.Println("\nThe last song you played was", lastSong.Title, "by", lastSong.Artist)
		fmt.Println("Use this as a starting point? (Yes/No) ")
		resp, _ := reader.ReadString('\n')
		resp = strings.TrimSpace(resp)

		result, valid := checkYesOrNo(resp)

		for valid == false {
			fmt.Println("\nInvalid response. Please type yes or no.")
			resp, _ = reader.ReadString('\n')
			resp = strings.TrimSpace(resp)

			result, valid = checkYesOrNo(resp)
		}
		if result == true {
			fmt.Println("\nOkay! I'll use that as the initial song.")
			args.song = lastSong.Title
			args.artist = lastSong.Artist
		} else {
			seedEntered := false
			for seedEntered == false {
				fmt.Println("\nI won't use that as the song. Enter a song name and artist to use (separated by new lines)")
				respTitle, _ := reader.ReadString('\n')
				respTitle = strings.TrimSpace(respTitle)
				respArtist, _ := reader.ReadString('\n')
				respArtist = strings.TrimSpace(respArtist)

				fmt.Println("\nIs", respTitle, "by", respArtist, "okay? (Yes/No) ")
				resp, _ = reader.ReadString('\n')
				resp = strings.TrimSpace(resp)

				result, valid = checkYesOrNo(resp)
				for valid == false {
					fmt.Println("\nInvalid response. Please type yes or no.")
					resp, _ = reader.ReadString('\n')
					resp = strings.TrimSpace(resp)
					result, valid = checkYesOrNo(resp)
				}
				if result == true {
					seedEntered = true
					args.artist = respArtist
					args.song = respTitle
				}
			}
		}
	}
	length := 20
	if args.playlistLength > 0 {
		length = args.playlistLength
	}
	list, err := markov.GenerateSongList(length, 1, lastFm.Song{Artist: args.artist, Title: args.song}, chain)
	createPlaylist := true
	if err != nil {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("An error was encountered when creating the list:", err)
		fmt.Println("Go ahead and create the playlist anyway? (Yes/No)")
		resp, _ := reader.ReadString('\n')
		resp = strings.TrimSpace(resp)

		result, valid := checkYesOrNo(resp)
		for valid == false {
			fmt.Println("\nInvalid response. Please type yes or no.")
			resp, _ = reader.ReadString('\n')
			resp = strings.TrimSpace(resp)
			result, valid = checkYesOrNo(resp)
		}
		if result == true {
			createPlaylist = true
		} else {
			fmt.Println("Okay. No changes to the playlist were made.")
			createPlaylist = false
		}
	}
	if createPlaylist == true {
		spotifyPlaylistGenerator.CreatePlaylist(list, client, user.ID)
	}
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

// processes the arguments passed and returns whether execution should continue.
func handleArgs() (flags, bool) {
	help := flag.Bool("help", false, "Description of the program and arguments")
	lastFm := flag.String("lastFm", "", "Your Last.FM User ID")
	publicPlaylist := flag.Bool("public", false, "Make the generated playlist public")
	playlistLength := flag.Int("length", 20, "Length of the generated playlist")
	songTitle := flag.String("title", "", "Title of the song to start with")
	songArtist := flag.String("artist", "", "Artist of the song to start with")

	flag.Parse()

	allFlags := flags{}

	if *help == true {
		fmt.Println("Spotkov is a Markov chain generator that uses your scrobbling history on Last.FM to find a playlist of songs that you've listened to and might like together and adds it to your Spotify profile.")
		fmt.Println("Use it like this:")
		fmt.Println("./spotkov -lastFm=your_Last.FM_user_id")
		fmt.Println("./spotkov -lastFm=your_Last.FM_user_id -public")
		fmt.Println("./spotkov -lastFm=your_Last.FM_user_id -length=45 -title=Madness -artist=Muse")
		return flags{}, false
	}

	if *lastFm == "" {
		var userId string
		fmt.Printf("Please enter your Last.FM user ID: ")
		_, err := fmt.Scanf("%s\n", &userId)
		if err != nil {
			panic(err)
		}
		allFlags.lastFmUserId = userId

	} else {
		allFlags.lastFmUserId = *lastFm
	}
	allFlags.publicPlaylist = *publicPlaylist
	allFlags.playlistLength = *playlistLength
	allFlags.song = *songTitle
	allFlags.artist = *songArtist

	return allFlags, true

}

func checkYesOrNo(resp string) (result, valid bool) {
	if strings.EqualFold(resp, "yes") || strings.EqualFold(resp, "y") {
		return true, true

	} else if strings.EqualFold(resp, "no") || strings.EqualFold(resp, "n") {
		return false, true
	} else {
		return false, false
	}
}
