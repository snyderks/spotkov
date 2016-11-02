// Takes the tracks played and creates a Markov chain to use

package markov

import (
	"sort"
	"math/rand"
	"time"
	"fmt"
	"math"
	"errors"

	"github.com/snyderks/spotkov/lastFm"
)

type Suffixes struct {
	Suffixes []Suffix
	Total    int // total number of Frequencies
}

type Suffix struct {
	Name      string
	Artist    string // for more accurate lookup in Spotify
	Frequency int // number of times the suffix happens
}

type CDF [][2]int

const repeatDiscount = 0.5
/* the percentage of the Chance to discount
 * the suffix to by if it's a repeat of the prefix
 * AFTER TAKING THE NATURAL LOG OF IT
 */

func BuildChain(songs []lastFm.Song) map[string]Suffixes {
	// A prefix length of 1 is used (for now, it makes it super easy)
	chain := make(map[string]Suffixes, len(songs))
	for i, song := range songs {
		if i != len(songs)-1 {
			suffixes, exists := chain[song.Title] // try and get the suffixes
			if exists {
				nextSong := songs[i+1]
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
			} else {
				suffix := Suffix{Name: songs[i+1].Title, Artist: songs[i+1].Artist, Frequency: 1}
				chain[song.Title] = Suffixes{Suffixes: append(make([]Suffix, 0), suffix)}
			}
		}

	}
	return chain
}

func GenerateSongList(length int, startingSong lastFm.Song, chain map[string]Suffixes) []lastFm.Song {
	repeats := 0 // count of the number of repeats in a row
	deletionAttempts := 0
	foundSuffix := false
	list := make([]lastFm.Song, 0, length)
	list = append(list, startingSong)
	// Basic length loop
	for i := 0; i < length; i++ {
		foundSuffix = false;
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
				foundSuffix = true;
			} else {
				fmt.Println(err)
			}
			/*
			 * Checking for repeated sequences here. If 2+ songs appear in
			 * the same order, remove them, go back to the spot before that,
			 * and try again. If more than 100 attempts were made, cut short the playlist.
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
			} else if deleted == false {
				deletionAttempts = 0
			}
		}
		if deletionAttempts == 100 {
			fmt.Println("Couldn't continue. Ending generation.")
			break
		}

	}
	return list
}

func adjustRepeatFrequency(baseFreq int, repeats int) int {
	return int(math.Log(float64(baseFreq)) * (repeatDiscount / float64(repeats)))
}

func findDuplicateSequences(originalList *[]lastFm.Song) (list []lastFm.Song, deleting bool, err error) {
	list = *originalList
	defer func() {
		if r := recover(); r != nil {
			deleting = false
			err = errors.New("Something went wrong in deletion or seeking through the list.")
		}
	}()
	indicesToDelete := make([]int, 0)
	for i, song := range list { // walk forward through the songs.
		found := false
		for j := len(list) - 1; j >= 0; j-- { // check starting from the end
			// Make sure the indices aren't equal. Ugh.
			if j != i && list[j].Title == song.Title && list[j].Artist == song.Artist {
				found = true
				indicesToDelete = append(indicesToDelete, j)
				break
			} else if len(indicesToDelete) == 1 {
				indicesToDelete = make([]int, 0)
			}
		}
		// This check is to determine if it's found multiple duplicates in a row
		// and to make sure either the last one is at the end (can't continue) or
		// if it hit the end of the sequence.
		if len(indicesToDelete) > 1 && (found == false || indicesToDelete[len(indicesToDelete) - 1] == len(list) - 1) {
			deleting = true
			break
		}
	}
	if deleting {
		fmt.Println("indices:", indicesToDelete)
		deleted := 0
		for _, index := range indicesToDelete {
			fmt.Println("deleting", index)
			index -= deleted
			fmt.Println("adjusted to", index)
			if index < len(list) - 1 {
				list = append(list[:index], list[index+1:]...)
			} else {
				list = list[:index]
			}
			deleted++
			
		}
	}
	return list, deleting, nil

}

func selectSuffix(chain map[string]Suffixes, prefix string, repeats *int) (lastFm.Song, error) {
	_, exists := chain[prefix]
	song := lastFm.Song{}
		if exists == true {
			if len(chain[prefix].Suffixes) > 1 {
				suffixes := chain[prefix].Suffixes
				cdf := make(CDF, 0, len(suffixes)) // cumulative distribution array with index 0 as the value, 1 as the Suffix index
				for j, suffix := range suffixes {
					var freq int
					if suffix.Name == prefix {
						*repeats = *repeats + 1
						freq = adjustRepeatFrequency(suffix.Frequency, *repeats)
					} else {
						freq = suffix.Frequency
						*repeats = 0 // wasn't a repeat, reset
					}
					if (freq > 0) {
						cdf = append(cdf, [2]int {freq, j})
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
		} else {
			return lastFm.Song{}, errors.New("Couldn't find the song for that prefix.")
		}
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
	num := r.Intn(len(cdf) - 1) + 1 // Doing the -1 and +1 because Intn can return 0, which isn't valid
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
			if m == 0  || m == len(cdf) - 1 {
				index = m
				done = true
			} else if cdf[m+1][0] > num {
				index = m+1
				done = true
			} else {
				left = m+1
			}
		} else {
			if m == 0 || m == len(cdf) - 1 {
				index = m
				done = true
			} else if cdf[m-1][0] < num {
				index = m
				done = true
			} else {
				right = m-1
			}
		}
	}
	if done == false || index < 0 || index > len(cdf) - 1 {
		panic("Something went wrong in the binary search")
	}
	return cdf[index][1]
}