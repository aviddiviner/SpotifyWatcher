package main

import (
	"log"
	"strconv"
	"strings"

	"github.com/aviddiviner/docopt-go"
)

var usage = `Monitor Spotify CPU usage and kill it if it misbehaves.

Usage:
  SpotifyWatcher [-t SECONDS] [-i CPU] [-p CPU] [-s SAMPLES]
  SpotifyWatcher -h | --help | --version

Options:
  -t SECONDS    Interval in secs with which to poll 'top' [default: 3].
  -i CPU        Idle CPU threshold at which to kill Spotify [default: 8.0].
  -p CPU        Playback CPU threshold at which to kill Spotify [default: 25.0].
  -s SAMPLES    Moving average sample window size [default: 5].
  -h --help     Show this screen.
  --version     Show version.`

type options struct {
	topInterval   int
	idleThreshold float64
	busyThreshold float64
	avgWindow     int
}

var opts = options{}

type tracker struct {
	avgCpu *MovingAvg
}

func (t *tracker) kill(p Process) error {
	log.Printf(">>> Killing Spotify!")
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
		log.Printf(">>> Spotify CPU: %.2f (avg: %.2f)\n", cpu, avgCpu)
		if avgCpu > opts.busyThreshold {
			return t.kill(p)
		}
		if avgCpu > opts.idleThreshold {
			state, err := SpotifyState()
			if err != nil {
				return err
			}
			log.Printf(">>> Spotify State: %s\n", state)
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
	log.Printf("Starting with options: %+v\n", opts)

	tracker := &tracker{avgCpu: NewMovingAvg(opts.avgWindow)}
	top := NewTop(opts.topInterval)
	log.Println("Waiting...")
	for {
		select {
		case <-top.NextTick:
			var spotify Process
			for _, p := range top.ProcessList() {
				if strings.HasPrefix(p.Command, "Spotify") {
					log.Printf("%#v\n", p)
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
