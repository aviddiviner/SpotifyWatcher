# Spotify Watcher
Monitors the output from `top` periodically to see if Spotify is misbehaving, and then kills it unceremoniously.

## Build, install, run
```bash
git clone https://github.com/aviddiviner/SpotifyWatcher.git
cd SpotifyWatcher
go build .
./SpotifyWatcher
```

Currently only works for macOS.
