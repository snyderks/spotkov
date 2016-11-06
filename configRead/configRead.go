package configRead

import (
	"encoding/json"
	"io/ioutil"
)

// Struct for config
type Config struct {
	SpotifyKey    string `json:"spotify-key"`
	SpotifySecret string `json:"spotify-secret"`
	LastFmKey     string `json:"lastfm-key"`
	LastFmSecret  string `json:"lastfm-secret"`
}

func ReadConfig(path string) (Config, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	config := Config{}
	err = json.Unmarshal(file, &config)
	if err != nil {
		return Config{}, err
	}
	return config, err
}
