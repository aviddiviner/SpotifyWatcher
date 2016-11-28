package main

import (
	"bufio"
	"os"
)

type Top struct {
	cmd     *IdleCmd
	scanner *bufio.Scanner
	results []Process

	// NextTick sends a Tick whenever new results are available.
	NextTick chan Tick
}

type Process struct {
	Pid     string
	Command string
	Cpu     string
	Threads string
	State   string
	Time    string
	Pageins string
}

func kill(pid int) (err error) {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return
	}
	return proc.Kill()
}
