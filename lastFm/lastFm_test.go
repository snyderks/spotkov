package lastFm

import "testing"

func TestSongMapCaching(t *testing.T) {
	// create a very basic test object
	s := SongMap{}
	s.Songs = make(map[BaseSong]bool)
	s.Songs[BaseSong{"title", "artist"}] = true

	// make a caching attempt
	err := cacheUniqueSongs("test", s)
	if err != nil {
		t.Error(err.Error())
		t.Error("Do you have a Redis test server running?")
	}
}

func TestSongMapReading(t *testing.T) {
	// create a very basic test object
	s := SongMap{}
	s.Songs = make(map[BaseSong]bool)
	s.Songs[BaseSong{"title", "artist"}] = true

	// make a caching attempt
	err := cacheUniqueSongs("test", s)
	if err != nil {
		t.Error(err.Error())
		t.Error("Do you have a Redis test server running?")
		return
	}

	// see if we can read it back
	s = SongMap{}
	err = ReadCachedUniqueSongs("test", &s)
	if err != nil {
		t.Error(err.Error())
	}
	_, ok := s.Songs[BaseSong{"title", "artist"}]
	if !ok {
		t.Error("The returned cache didn't contain the correct item.")
	}
}
