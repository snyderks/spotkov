// Package configRead reads configuration information from either a JSON file or from
// static environment variables as a fallback.
package configRead

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"strings"
)

// Config holds the different configuration options.
type Config struct {
	SpotifyKey      string `json:"spotify-key"`
	SpotifySecret   string `json:"spotify-secret"`
	LastFmKey       string `json:"lastfm-key"`
	LastFmSecret    string `json:"lastfm-secret"`
	HTTPPort        string `json:"http-port"`
	Hostname        string `json:"hostname,omitempty"`
	AuthRedirectURL string `json:"auth-redirect-url"`
	Debug           bool   `json:"debug"`
}

// Read takes a path to a JSON file.
// If it fails to read the file, it falls back to environment variables.
// Returns an error if it can't parse the JSON file or if it can't read environment variables.
func Read(path string) (Config, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil { // not using json config. Try to get it from env vars
		config := Config{
			SpotifyKey:      os.Getenv("SPOTIFY_KEY"),
			SpotifySecret:   os.Getenv("SPOTIFY_SECRET"),
			LastFmKey:       os.Getenv("LASTFM_KEY"),
			LastFmSecret:    os.Getenv("LASTFM_SECRET"),
			HTTPPort:        os.Getenv("PORT"),
			Hostname:        os.Getenv("HOSTNAME"),
			AuthRedirectURL: os.Getenv("AUTH_REDIRECT"),
			Debug:           os.Getenv("DEBUG") == "1",
		}
		if !strings.Contains(config.HTTPPort, ":") {
			config.HTTPPort = ":" + config.HTTPPort
		}
		if len(config.AuthRedirectURL) == 0 {
			return Config{}, errors.New("Couldn't read environment variables")
		}
		return config, nil
	}
	config := Config{}
	err = json.Unmarshal(file, &config)
	if err != nil {
		return Config{}, err
	}
	return config, err
}
