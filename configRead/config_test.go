package configRead

import (
	"fmt"
	"testing"
)

const path = "testConfig.json"
const failPath = "failConfig.json"
const incorrectJsonPath = "badConfig.json"

func TestReadConfig(t *testing.T) {
	config, err := ReadConfig(path)
	if err != nil {
		t.Error(fmt.Sprintf("ReadConfig returned an error: %s", err))
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
	if len(config.CertPath) == 0 {
		t.Error("CertPath was not read.")
	}
	if len(config.CertKeyPath) == 0 {
		t.Error("CertKeyPath was not read.")
	}
	if len(config.HTTPPort) == 0 {
		t.Error("HTTPPort was not read.")
	}
	if len(config.TLSPort) == 0 {
		t.Error("TLSPort was not read.")
	}
	if len(config.Hostname) == 0 {
		t.Error("Hostname was not read.")
	}
}

func TestFailOpenConfig(t *testing.T) {
	_, err := ReadConfig(failPath)
	if err == nil {
		t.Error("No error was returned")
	}

}

func TestFailParseConfig(t *testing.T) {
	_, err := ReadConfig(incorrectJsonPath)
	if err == nil {
		t.Error("No error was returned")
	}
}
