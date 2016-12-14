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
  --version     Show version.`

type options struct {
	TopInterval     int     `docopt:"-s"`
	CpuThreshold    float64 `docopt:"-t"`
	WindowLength    int     `docopt:"-w"`
	AllowedBreaches int     `docopt:"-n"`
	Quiet           bool
	Force           bool
	Verbose         bool
}

var opts options

type tracker struct {
	avgCpu   *FloatWindow
	breaches int
}

func newTracker() *tracker {
	return &tracker{avgCpu: NewFloatWindow(opts.WindowLength)}
}

func (t *tracker) Kill(p Process) error {
	if t.breaches < opts.AllowedBreaches {
		log.Println("Spotify is misbehaving!")
		t.breaches += 1
		return nil
	}
	log.Println("Okay, that's enough now. Killing Spotify!")
	pid, err := strconv.Atoi(p.Pid)
	if err != nil {
		return err
	}
	t.breaches = 0
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
	if state == StateForeground && !opts.Force {
		if !opts.Quiet {
			log.Printf("Spotify: foreground (ignored), CPU: %.2f\n", cpu)
		}
		return nil
	}

	t.avgCpu.Append(cpu)
	samples := t.avgCpu.Len()
	median := t.avgCpu.Median()
	if !opts.Quiet {
		log.Printf("Spotify: %s, CPU: %.2f (%.2f median, samples: %d)\n", state, cpu, median, samples)
	}

	// Take action if we have sufficient samples.
	if samples == opts.WindowLength {
		// Too busy; kill.
		if median > opts.CpuThreshold {
			return t.Kill(p)
		}
	}
	return nil
}

func parseOptions(argv []string) (o options) {
	args, _ := docopt.ParseArgs(usage, argv, "0.3")
	err := args.Bind(&o)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func main() {
	opts = parseOptions(nil)
	log.Printf("Starting with options: %+v\n", opts)

	metrics := newInfluxAgent()
	tracker := newTracker()
	top := NewTop(opts.TopInterval)
	log.Println("Waiting to observe Spotify...")
	for {
		select {
		case <-top.NextTick:
			var spotify Process
			batch := metrics.NewBatch()
			addMetricPoint := func(p Process) {
				// TODO: Better Process struct, do this in parseTopLine() rather.
				pid, _ := strconv.Atoi(p.Pid)
				cpu, _ := strconv.ParseFloat(p.Cpu, 64)
				threads, _ := strconv.Atoi(p.Threads)
				pageins, _ := strconv.Atoi(p.Pageins)
				metrics.AddPoint(batch, "process",
					metricTags{
						"command": p.Command,
					},
					metricFields{
						"pid":     pid,
						"cpu":     cpu,
						"threads": threads,
						"state":   p.State,
						"time":    p.Time,
						"pageins": pageins,
						"command": p.Command,
					})
			}
			headerShown := false
			showProcessLine := func(p Process) {
				if !opts.Verbose {
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
