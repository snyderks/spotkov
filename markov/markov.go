// Takes the tracks played and creates a Markov chain to use

package markov

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/snyderks/spotkov/lastFm"
	"github.com/snyderks/spotkov/utils"
)

// Suffixes holds all suffixes for a specific song
type Suffixes struct {
	Suffixes []Suffix
	Total    int // total number of Frequencies
}

// Suffix holds a song that occurs after another song.
// Multiple suffixes with the same name can be duplicated
// across multiple source songs.
type Suffix struct {
	Name      string
	Artist    string // for more accurate lookup in Spotify
	Frequency int    // number of times the suffix happens
}

// CDF is a structure for a continuous distribution function,
// generated from the chain.
type CDF [][2]int

const maxAttempts = 200

// BuildChain determines what songs are played after others and creates a
// chain to then randomly select from.
// Takes an array of songs and returns a map.
func BuildChain(songs []lastFm.Song) map[string]Suffixes {
	// A prefix length of 1 is used (for now, it makes it super easy to get subsequent songs)
	chain := make(map[string]Suffixes, len(songs))
	// Creating suffixes, so the last song played doesn't have any yet.
	for i, song := range songs[:len(songs)-1] {
		// try and get the suffixes
		suffixes, exists := chain[song.Title]
		if exists {
			nextSong := songs[i+1]
			// don't want to add duplicates
			if nextSong.Title != song.Title || nextSong.Artist != song.Artist {
				timeSplit := song.Timestamp.Sub(nextSong.Timestamp)
				if timeSplit < time.Hour {
					found := false
					for i, suffix := range suffixes.Suffixes {
						if suffix.Name == nextSong.Title {
							suffixes.Suffixes[i].Frequency++
							found = true
							break
						}
					}
					if !found {
						suffixes.Suffixes = append(suffixes.Suffixes,
							Suffix{Name: nextSong.Title, Artist: nextSong.Artist, Frequency: 1})
					}
					suffixes.Total += 1
					chain[song.Title] = suffixes
				}
			}
		} else {
			suffix := Suffix{
				Name:      songs[i+1].Title,
				Artist:    songs[i+1].Artist,
				Frequency: 1,
			}
			chain[song.Title] = Suffixes{
				Suffixes: append(make([]Suffix, 0), suffix),
			}
		}
	}
	return chain
}

// GenerateSongList takes a seed song, a chain to select from, a length, and the maximum songs by one artist in a row.
// It returns a list of songs and an optional error.
func GenerateSongList(length int, maxBySameArtist int, startingSong lastFm.Song, chain map[string]Suffixes) ([]lastFm.Song, error) {
	foundSuffix := false
	var genError error
	list := make([]lastFm.Song, 0, length)
	list = append(list, startingSong)
	// Basic length loop
	for i := 0; i < length-1; i++ {
		foundSuffix = false
		// Start at the end of the list and use that as the prefix.
		// Try it and if it doesn't work, keep going back to the start.
		// If we reach the start of the list and it still can't find a suffix,
		// kill the loop and return what was found.
		for j := i; j >= 0 && foundSuffix == false; j-- {
			attempts := 0
			for attempts < maxAttempts {
				song, err := selectSuffix(chain, list[j].Title)
				if err == nil {
					// do not add the song if it's already in the list.
					isDupe := false
					for _, s := range list {
						// this is considered a match
						if s.Title == song.Title && s.Artist == song.Artist {
							isDupe = true
							break
						}
					}
					// if there are maxBySameArtist songs previously added by the same artist,
					// don't add this one.
					isRepeatArtist := false
					if len(list) > 1 && !isDupe {
						// start at the end
						checked := 0
						repeats := 0
						for checked < maxBySameArtist {
							if list[i-checked].Artist == song.Artist {
								repeats++
								if repeats >= maxBySameArtist {
									isRepeatArtist = true
									break
								}
							}
							checked++
						}
					}
					if !isDupe && !isRepeatArtist {
						list = append(list, song)
						foundSuffix = true
						break
					} else {
						attempts++
					}
				} else {
					return list, err
				}
			}
		}
		if !foundSuffix {
			genError = errors.New("An error occurred in generating your playlist. Please try again.")
			break
		}
	}
	return list, genError
}

func selectSuffix(chain map[string]Suffixes, prefix string) (lastFm.Song, error) {
	prefix = strings.Map(func(r rune) rune {
		if unicode.IsPunct(r) == true {
			return -1
		}
		return r
	}, strings.ToLower(prefix))
	exists := false
	for key := range chain {
		fmtPrefix := utils.LowerAndStripPunct(prefix)
		fmtKey := utils.LowerAndStripPunct(key)
		if fmtKey == fmtPrefix || strings.HasPrefix(fmtKey, fmtPrefix) {
			exists = true
			// It might be slightly different in the chain. This will allow it to continue if it is.
			prefix = key
			break
		}
	}
	song := lastFm.Song{}
	if exists {
		if len(chain[prefix].Suffixes) > 1 {
			suffixes := chain[prefix].Suffixes
			cdf := make(CDF, 0, len(suffixes)) // cumulative distribution array with index 0 as the value, 1 as the Suffix index
			for j, suffix := range suffixes {
				freq := suffix.Frequency

				if freq > 0 {
					cdf = append(cdf, [2]int{freq, j})
				}
			}

			sort.Sort(cdf) // making the CDF is much easier with sorting first.

			// Creating the cdf here
			for j := 1; j < len(cdf); j++ {
				cdf[j][0] = cdf[j-1][0]
			}
			// Now to do the search
			suffix := suffixes[searchCDF(cdf)]
			name := suffix.Name
			artist := suffix.Artist
			fmt.Println("I chose to add", name)
			song = lastFm.Song{Artist: artist, Title: name}
		} else { // there's only one choice.
			name := chain[prefix].Suffixes[0].Name
			artist := chain[prefix].Suffixes[0].Artist
			fmt.Println("Only one choice. I chose to add", name)
			song = lastFm.Song{Artist: artist, Title: name}
		}
		return song, nil
	}
	return lastFm.Song{}, errors.New("The song you entered couldn't be found. Please try again.")
}

// Sort interface implementation
func (cdf CDF) Len() int {
	return len(cdf)
}

func (cdf CDF) Less(i, j int) bool {
	return cdf[i][0] < cdf[j][0]
}

func (cdf CDF) Swap(i, j int) {
	temp := cdf[i]
	cdf[i] = cdf[j]
	cdf[j] = temp
}

// Binary search
func searchCDF(cdf CDF) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	num := r.Intn(len(cdf)-1) + 1 // Doing the -1 and +1 because Intn can return 0, which isn't valid
	// Binary search!
	right := len(cdf) - 1
	left := 0
	done := false
	index := -1
	for done == false {
		m := (left + right) / 2
		am := cdf[m][0]
		if am == num {
			index = m
			done = true
		} else if am < num {
			if m == 0 || m == len(cdf)-1 {
				index = m
				done = true
			} else if cdf[m+1][0] > num {
				index = m + 1
				done = true
			} else {
				left = m + 1
			}
		} else {
			if m == 0 || m == len(cdf)-1 {
				index = m
				done = true
			} else if cdf[m-1][0] < num {
				index = m
				done = true
			} else {
				right = m - 1
			}
		}
	}
	if done == false || index < 0 || index > len(cdf)-1 {
		panic("Something went wrong in the binary search")
	}
	return cdf[index][1]
}
