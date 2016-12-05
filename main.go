package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aviddiviner/docopt-go"
)

var usage = `Monitor Spotify background CPU usage and kill it if it misbehaves.

Usage:
  SpotifyWatcher [-t SECONDS] [-i CPU] [-p CPU] [-s SAMPLES] [-f] [-v]
  SpotifyWatcher -h | --help | --version

Options:
  -t SECONDS    Interval in secs with which to poll 'top' [default: 3].
  -i CPU        Idle CPU threshold at which to kill Spotify [default: 8.0].
  -p CPU        Playback CPU threshold at which to kill Spotify [default: 25.0].
  -s SAMPLES    Moving average sample window size [default: 5].
  -f --force    Kill Spotify even if it's the frontmost (active) window.
  -v --verbose  Show details of all matching Spotify processes each tick.
  -h --help     Show this screen.
  --version     Show version.`

type options struct {
	topInterval   int
	idleThreshold float64
	busyThreshold float64
	avgWindow     int
	forceful      bool
	verbose       bool
}

var opts = options{}

type tracker struct {
	avgCpu *MovingAvg
}

func (t *tracker) kill(p Process) error {
	fmt.Println(">>> Killing Spotify!")
	pid, err := strconv.Atoi(p.Pid)
	if err != nil {
		return err
	}
	return kill(pid)
}

func (t *tracker) observe(p Process) error {
	if p == (Process{}) {
		// Nil process; reset the moving average and return.
		t.avgCpu.Reset()
		return nil
	}
	cpu, err := strconv.ParseFloat(p.Cpu, 64)
	if err != nil {
		return err
	}
	t.avgCpu.Append(cpu)
	if t.avgCpu.Samples() == opts.avgWindow {
		avgCpu := t.avgCpu.Value()
		fmt.Printf(">>> Spotify CPU: %.2f (avg: %.2f)\n", cpu, avgCpu)
		if avgCpu > opts.idleThreshold {
			// If we're forceful and over the high threshold, we're gonna die.
			if avgCpu > opts.busyThreshold && opts.forceful {
				return t.kill(p)
			}
			// Check if we're frontmost (active) or playing in the background.
			state, err := SpotifyState()
			if err != nil {
				return err
			}
			fmt.Printf(">>> Spotify State: %s\n", state)
			// Active in the foreground; leave alone, unless forceful.
			if state == StateActive && !opts.forceful {
				return nil
			}
			// Too busy in the background; kill.
			if avgCpu > opts.busyThreshold {
				return t.kill(p)
			}
			// Not playing in the background, but idling high; kill.
			if state != StatePlaying {
				return t.kill(p)
			}
		}
	}
	return nil
}

func main() {
	args, _ := docopt.ParseArgs(usage, nil, "0.2")
	if val, err := args.Int("-t"); err == nil {
		opts.topInterval = val
	}
	if val, err := args.Float64("-i"); err == nil {
		opts.idleThreshold = val
	}
	if val, err := args.Float64("-p"); err == nil {
		opts.busyThreshold = val
	}
	if val, err := args.Int("-s"); err == nil {
		opts.avgWindow = val
	}
	if val, err := args.Bool("--force"); err == nil {
		opts.forceful = val
	}
	if val, err := args.Bool("--verbose"); err == nil {
		opts.verbose = val
	}
	fmt.Printf("Starting with options: %+v\n", opts)

	tracker := &tracker{avgCpu: NewMovingAvg(opts.avgWindow)}
	top := NewTop(opts.topInterval)
	fmt.Printf("Collecting samples (~%d secs)...\n", opts.topInterval*opts.avgWindow)
	for {
		select {
		case <-top.NextTick:
			var spotify Process
			headerShown := false
			showProcessLine := func(p Process) {
				if !opts.verbose {
					return
				}
				if !headerShown {
					fmt.Printf("%-6s %-4s %-5s %-8s %-8s %-8s %s\n", "PID", "CPU", "#TH", "STATE", "TIME", "PAGEINS", "COMMAND")
					headerShown = true
				}
				fmt.Printf("%-6s %-4s %-5s %-8s %-8s %-8s %s\n", p.Pid, p.Cpu, p.Threads, p.State, p.Time, p.Pageins, p.Command)
			}
			for _, p := range top.ProcessList() {
				if strings.HasPrefix(p.Command, "Spotify") {
					showProcessLine(p)
					if p.Command == "Spotify" {
						spotify = p
					}
				}
			}
			if err := tracker.observe(spotify); err != nil {
				log.Fatal(err)
			}
		}
	}
}
