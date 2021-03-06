// Package lastFm handles the retrieval of song data from Last.FM.
package lastFm

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/snyderks/spotkov/configRead"
	"github.com/snyderks/spotkov/tools"
)

// SongsPage holds a list of tracks in a page.
type SongsPage struct {
	RecentTracks tracksWrapper `json:"recentTracks"`
}

// tracksWrapper holds a list of tracks and various information about the page
// retrieved from the server.
type tracksWrapper struct {
	Tracks   []track  `json:"track"`
	Metadata metadata `json:"@attr"`
}

// metadata contains general information about the page retrieved.
type metadata struct {
	UserId       string `json:"user"`
	Page         string `json:"page"`
	SongsPerPage string `json:"SongsPerPage"`
	TotalPages   string `json:"totalPages"`
	TotalSongs   string `json:"total"`
}

// track holds all information about the song retrieved.
type track struct {
	Artist     artist                 `json:"artist"`
	Title      string                 `json:"name"`
	Album      album                  `json:"album"`
	Timestamp  trackDate              `json:"date"`
	Attributes map[string]interface{} `json:"@attr"`
}

// trackDate holds both a Unix representation
// and a text representation of the date and time
// the track was scrobbled.
type trackDate struct {
	UnixTime string `json:"uts"`
	TextDate string `json:"#text"`
}

// artist is the name of an artist.
type artist struct {
	Title string `json:"#text"`
}

// album is the name of an album.
type album struct {
	Title string `json:"#text"`
}

// Song has an artist name, the title of the song, and when the song
// was scrobbled by the user.
type Song struct {
	Artist    string
	Title     string
	Timestamp time.Time
}

// BaseSong has an artist name and the title of the song.
type BaseSong struct {
	Artist string
	Title  string
}

// songFile contains a list of Songs.
type songFile struct {
	Songs []Song
}

// SongMap wraps a map of songs for easy serialization.
type SongMap struct {
	Songs map[BaseSong]bool
}

// lastFMError contains the format of an error received if something
// went wrong during an API call.
type lastFMError struct {
	Error   int    `json:"error"`
	Message string `json:"message"`
}

// Redis key prefixes for reading and writing song data.
const allSongCachePrefix = "songCache."
const uniqueCachePrefix = "uniqueCache."

var UseRedis bool
var c *redis.Client

func init() {
	UseRedis = true
	config, err := configRead.Read("config.json")
	rURL := "localhost:6379"
	if err == nil {
		rURL = config.RedisURL
	}
	password := ""

	// Need to construct a URL if it's not using localhost
	if !strings.Contains(rURL, "localhost") {
		parsedURL, _ := url.Parse(rURL)
		password, _ = parsedURL.User.Password()
		rURL = parsedURL.Host
	}
	c = redis.NewClient(&redis.Options{
		Addr:     rURL,
		Password: password,
		DB:       0, // use default DB
	})
	_, err = c.Ping().Result()
	if err != nil {
		UseRedis = false
		fmt.Println(err.Error())
	}
}

func ReadCache(userID string, cachePrefix string, songs interface{}) error {
	if UseRedis {
		// Send the command to retrieve the cache to Redis.
		s, err := c.Get(cachePrefix + userID).Result()
		if err != nil {
			return errors.New(fmt.Sprintf("Error occurred in Redis request: %s", err.Error()))
		}

		// Attempt to convert from a base64 representation.
		err = tools.FromBase64(s, songs)
		if err != nil {
			return errors.New(fmt.Sprintf("Couldn't convert Redis response: %s", err.Error()))
		}
		return nil
	}
	return errors.New("Attempted to read the cache without a connection to Redis.")
}

func WriteCache(userID string, cachePrefix string, songs interface{}) error {
	if UseRedis {
		// Attempt to convert to a base64 representation.
		b64, err := tools.ToBase64(songs)

		if err != nil {
			return errors.New(fmt.Sprintf("Error encoding the SongMap: %s", err.Error()))
		}
		err = c.Set(cachePrefix+userID, b64, 0).Err()
		if err != nil {
			return errors.New(fmt.Sprintf("Error sending the SET request to Redis: %s", err.Error()))
		}
		return nil
	}
	return errors.New("Attempted to write to cache without a connection to Redis.")
}

// ReadCachedUniqueSongs reads back a cache of mapped songs from the local directory.
func ReadCachedUniqueSongs(userID string, songs *SongMap) error {
	return ReadCache(userID, uniqueCachePrefix, songs)
}

// readCachedSongs reads any existing song data about a user and
// stores that data into the songs argument.
func readCachedSongs(userID string, songs *songFile) error {
	return ReadCache(userID, allSongCachePrefix, songs)
}

// cacheSongs takes song data and stores it in a binary data format
// used by golang called a gob.
func cacheSongs(userID string, songs songFile) error {
	return WriteCache(userID, allSongCachePrefix, songs)
}

// cacheUniqueSongs saves a map of songs to the local directory.
func cacheUniqueSongs(userID string, songs SongMap) error {
	return WriteCache(userID, uniqueCachePrefix, songs)
}

// pagesWg manages the number of pages currently being searched for.
var pagesWg sync.WaitGroup

// baseLastURI is the root of the API path for Last.FM.
const baseLastURI = "http://ws.audioscrobbler.com/2.0/"

// ReadLastFMSongs retrieves all scrobbled Last.FM songs for a specific user.
// Returns an error on failure.
func ReadLastFMSongs(userID string) ([]Song, error) {
	var uniques SongMap
	err := ReadCachedUniqueSongs(userID, &uniques)

	// didn't find or couldn't access the cache.
	// make a new map instead
	if err != nil {
		uniques.Songs = make(map[BaseSong]bool)
		err = nil
	}

	file := songFile{}
	err = readCachedSongs(userID, &file)
	titlesConcat := file.Songs
	if len(titlesConcat) == 0 {
		err = errors.New("Length of cached songs is 0. Regenerating...")
	}

	var errLastFM lastFMError

	if err != nil { // couldn't retrieve a cached version
		titlesConcat, errLastFM = getAllTitles(make([]Song, 0), &uniques, time.Time{}, userID)
	} else {
		var lastDate time.Time
		for _, song := range titlesConcat {
			if !song.Timestamp.IsZero() {
				lastDate = song.Timestamp
				break
			}
		}
		titlesConcat, errLastFM = getAllTitles(titlesConcat, &uniques, lastDate, userID)
	}

	if errLastFM.Error != 0 {
		return nil, errors.New("Generating the playlist failed. Please try again with the same or a different song.")
	}

	err = cacheSongs(userID, songFile{titlesConcat})
	if err != nil {
		fmt.Println("Couldn't cache the songs:", err.Error())
		// Don't actually want to return an error to the caller. Printing is enough.
		err = nil
	}

	err = cacheUniqueSongs(userID, uniques)
	if err != nil {
		fmt.Println("Couldn't cache unique songs:", err.Error())
		// See above. Don't want to return an error.
		err = nil
	}

	if len(titlesConcat) == 0 {
		err = errors.New("Failed to retrieve any play history. Please try again.")
	}

	return titlesConcat, err

}

// getAllTitles takes a list of songs and returns the songs for the user scrobbled after a certain time.
// Returns an error if something goes wrong.
func getAllTitles(titles []Song, uniques *SongMap, startTime time.Time, user_id string) (newTitles []Song, errLastFM lastFMError) {
	defer func() {
		if r := recover(); r != nil {
			errLastFM.Error = r.(int)
			errLastFM.Message = r.(string)
			newTitles = make([]Song, 0)
		}
	}()
	// try to do things with last.fm
	method := "user.getrecenttracks"
	api_key, key_success := os.LookupEnv("LASTFM_KEY")
	get_json := true
	if key_success == false {
		config, err := configRead.Read("config.json")
		if err != nil {
			panic("Couldn't read config or get env vars")
		} else {
			api_key = config.LastFmKey
		}
	}
	urlTime := "0"
	if !startTime.IsZero() {
		timeInt := startTime.UTC().Unix()
		if timeInt > 0 {
			urlTime = strconv.FormatInt(timeInt+1, 10)
		}
	}
	last_url := baseLastURI + "?method=" + method + "&user=" + user_id + "&api_key=" + api_key +
		"&limit=200" + "&from=" + urlTime
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
	// We don't want the currently playing track there. This checks for that.
	containsNowPlaying := false
	if len(songs.RecentTracks.Tracks) > 0 {
		if songs.RecentTracks.Tracks[0].Attributes != nil &&
			songs.RecentTracks.Tracks[0].Attributes["nowplaying"].(string) == "true" {
			containsNowPlaying = true
		}
	}
	if containsNowPlaying {
		songs.RecentTracks.Tracks = songs.RecentTracks.Tracks[1:]
	}
	topIndex := len(songs.RecentTracks.Tracks) - 1
	for i := topIndex; i >= 0; i-- {
		if topIndex-i != i {
			temp := songs.RecentTracks.Tracks[i]
			songs.RecentTracks.Tracks[i] = songs.RecentTracks.Tracks[topIndex-i]
			songs.RecentTracks.Tracks[topIndex-i] = temp
		}
	}
	pageSongs := make([]Song, 0, 50)

	for _, track := range songs.RecentTracks.Tracks {
		utime, err := strconv.ParseInt(track.Timestamp.UnixTime, 10, 64)
		var ts time.Time
		if err == nil {
			ts = time.Unix(utime, 0)
		}
		pageSongs = append(pageSongs, Song{track.Artist.Title, track.Title, ts})
	}

	max_page, _ := strconv.Atoi(songs.RecentTracks.Metadata.TotalPages)

	if max_page < 1 {
		max_page = 1
	}

	songPages := make([][]Song, max_page)

	songPages[0] = pageSongs

	batchAmt := 100

	if max_page > batchAmt { // have to batch to avoid socket overload
		for i := 0; i <= max_page/batchAmt; i++ {
			var maxBatch int
			if (i+1)*batchAmt > max_page {
				maxBatch = max_page
			} else {
				maxBatch = (i + 1) * batchAmt
			}
			if i == 0 {
				for j := 2; j <= maxBatch; j++ {
					pagesWg.Add(1)
					go getLastFMPagesAsync(last_url, j, maxBatch, songPages)
				}
			} else {
				for j := i*batchAmt + 1; j <= maxBatch; j++ {
					pagesWg.Add(1)
					go getLastFMPagesAsync(last_url, j, maxBatch, songPages)
				}
			}
			pagesWg.Wait()
		}
	} else {
		for i := 2; i <= max_page; i++ {
			pagesWg.Add(1)
			go getLastFMPagesAsync(last_url, i, max_page, songPages)
		}
		pagesWg.Wait()
	}

	// reversing all of the pages
	topIndex = len(songPages) - 1
	for i := topIndex; i >= 0; i-- {
		if topIndex-i != i {
			temp := songPages[i]
			songPages[i] = songPages[topIndex-i]
			songPages[topIndex-i] = temp
		}
	}
	if !startTime.IsZero() && len(titles) > 0 {
		// normal append for a new list
		for i := 0; i < max_page; i++ {
			titles = append(songPages[i], titles...)
			for _, el := range songPages[i] {
				s := BaseSong{Artist: el.Artist, Title: el.Title}
				if !uniques.Songs[s] {
					uniques.Songs[s] = true
				}
			}
		}
	} else {
		// if we're adding to the cache, have to prepend the new songs
		for i := 0; i < max_page; i++ {
			titles = append(titles, songPages[i]...)
			for _, el := range songPages[i] {
				s := BaseSong{Artist: el.Artist, Title: el.Title}
				if !uniques.Songs[s] {
					uniques.Songs[s] = true
				}
			}
		}
	}

	return titles, lastFMError{}
}

// getLastFMPagesAsync populates allTitles with lists of lists of songs.
// Fully encapsulates all async work.
func getLastFMPagesAsync(url string, page int, max_page int, allTitles [][]Song) {
	defer pagesWg.Done()
	pageStr := strconv.Itoa(page)
	songs := SongsPage{}
	rateLimited := true
	tries := 0
	for rateLimited == true && tries < 4 {
		tr := &http.Transport{
			DisableKeepAlives: true,
		}
		c := &http.Client{Transport: tr, Timeout: 5 * time.Second}
		resp, err := c.Get(url + "&page=" + pageStr)
		if err == nil && resp.StatusCode == http.StatusOK {
			songsJSON, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			err = json.Unmarshal(songsJSON, &songs)
			if err != nil {
				rateLimited = true
				tries = tries + 1
			} else {
				rateLimited = false
			}
		}
	}

	var tracksRaw []track
	// Eliminate currently playing track if returned.
	containsNowPlaying := false
	if len(songs.RecentTracks.Tracks) > 0 {
		if songs.RecentTracks.Tracks[0].Attributes != nil &&
			songs.RecentTracks.Tracks[0].Attributes["nowplaying"].(string) == "true" {
			containsNowPlaying = true
		}
	}
	if containsNowPlaying {
		tracksRaw = songs.RecentTracks.Tracks[1:]
	} else {
		tracksRaw = songs.RecentTracks.Tracks
	}
	// Reverse the array so that the suffixes are built in the right order.
	topIndex := len(tracksRaw) - 1
	for i := topIndex; i >= 0; i-- {
		if topIndex-i != i {
			temp := tracksRaw[i]
			tracksRaw[i] = tracksRaw[topIndex-i]
			tracksRaw[topIndex-i] = temp
		}
	}
	titles := make([]Song, 0)
	for _, track := range tracksRaw {
		utime, err := strconv.ParseInt(track.Timestamp.UnixTime, 10, 64)
		var ts time.Time
		if err == nil {
			ts = time.Unix(utime, 0)
		}
		titles = append(titles, Song{track.Artist.Title, track.Title, ts})
	}
	allTitles[page-1] = titles
}
