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
Monitor Spotify background CPU usage and kill it if it misbehaves.

Usage:
  SpotifyWatcher [-s SECONDS] [-t CPU] [-w LENGTH] [-n ALLOWED] [-f] [-q|-v]
  SpotifyWatcher -h | --help | --version

Options:
  -s SECONDS    Interval in secs with which to poll 'top' [default: 4].
  -t CPU        CPU threshold at which to kill Spotify [default: 8.0].
  -w LENGTH     Median sample window size [default: 5].
  -n ALLOWED    Max intervals exceeding threshold before killing [default: 20].
  -f --force    Monitor CPU even if Spotify is the frontmost (active) window.
  -q --quiet    Only output console message when Spotify is misbehaving.
  -v --verbose  Show details of all matching Spotify processes each tick.
  -h --help     Show this screen.
  --version     Show version.
```
