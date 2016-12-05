// +build darwin

package main

import (
	"errors"
	"os/exec"
	"strings"
)

type State int

func (s State) String() string {
	switch s {
	case StateActive:
		return "active"
	case StateStopped:
		return "stopped"
	case StatePlaying:
		return "playing"
	case StatePaused:
		return "paused"
	case StateClosed:
		return "closed"
	}
	return "unknown"
}

const (
	StateUnknown State = iota
	StateActive
	StateStopped
	StatePlaying
	StatePaused
	StateClosed
)

func SpotifyState() (s State, err error) {
	out, err := exec.Command("./SpotifyState.applescript").CombinedOutput()
	if err != nil {
		return
	}
	switch strings.TrimSpace(string(out)) {
	case "active":
		s = StateActive
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
