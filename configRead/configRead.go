package configRead

import (
	"encoding/json"
	"errors"
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
		if len(config.AuthRedirectURL) == 0 {
			return Config{}, errors.New("Couldn't read environment variables")
		}
	}
	config := Config{}
	err = json.Unmarshal(file, &config)
	if err != nil {
		return Config{}, err
	}
	return config, err
}
