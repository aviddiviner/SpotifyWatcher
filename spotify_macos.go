// +build darwin

package main

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
)

type State string

const (
	StateUnknown    State = ""
	StateForeground State = "foreground"
	StateStopped    State = "stopped"
	StatePlaying    State = "playing"
	StatePaused     State = "paused"
	StateClosing    State = "closing"
	StateClosed     State = "closed"
)

func (s State) String() string {
	if s == StateUnknown {
		return "(unknown)"
	}
	return string(s)
}

func osascript(script string) *exec.Cmd {
	var buf bytes.Buffer
	buf.WriteString(strings.TrimSpace(script))
	cmd := exec.Command("/usr/bin/osascript")
	cmd.Stdin = &buf
	return cmd
}

var checkStateScript = `
if application "Spotify" is frontmost then
	return "foreground"
end if
if application "Spotify" is running then
	tell application "Spotify"
		return player state as string  -- stopped/playing/paused
	end tell
else
	return "closed"
end if
`

func SpotifyState() (s State, err error) {
	out, err := osascript(checkStateScript).CombinedOutput()
	if err != nil {
		return
	}
	switch strings.TrimSpace(string(out)) {
	case "foreground":
		s = StateForeground
	case "stopped":
		s = StateStopped
	case "playing":
		s = StatePlaying
	case "paused":
		s = StatePaused
	case "closed":
		s = StateClosed
	default:
		err = errors.New("unknown state: bad output")
	}
	return
}

func TellSpotifyToQuit() error {
	return osascript(`tell application "Spotify" to quit`).Run()
}
