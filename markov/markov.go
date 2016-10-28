// Takes the tracks played and creates a Markov chain to use

package markov

import (
	"sort"
	"math/rand"
	"time"
	"fmt"
	"math"

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
 * AFTER TAKING THE Log10 OF IT
 */

func BuildChain(songs []lastFm.Song) map[string]Suffixes {
	// A prefix length of 1 is used (for now, it makes it super easy)
	chain := make(map[string]Suffixes, len(songs))
	for i, song := range songs {
		if i != len(songs)-1 {
			suffixes, exists := chain[song.Title] // try and get the suffixes
			if exists {
				nextSong := songs[i+1]
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
			} else {
				suffix := Suffix{Name: songs[i+1].Title, Artist: songs[i+1].Artist, Frequency: 1}
				chain[song.Title] = Suffixes{Suffixes: append(make([]Suffix, 0), suffix)}
			}
		}

	}
	// Make byte array
	/*fo, err := os.Create("output.txt")
	if err != nil {
		panic(err)
	}
	bf := bufio.NewWriter(fo)
	defer func() { // defer close on exit
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()
	for key, suffixes := range chain {
		line := "Prefix: " + key + " with " + strconv.Itoa(suffixes.Total) + " suffixes.\n"
		for _, suffix := range suffixes.Suffixes {
			line += ("  " + suffix.Name + ": " + strconv.Itoa(suffix.Frequency) + "\n")
		}

		_, err := bf.WriteString(line)
		if err != nil {
			panic(err)
		}
	}*/
	return chain
}

func GenerateSongList(length int, startingSong lastFm.Song, chain map[string]Suffixes) []lastFm.Song {
	currentPrefix := startingSong.Title
	list := make([]lastFm.Song, 0, length)
	list = append(list, startingSong)
	for i := 0; i < length; i++ {
		_, exists := chain[currentPrefix]
		if exists == true {
			if len(chain[currentPrefix].Suffixes) > 1 {
				suffixes := chain[currentPrefix].Suffixes
				cdf := make(CDF, 0, len(suffixes)) // cumulative distribution array with index 0 as the value, 1 as the Suffix index
				for j, suffix := range suffixes {
					var freq int
					if suffix.Name == currentPrefix {
						freq = int(math.Log10(float64(suffix.Frequency)) * repeatDiscount) // float arithmetic with a truncation
						if freq == 0 {
							freq += 1 // don't want to make repeats impossible
						}
					}
					cdf = append(cdf, [2]int {suffix.Frequency, j})
				}
				sort.Sort(cdf)
				// Creating the cdf here
				for j := 1; j < len(cdf); j++ {
					cdf[j][0] = cdf[j-1][0]
				}
				// Now to do the search
				suffix := suffixes[searchCDF(cdf)]
				name := suffix.Name
				artist := suffix.Artist

				currentPrefix = name
				fmt.Println("I chose to add", name)
				list = append(list, lastFm.Song{Artist: artist, Title: name})
			} else {
				name := chain[currentPrefix].Suffixes[0].Name
				artist := chain[currentPrefix].Suffixes[0].Artist
				fmt.Println("Only one choice. I chose to add", name)
				list = append(list, lastFm.Song{Artist: artist, Title: name})
				currentPrefix = name

			}
		} else {
			panic("Couldn't find the prefix") // TODO: fail silently or find another way around not being able to find the prefix
		}
	}
	return list
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