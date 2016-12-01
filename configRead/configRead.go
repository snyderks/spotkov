package configRead

import (
	"encoding/json"
	"io/ioutil"
)

// Struct for config
type Config struct {
	SpotifyKey      string `json:"spotify-key"`
	SpotifySecret   string `json:"spotify-secret"`
	LastFmKey       string `json:"lastfm-key"`
	LastFmSecret    string `json:"lastfm-secret"`
	CertPath        string `json:"cert-path"`
	CertKeyPath     string `json:"cert-key-path"`
	HTTPPort        string `json:"http-port"`
	TLSPort         string `json:"tls-port"`
	Hostname        string `json:"hostname,omitempty"`
	AuthRedirectURL string `json:"auth-redirect-url"`
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
