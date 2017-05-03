package configRead

import "testing"

const path = "testConfig.json"
const failPath = "failConfig.json"
const incorrectJSONPath = "badConfig.json"

// TestReadConfig attempts to read a valid configuration.
func TestReadConfig(t *testing.T) {
	config, err := Read(path)
	if err != nil {
		t.Errorf("ReadConfig returned an error: %s", err)
	}
	if len(config.LastFmKey) == 0 {
		t.Error("LastFmKey was not read.")
	}
	if len(config.LastFmSecret) == 0 {
		t.Error("LastFmSecret was not read.")
	}
	if len(config.SpotifyKey) == 0 {
		t.Error("SpotifyKey was not read.")
	}
	if len(config.SpotifySecret) == 0 {
		t.Error("SpotifySecret was not read.")
	}
	if len(config.HTTPPort) == 0 {
		t.Error("HTTPPort was not read.")
	}
	if len(config.Hostname) == 0 {
		t.Error("Hostname was not read.")
	}
	if len(config.AuthRedirectURL) == 0 {
		t.Error("AuthRedirectURL was not read.")
	}
}

// TestFailOpenConfig tests correct failure if the path to a config file is invalid.
func TestFailOpenConfig(t *testing.T) {
	_, err := Read(failPath)
	if err == nil {
		t.Error("No error was returned")
	}

}

// TestFailParseConfig tests correct failure if the config file could not be parsed.
func TestFailParseConfig(t *testing.T) {
	_, err := Read(incorrectJSONPath)
	if err == nil {
		t.Error("No error was returned")
	}
}
