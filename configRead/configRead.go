package configRead

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// Struct for config
type Config struct {
	SpotifyKey      string `json:"spotify-key"`
	SpotifySecret   string `json:"spotify-secret"`
	LastFmKey       string `json:"lastfm-key"`
	LastFmSecret    string `json:"lastfm-secret"`
	HTTPPort        string `json:"http-port"`
	Hostname        string `json:"hostname,omitempty"`
	AuthRedirectURL string `json:"auth-redirect-url"`
}

func ReadConfig(path string) (Config, error) {
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
