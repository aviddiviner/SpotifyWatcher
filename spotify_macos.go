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
	StateUnknown    State = "unknown"
	StateForeground State = "foreground"
	StateStopped    State = "stopped"
	StatePlaying    State = "playing"
	StatePaused     State = "paused"
	StateClosing    State = "closing"
	StateClosed     State = "closed"
)

func SpotifyState() (s State, err error) {
	out, err := exec.Command("./SpotifyState.applescript").CombinedOutput()
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
	var script bytes.Buffer
	script.WriteString(`tell application "Spotify" to quit`)
	cmd := exec.Command("/usr/bin/osascript")
	cmd.Stdin = &script
	return cmd.Run()
}
