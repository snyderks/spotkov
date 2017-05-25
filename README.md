[![Build Status](https://travis-ci.org/snyderks/spotkov.svg?branch=master)](https://travis-ci.org/snyderks/spotkov)
# Spotkov
A command line tool to generate a Spotify playlist of songs you might like together using your Last.FM scrobbling history and a Markov chain.

`$GOPATH/bin/spotkov -help`

## Installation
Clone the repository and install Go or if you have Go already, enter `go get github.com/snyderks/spotkov`.
You'll also need API keys for the [Spotify](https://developer.spotify.com/web-api/) and [Last.FM](http://www.last.fm/api/account/create) web API. 
Set environment variables (on UNIX, refer to [this article](http://www.cyberciti.biz/faq/set-environment-variable-linux/) (the process is identical on macOS), on Windows, edit your *user* environment variables at Control Panel -> System -> Advanced System Settings -> Environment Variables...)
 - SPOTIFY_ID: your Spotify API key
 - SPOTIFY_SECRET: your Spotify API secret key
 - LASTFM_KEY: your Last.FM API key
 - LASTFM_SECRET: (optional) your Last.FM secret key (currently not used by the application)

 *None* of the above environment variables are stored by the application. The application makes no network requests to servers other than Last.fm and Spotify.
 
This tool works best with Redis for caching, but isn't necessary. A local instance listening on `localhost:6379` is the default, but the environment variable `REDIS_URL` (same name as what Heroku uses) can also be set.

If you'd rather use a config file, check out the `configRead` package for the format.

---
### Why would this give good recommendations?
Beyond the idea of how [Markov chains](https://en.wikipedia.org/wiki/Markov_chain) [work](http://setosa.io/ev/markov-chains/), Spotkov using the Last.FM scrobber list as training is effective because the scrobble only takes place around [50% through the song](https://community.spotify.com/t5/Other-Partners-Windows-Phone-etc/How-long-do-you-have-to-listen-for-a-song-to-count-as-quot/td-p/978261), which means that songs you skipped or couldn't get through probably shouldn't be recommended, and since Last.FM doesn't store them Spotkov won't see it as options.

Spotkov doesn't repeat any song in its playlist and prevents multiple songs by the same artist from appearing in a row.

### How can I get it to give better recommendations?
Try to listen with purpose. If you throw playlists on shuffle and never skip, Spotkov will most likely give you more of the same (which you might want!). Even a few skips helps enormously with determining what songs you like together.
