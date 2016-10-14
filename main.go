// Attempting to authenticate to spotify

package main

import (
  "fmt"
  "log"
  "net/http"
  "os"
  "io/ioutil"

  "github.com/zmb3/spotify"
)

const redirectURI = "http://localhost:8080/callback"
const baseLastURI = "http://ws.audioscrobbler.com/2.0/"

var (
  auth = spotify.NewAuthenticator(redirectURI, spotify.ScopeUserReadPrivate)
  ch = make(chan *spotify.Client)
  state = "abc123"
)

func main() {
  // start a local HTTP server
  http.HandleFunc("/callback", completeAuth) // paths ending in /callback call this!
  http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
    log.Println("Got request for:", r.URL.String())
  })
  go http.ListenAndServe(":8080", nil)

  url := auth.AuthURL(state)
  fmt.Println("Please log in to spotify by visiting the following page: ", url)

  // this assigns from the channel when it sees that the channel's been assigned to!
  client := <-ch

  // try and make a call that would fail if user wasn't logged in
  user, err := client.CurrentUser()
  if err != nil {
    log.Fatal(err)
  }
  fmt.Println("You are logged in as:", user.ID)

  // try to do things with last.fm
  method := "user.getrecenttracks"
  user_id := "snyderks"
  api_key, key_success := os.LookupEnv("LASTFM_KEY")
  fmt.Println("Key: ", api_key)
  get_json := true
  if key_success == false {
    log.Fatal("couldn't get API key for LastFM from the env vars")
  }
  last_url := baseLastURI + "?method=" + method + "&user=" + user_id + "&api_key=" + api_key
  if (get_json) {
    last_url += "&format=json"
  }
  fmt.Println(last_url)

  resp, err := http.Get(last_url)
  if err == nil {
    songs, err := ioutil.ReadAll(resp.Body)
    resp.Body.Close()
    if err != nil {
      log.Fatal("couldn't read the body of the last.fm response")
    }
    fmt.Printf("%s", songs)
  } else {
    fmt.Println(err);
  }
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
  tok, err := auth.Token(state, r)
  if err != nil {
    http.Error(w, "Couldn't get token", http.StatusForbidden)
    log.Fatal(err)
  }
  if st := r.FormValue("state"); st != state { // spotify returns the state key
    http.NotFound(w, r)      // passed to make sure the call wasn't intercepted
    log.Fatalf("State mismatch: %s != %s\n", st, state)
  }

  // get a client
  client := auth.NewClient(tok)
  fmt.Fprintf(w, "Login Completed!")
  ch <- &client // throw the channel a *reference* to the client (it wants a pointer)
}
