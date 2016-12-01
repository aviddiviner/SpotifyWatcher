# Spotify Watcher
Monitors the output from `top` periodically to see if Spotify is misbehaving, and then kills it unceremoniously.

## Build, install, run
```console
$ git clone https://github.com/aviddiviner/SpotifyWatcher.git
$ cd SpotifyWatcher
$ go build .
$ ./SpotifyWatcher
```

Currently only works for macOS.

## Usage
```console
$ ./SpotifyWatcher --help
Monitor Spotify CPU usage and kill it if it misbehaves.

Usage:
  SpotifyWatcher [-i INTERVAL] [-t THRESHOLD] [-w WINDOW]
  SpotifyWatcher -h | --help | --version

Options:
  -i INTERVAL   Interval in secs with which to poll 'top' [default: 3].
  -t THRESHOLD  CPU threshold at which to kill Spotify [default: 25.0].
  -w WINDOW     Moving average sample window size [default: 5].
  -h --help     Show this screen.
  --version     Show version.
```
