package main

import (
	"log"
	"strconv"
	"strings"
)

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
	if avgCpu > 25.0 {
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
	spotify := &tracker{avgCpu: NewMovingAvg(5)}
	top := NewTop(3)
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
