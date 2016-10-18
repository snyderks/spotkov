[![Build Status](https://travis-ci.org/snyderks/spotkov.svg?branch=master)](https://travis-ci.org/snyderks/spotkov)
# spotkov
A command line tool to generate a Spotify playlist of songs you might like together using your Last.FM scrobbling history and a Markov chain.

Install by cloning the repository and installing Go or with `go get github.com/snyderks/spotkov`.

---
## Why would this give good recommendations?
Beyond the idea of how [Markov chains](https://en.wikipedia.org/wiki/Markov_chain) [work](http://setosa.io/ev/markov-chains/), Spotkov using the Last.FM scrobber list as training is effective because the scrobble only takes place around [50% through the song](https://community.spotify.com/t5/Other-Partners-Windows-Phone-etc/How-long-do-you-have-to-listen-for-a-song-to-count-as-quot/td-p/978261), which means that songs you skipped or couldn't get through probably shouldn't be recommended, and since Last.FM doesn't store them Spotkov won't see it as options.

Also, there's a discounter for repeats of the same song that is designed to offer more variety. Occasionally (if you've listened to one song over and over for quite a bit of time) it will recommend a song twice. However, this isn't designed to happen very often.

## How can I get it to give better recommendations?
Try to listen with purpose. If you throw playlists on shuffle and never skip, Spotkov will most likely give you more of the same (which you might want!). Even a few skips helps enormously with determining what songs you like together.