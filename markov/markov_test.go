package markov

import (
  "testing"
  "fmt"

  "github.com/snyderks/spotkov/lastFm"
)

func buildSampleList() []lastFm.Song {
  length := 20
  list := make([]lastFm.Song, 0, length)
  for i := 0; i < length; i++ {
    var num int
    repeatLoc := 10 // used to generate a duplicate sequence
    // need nonzero check to avoid division by zero
    if i != 0 && i % repeatLoc > 0 && i % repeatLoc < 4 { 
      num = i % repeatLoc
    } else {
      num = i
    }
    title := fmt.Sprintf("SongTitle%d", num)
    artist := fmt.Sprintf("SongArtist%d", num)
    list = append(list, 
      lastFm.Song{ Title: title, Artist: artist })
  }
  return list
}

func TestDeletingDuplicateSequences(t *testing.T) {
  // This uses a list of length 20 and a duplicate sequence of length 3.
  // It should find it, delete it, and have a list of length 20 at the end.
  list := buildSampleList()
  fmt.Println(list)
  if len(list) != 20 {
    t.Error("Length of sample list was not 20.")
  }
  new_list, deleted, err := findDuplicateSequences(&list)
  if err != nil {
    t.Error("An error was caught")
  }
  if deleted != true {
    t.Error("Didn't delete items from the list.")
  }
  if len(new_list) != 17 {
    t.Error("Didn't delete the correct number from the list.")
  }
}