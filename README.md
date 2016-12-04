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
  SpotifyWatcher [-t SECONDS] [-i CPU] [-p CPU] [-s SAMPLES]
  SpotifyWatcher -h | --help | --version

Options:
  -t SECONDS    Interval in secs with which to poll 'top' [default: 3].
  -i CPU        Idle CPU threshold at which to kill Spotify [default: 8.0].
  -p CPU        Playback CPU threshold at which to kill Spotify [default: 25.0].
  -s SAMPLES    Moving average sample window size [default: 5].
  -h --help     Show this screen.
  --version     Show version.
```
