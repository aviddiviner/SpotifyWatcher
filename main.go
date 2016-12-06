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
  SpotifyWatcher [-s SECONDS] [-i CPU] [-p CPU] [-n SAMPLES] [-f] [-v]
  SpotifyWatcher -h | --help | --version

Options:
  -s SECONDS    Interval in secs with which to poll 'top' [default: 3].
  -i CPU        Idle CPU threshold at which to kill Spotify [default: 8.0].
  -p CPU        Playback CPU threshold at which to kill Spotify [default: 25.0].
  -n SAMPLES    Moving average sample window size [default: 5].
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

func newTracker() *tracker {
	return &tracker{avgCpu: NewMovingAvg(opts.avgWindow)}
}

func (t *tracker) Kill(p Process) error {
	fmt.Println(">>> Killing Spotify!")
	pid, err := strconv.Atoi(p.Pid)
	if err != nil {
		return err
	}
	return kill(pid)
}

func (t *tracker) Observe(p Process) error {
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
		fmt.Printf("Spotify: foreground (ignored), CPU: %.2f\n", cpu)
		return nil
	}

	t.avgCpu.Append(cpu)
	samples := t.avgCpu.Length()
	average := t.avgCpu.Average()
	fmt.Printf("Spotify: %s, CPU: %.2f (%.2f avg, samples: %d)\n", state, cpu, average, samples)

	// Take action if we have sufficient samples.
	if samples == opts.avgWindow {
		// Too busy; kill.
		if average > opts.busyThreshold {
			return t.Kill(p)
		}
		if state == StatePlaying || state == StateForeground {
			return nil
		}
		// In the background, not playing, but idling high; kill.
		if average > opts.idleThreshold {
			return t.Kill(p)
		}
	}
	return nil
}

func parseOptions(argv []string) (o options) {
	args, _ := docopt.ParseArgs(usage, argv, "0.2")
	if val, err := args.Int("-s"); err == nil {
		o.topInterval = val
	}
	if val, err := args.Float64("-i"); err == nil {
		o.idleThreshold = val
	}
	if val, err := args.Float64("-p"); err == nil {
		o.busyThreshold = val
	}
	if val, err := args.Int("-n"); err == nil {
		o.avgWindow = val
	}
	if val, err := args.Bool("--force"); err == nil {
		o.forceful = val
	}
	if val, err := args.Bool("--verbose"); err == nil {
		o.verbose = val
	}
	return
}

func main() {
	opts = parseOptions(nil)
	fmt.Printf("Starting with options: %+v\n", opts)

	metrics := newInfluxAgent()
	tracker := newTracker()
	top := NewTop(opts.topInterval)
	fmt.Println("Waiting to observe Spotify...")
	for {
		select {
		case <-top.NextTick:
			var spotify Process
			batch := metrics.NewBatch()
			addMetricPoint := func(p Process) {
				metrics.AddPoint(batch, "process",
					metricTags{
						"command": p.Command,
					},
					metricFields{
						"pid":     p.Pid,
						"cpu":     p.Cpu,
						"threads": p.Threads,
						"state":   p.State,
						"time":    p.Time,
						"pageins": p.Pageins,
						"command": p.Command,
					})
			}
			headerShown := false
			showProcessLine := func(p Process) {
				if !opts.verbose {
					return
				}
				if !headerShown {
					fmt.Printf("  %-6s %-4s %-5s %-8s %-8s %-8s %s\n", "PID", "CPU", "#TH", "STATE", "TIME", "PAGEINS", "COMMAND")
					headerShown = true
				}
				fmt.Printf("  %-6s %-4s %-5s %-8s %-8s %-8s %s\n", p.Pid, p.Cpu, p.Threads, p.State, p.Time, p.Pageins, p.Command)
			}
			for _, p := range top.ProcessList() {
				if strings.HasPrefix(p.Command, "Spotify") {
					showProcessLine(p)
					addMetricPoint(p)
					if p.Command == "Spotify" {
						spotify = p
					}
				}
			}
			if err := metrics.Write(batch); err != nil {
				// log.Fatal(err)
			}
			if err := tracker.Observe(spotify); err != nil {
				log.Fatal(err)
			}
		}
	}
}
