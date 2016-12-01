package main

import (
	"log"
	"strconv"
	"strings"

	"github.com/aviddiviner/docopt-go"
)

var usage = `Monitor Spotify CPU usage and kill it if it misbehaves.

Usage:
  SpotifyWatcher [-i INTERVAL] [-t THRESHOLD] [-w WINDOW]
  SpotifyWatcher -h | --help | --version

Options:
  -i INTERVAL   Interval in secs with which to poll 'top' [default: 3].
  -t THRESHOLD  CPU threshold at which to kill Spotify [default: 25.0].
  -w WINDOW     Moving average sample window size [default: 5].
  -h --help     Show this screen.
  --version     Show version.`

type options struct {
	topInterval  int
	cpuThreshold float64
	avgWindow    int
}

var opts = options{}

type tracker struct {
	avgCpu *MovingAvg
}

func (t *tracker) observe(p Process) error {
	cpu, err := strconv.ParseFloat(p.Cpu, 64)
	if err != nil {
		return err
	}
	t.avgCpu.Append(cpu)
	avgCpu := t.avgCpu.Value()
	log.Printf(">>> Spotify CPU: %.2f (avg: %.2f)\n", cpu, avgCpu)
	if avgCpu > opts.cpuThreshold {
		log.Printf(">>> Killing Spotify!")
		pid, err := strconv.Atoi(p.Pid)
		if err != nil {
			return err
		}
		return kill(pid)
	}
	return nil
}

func main() {
	args, _ := docopt.ParseDoc(usage)
	if val, err := args.Int("-i"); err == nil {
		opts.topInterval = val
	}
	if val, err := args.Float64("-t"); err == nil {
		opts.cpuThreshold = val
	}
	if val, err := args.Int("-w"); err == nil {
		opts.avgWindow = val
	}
	log.Printf("Starting with options: %+v\n", opts)

	spotify := &tracker{avgCpu: NewMovingAvg(opts.avgWindow)}
	top := NewTop(opts.topInterval)
	log.Println("Waiting...")
	for {
		select {
		case <-top.NextTick:
			for _, p := range top.ProcessList() {
				if strings.HasPrefix(p.Command, "Spotify") {
					log.Printf("%#v\n", p)
					if p.Command == "Spotify" {
						if err := spotify.observe(p); err != nil {
							log.Fatal(err)
						}
					}
				}
			}
		}
	}
}
