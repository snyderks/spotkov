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

type Suffixes struct {
	Suffixes []Suffix
	Total    int // total number of Frequencies
}

type Suffix struct {
	Name      string
	Artist    string // for more accurate lookup in Spotify
	Frequency int    // number of times the suffix happens
}

type CDF [][2]int

const maxDeletionAttempts = 200

func BuildChain(songs []lastFm.Song) map[string]Suffixes {
	// A prefix length of 1 is used (for now, it makes it super easy)
	chain := make(map[string]Suffixes, len(songs))
	for i, song := range songs {
		if i != len(songs)-1 {
			suffixes, exists := chain[song.Title] // try and get the suffixes
			if exists {
				nextSong := songs[i+1]
				if nextSong.Title != song.Title || nextSong.Artist != song.Artist { // don't want to add duplicates
					timeSplit := song.Timestamp.Sub(nextSong.Timestamp)
					if timeSplit < time.Hour {
						var found bool = false
						for i, suffix := range suffixes.Suffixes {
							if suffix.Name == nextSong.Title {
								suffixes.Suffixes[i].Frequency += 1
								found = true
								break
							}
						}
						if found == false {
							suffixes.Suffixes = append(suffixes.Suffixes,
								Suffix{Name: nextSong.Title, Artist: nextSong.Artist, Frequency: 1})
						}
						suffixes.Total += 1
						chain[song.Title] = suffixes
					}
				}
			} else {
				suffix := Suffix{Name: songs[i+1].Title, Artist: songs[i+1].Artist, Frequency: 1}
				chain[song.Title] = Suffixes{Suffixes: append(make([]Suffix, 0), suffix)}
			}
		}

	}
	return chain
}

func GenerateSongList(length int, startingSong lastFm.Song, chain map[string]Suffixes) ([]lastFm.Song, error) {
	repeats := 0 // count of the number of repeats in a row
	deletionAttempts := 0
	clearAttempts := false
	foundSuffix := false
	var genError error
	list := make([]lastFm.Song, 0, length)
	list = append(list, startingSong)
	// Basic length loop
	for i := 0; i < length-1; i++ {
		foundSuffix = false
		/*
		 * Start at the end of the list and use that as the prefix.
		 * Try it and if it doesn't work, keep going back to the start.
		 * If we reach the start of the list and it still can't find a suffix,
		 * kill the loop and return what was found.
		 */
		for j := i; j >= 0 && foundSuffix == false; j-- {
			song, err := selectSuffix(chain, list[j].Title, &repeats)
			if err == nil {
				list = append(list, song)
				foundSuffix = true
			} else {
				return list, err
			}
			/*
			 * Checking for repeated sequences here. If 2+ songs appear in
			 * the same order, remove them, go back to the spot before that,
			 * and try again. If more than max attempts were made, cut short the playlist.
			 */
			new_list, deleted, err := findDuplicateSequences(&list)
			if err == nil {
				list = new_list
			}
			if deleted {
				// override the loop index because any deleted songs can't
				// be included in the length. Have to go back 2 due to the increment.
				i = len(list) - 2
				deletionAttempts++
				clearAttempts = false
			} else if deleted == false {
				/*
				 * Sometimes there are cycles that the generation gets in where it will
				 * add duplicates, go to another sequence that also contains duplicates,
				 * and repeat that. Requiring two successful additions should help. May
				 * have to revise.
				 */
				if clearAttempts {
					deletionAttempts = 0
				}
				clearAttempts = true
			}
		}
		if deletionAttempts == maxDeletionAttempts {
			genError = errors.New("An error occurred in generating your playlist. Please try again.")
			break
		}

	}
	return list, genError
}

func findDuplicateSequences(originalList *[]lastFm.Song) (list []lastFm.Song, deleting bool, err error) {
	list = *originalList
	err = nil
	defer func() {
		if r := recover(); r != nil {
			deleting = false
			err = errors.New("Something went wrong in deletion or seeking through the list.")
		}
	}()
	indicesToDelete := make([]int, 0)
	for i, song := range list { // walk forward through the songs.
		found := false

		var j int
		if len(indicesToDelete) > 0 {
			j = indicesToDelete[len(indicesToDelete)-1] + 1
			// The loop goes backwards. No need to start from the end.
			// Checking the next one first is plenty.
		} else {
			j = len(list) - 1 // Start from the end all normal-like
		}
		for ; j >= 0; j-- { // check in reverse
			// Make sure the indices aren't equal. Ugh.
			if j != i && list[j].Title == song.Title && list[j].Artist == song.Artist {
				found = true
				indicesToDelete = append(indicesToDelete, j)
				break
				// There's a check for the index being 0. This catches an issue where there's a duplicate
				// sequence in the middle of the list; it'll find the first one, start from the end, and
				// hit this else if. It should run through the whole list.
			} else if len(indicesToDelete) == 1 && j == 0 {
				indicesToDelete = make([]int, 0)
			}
		}
		// This check is to determine if it's found multiple duplicates in a row
		// and to make sure either the last one is at the end (can't continue) or
		// if it hit the end of the sequence.
		if len(indicesToDelete) > 1 && (found == false || indicesToDelete[len(indicesToDelete)-1] == len(list)-1) {
			deleting = true
			break
		}
	}
	if deleting {
		deleted := 0
		for _, index := range indicesToDelete {
			index -= deleted
			if index < len(list)-1 {
				list = append(list[:index], list[index+1:]...)
			} else {
				list = list[:index]
			}
			deleted++

		}
	}
	return list, deleting, err

}

func selectSuffix(chain map[string]Suffixes, prefix string, repeats *int) (lastFm.Song, error) {
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
	if exists == true {
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
