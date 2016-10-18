// handles the retrieval of data from Last.FM (currently without authentication)

package lastFm

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
)

// Types to interpret JSON data returned from tracks played
type SongsPage struct {
	RecentTracks tracksWrapper `json:"recentTracks"`
}

type tracksWrapper struct {
	Tracks   []track  `json:"track"`
	Metadata metadata `json:"@attr"`
}

type metadata struct {
	UserId       string `json:"user"`
	Page         string `json:"page"`
	SongsPerPage string `json:"SongsPerPage"`
	TotalPages   string `json:"totalPages"`
	TotalSongs   string `json:"total"`
}

type track struct {
	Artist artist `json:"artist"`
	Title  string `json:"name"`
	Album  album  `json:"album"`
}

type artist struct {
	Title string `json:"#text"`
}

type album struct {
	Title string `json:"#text"`
}

type Song struct {
	Artist string
	Title string
}

var pagesWg sync.WaitGroup

const baseLastURI = "http://ws.audioscrobbler.com/2.0/"

func ReadLastFMSongs(user_id string) []Song {
	// try to do things with last.fm
	method := "user.getrecenttracks"
	api_key, key_success := os.LookupEnv("LASTFM_KEY")
	get_json := true
	if key_success == false {
		log.Fatal("couldn't get API key for LastFM from the env vars")
	}
	last_url := baseLastURI + "?method=" + method + "&user=" + user_id + "&api_key=" + api_key + "&limit=200"
	if get_json {
		last_url += "&format=json"
	}
	resp, err := http.Get(last_url)
	var songsJSON []byte
	if err == nil {
		songsJSON, err = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Fatal("Couldn't read the body of the last.fm response")
		}
	} else {
		fmt.Println(err)
	}

	songs := SongsPage{}
	err = json.Unmarshal(songsJSON, &songs)

	pageSongs := make([]Song, 0, 50)

	for _, track := range songs.RecentTracks.Tracks {
		pageSongs = append(pageSongs, Song{track.Artist.Title, track.Title})
	}

	max_page, _ := strconv.Atoi(songs.RecentTracks.Metadata.TotalPages)

	songPages := make([][]Song, max_page)

	songPages[0] = pageSongs

	for i := 2; i <= max_page; i++ {
		pagesWg.Add(1)
		go getLastFMPagesAsync(last_url, i, max_page, songPages)
	}

	pagesWg.Wait()

	titlesConcat := make([]Song, 0)

	for i := 0; i < max_page; i++ {
		titlesConcat = append(titlesConcat, songPages[i]...)
	}

	return titlesConcat

}

func getLastFMPagesAsync(url string, page int, max_page int, allTitles [][]Song) {
	defer pagesWg.Done()
	pageStr := strconv.Itoa(page)
	resp, err := http.Get(url + "&page=" + pageStr)
	songs := SongsPage{}
	if err == nil {
		songsJSON, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		err = json.Unmarshal(songsJSON, &songs)

		if err != nil {
			log.Fatal("couldn't read the body of the last.fm response")
		}
	} else {
		fmt.Println(err)
	}
	tracksRaw := songs.RecentTracks.Tracks
	titles := make([]Song, 0)
	for _, track := range tracksRaw {
		titles = append(titles, Song{track.Artist.Title, track.Title})
	}
	allTitles[page-1] = titles
}
