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
  -f --force    Monitor CPU even if Spotify is the frontmost (active) window.
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
	// Check state: foreground, background (playing/paused/etc).
	state, err := SpotifyState()
	if err != nil {
		return err
	}
	// Active in the foreground; ignore, unless forceful.
	if state == StateForeground && !opts.forceful {
		fmt.Printf(">>> Observed Spotify: foreground (ignored), CPU: %.2f\n", cpu)
		return nil
	}

	t.avgCpu.Append(cpu)
	samples := t.avgCpu.Samples()
	average := t.avgCpu.Value()
	fmt.Printf(">>> Observed Spotify: %s, CPU: %.2f (%.2f avg, %d samples)\n", state, cpu, average, samples)

	// Take action if we have sufficient samples.
	if samples == opts.avgWindow {
		// Too busy; kill.
		if average > opts.busyThreshold {
			return t.kill(p)
		}
		if state == StatePlaying || state == StateForeground {
			return nil
		}
		// In the background, not playing, but idling high; kill.
		if average > opts.idleThreshold {
			return t.kill(p)
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
