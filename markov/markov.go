// Takes the tracks played and creates a Markov chain to use

package markov

import (
	"bufio"
	"os"
	"strconv"
)

type Suffixes struct {
	Suffixes []Suffix
	Total    int // total number of Frequencies
}

type Suffix struct {
	Name      string
	Frequency int // number of times the suffix happens
}

const repeatDiscount = 0.5

/* the percentage of the Chance to discount
 * the suffix by if it's a repeat of the prefix
 */

func BuildChain(songs []string) map[string]Suffixes {
	// A prefix length of 1 is used (for now, it makes it super easy)
	chain := make(map[string]Suffixes, len(songs))
	for i, song := range songs {
		if i != len(songs)-1 {
			suffixes, exists := chain[song] // try and get the suffixes
			if exists {
				nextSong := songs[i+1]
				var found bool = false
				for i, suffix := range suffixes.Suffixes {
					if suffix.Name == nextSong {
						suffixes.Suffixes[i].Frequency += 1
						found = true
						break
					}
				}
				if found == false {
					suffixes.Suffixes = append(suffixes.Suffixes, Suffix{Name: nextSong, Frequency: 1})
				}
				suffixes.Total += 1
				chain[song] = suffixes
			} else {
				suffix := Suffix{Name: songs[i+1], Frequency: 1}
				chain[song] = Suffixes{Suffixes: append(make([]Suffix, 0), suffix)}
			}
		}

	}
	// Make byte array
	fo, err := os.Create("output.txt")
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
	}
	return chain
}

func GenerateSongList(length int, startingSong string, chain map[string]Suffixes) {
	currentPrefix := startingSong
	list := make([]string, 0, length)
	list = append(list, currentPrefix)
	for i := 0; i < length; i++ {
		suffixes, exists := chain[currentPrefix]
		if exists == true {

		} else {
			panic("Couldn't find the prefix") // TODO: fail silently or find another way around not being able to find the prefix
		}
	}
}
