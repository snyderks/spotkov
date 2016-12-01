package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/snyderks/spotkov/lastFm"
)

/*
 * These tests do not use the Spotify authentication.
 */

const lastFmID = "snyderks" // using my own id here
const baseLastURI = "http://ws.audioscrobbler.com/2.0/"

func getFirstLastFMPage(id string) lastFm.SongsPage {
	method := "user.getrecenttracks"
	api_key, key_success := os.LookupEnv("LASTFM_KEY")
	get_json := true
	if key_success == false {
		log.Fatal("couldn't get API key for LastFM from the env vars")
	}
	last_url := baseLastURI + "?method=" + method + "&user=" + id + "&api_key=" + api_key + "&limit=200"
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

	songs := lastFm.SongsPage{}
	err = json.Unmarshal(songsJSON, &songs)
	return songs
}

func TestGetListOfSongs(t *testing.T) {
	firstPage := getFirstLastFMPage(lastFmID)
	songsExpected, err := strconv.Atoi(firstPage.RecentTracks.Metadata.TotalSongs)

	if err != nil {
		t.Error("Couldn't convert songsExpected to a number.")
	}

	songs, _ := lastFm.ReadLastFMSongs(lastFmID)
	if len(songs) != songsExpected {
		t.Error("Didn't get the number of songs expected:", len(songs), "instead of", songsExpected)
	}

	for _, song := range songs {
		if len(song.Title) == 0 || len(song.Artist) == 0 {
			t.Error("A song wasn't retrieved with an artist and title:", song)
		}
	}
}
