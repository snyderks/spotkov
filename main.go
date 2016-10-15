// Attempting to authenticate to spotify

package main

import (
  "fmt"
  "log"
  "net/http"
  "os"
  "os/exec"
  "io/ioutil"
  //"strings"
  "runtime"
  "strconv"
  "encoding/json"
  "sync"

  "github.com/zmb3/spotify"
)

type SongsPage struct {
  RecentTracks tracksWrapper `json:"recentTracks"`
}

type tracksWrapper struct {
  Tracks []track `json:"track"`
  Metadata metadata `json:"@attr"`
}

type metadata struct {
  UserId string `json:"user"`
  Page string `json:"page"`
  SongsPerPage string `json:"SongsPerPage"`
  TotalPages string `json:"totalPages"`
  TotalSongs string `json:"total"`
}

type track struct {
  Artist artist `json:"artist"`
  Title string `json:"name"`
  Album album `json:"album"`
}

type artist struct {
  Title string `json:"#text"`
}

type album struct {
  Title string `json:"#text"`
}

const redirectURI = "http://localhost:8080/callback"
const baseLastURI = "http://ws.audioscrobbler.com/2.0/"

var (
  auth = spotify.NewAuthenticator(redirectURI, spotify.ScopeUserReadPrivate)
  ch = make(chan *spotify.Client)
  state = "abc123"
  pagesWg sync.WaitGroup
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

  titles := readLastFMSongs(user_id)

  if len(titles) > 0 {
    fmt.Println("Success! I got", len(titles), "titles.")
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

func readLastFMSongs (user_id string) []string {
  // try to do things with last.fm
  method := "user.getrecenttracks"
  api_key, key_success := os.LookupEnv("LASTFM_KEY")
  fmt.Println("Key: ", api_key)
  get_json := true
  if key_success == false {
    log.Fatal("couldn't get API key for LastFM from the env vars")
  }
  last_url := baseLastURI + "?method=" + method + "&user=" + user_id + "&api_key=" + api_key + "&limit=200"
  if (get_json) {
    last_url += "&format=json"
  }
  resp, err := http.Get(last_url)
  var songsJSON []byte
  if err == nil {
    songsJSON, err = ioutil.ReadAll(resp.Body)
    resp.Body.Close()
    if err != nil {
      log.Fatal("couldn't read the body of the last.fm response")
    }
    fmt.Println("Printed page 1")
  } else {
    fmt.Println(err);
  }

  songs := SongsPage{}
  err = json.Unmarshal(songsJSON, &songs)

  pageTitles := make([]string, 50)

  for _, track := range songs.RecentTracks.Tracks {
    pageTitles = append(pageTitles, track.Title)
  }

  max_page, _ := strconv.Atoi(songs.RecentTracks.Metadata.TotalPages)

  titles := make([][]string, max_page)

  titles[0] = pageTitles

  for i := 2; i <= max_page; i++ {
    pagesWg.Add(1)
    go getLastFMPagesAsync(last_url, i, max_page, titles)
  }

  pagesWg.Wait()

  titlesConcat := make([]string, 0)

  for i := 1; i < max_page; i++ {
    titlesConcat = append(titlesConcat, titles[i]...)
  }

  fmt.Println("I got", len(titles), "pages total.")
  fmt.Println("Expected", max_page, "pages.")

  return titlesConcat

}

func getLastFMPagesAsync (url string, page int, max_page int, allTitles [][]string) {
  defer pagesWg.Done()
  pageStr := strconv.Itoa(page)
  resp, err := http.Get(url + "&page=" + pageStr)
  songs := SongsPage{}
  if err == nil {
    fmt.Println("received page ", page)
    songsJSON, err := ioutil.ReadAll(resp.Body)
    resp.Body.Close()
    err = json.Unmarshal(songsJSON, &songs)

    if err != nil {
      log.Fatal("couldn't read the body of the last.fm response")
    }
  } else {
    fmt.Println(err);
  }
  tracksRaw := songs.RecentTracks.Tracks
  titles := make([]string, 0)
  for _, track := range tracksRaw {
    titles = append(titles, track.Title)
  }
  index, _ := strconv.Atoi(songs.RecentTracks.Metadata.Page)
  allTitles[index - 1] = titles
}
