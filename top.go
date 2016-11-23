package main

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Tick struct{}

type Top struct {
	interval int

	mutex  sync.RWMutex
	buffer bytes.Buffer
	idle   chan Tick

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

func NewTop(interval int) *Top {
	t := &Top{interval: interval, NextTick: make(chan Tick)}
	// log.Println("starting top: taking lock")
	t.mutex.Lock()
	// log.Println("starting top: taken lock")
	go t.run()
	return t
}

func (t *Top) ProcessList() []Process {
	return t.results
}

func (t *Top) run() {
	// Processes: 306 total, 2 running, 2 stuck, 302 sleeping, 1772 threads
	// 2016/11/20 20:18:55
	// Load Avg: 1.36, 1.41, 1.35
	// CPU usage: 3.70% user, 22.22% sys, 74.7% idle
	// SharedLibs: 150M resident, 19M data, 15M linkedit.
	// MemRegions: 83717 total, 3073M resident, 71M private, 868M shared.
	// PhysMem: 8688M used (3048M wired), 7694M unused.
	// VM: 2471G vsize, 533M framework vsize, 15758266(0) swapins, 17238545(0) swapouts.
	// Networks: packets: 26102141/14G in, 21138143/6128M out.
	// Disks: 6676021/171G read, 6960487/301G written.
	//
	// PID    %CPU #TH   STATE    TIME     PAGEINS  COMMAND
	// 99701  0.0  13    sleeping 02:54.09 3695+    gosublime.margo_
	// 99156  0.0  2     sleeping 00:00.55 190+     printtool
	// 83615  0.0  14    sleeping 03:06.73 6089+    Google Chrome He
	// 80917  0.0  10    sleeping 00:10.14 1+       Google Chrome He

	cmd := exec.Command("top", "-l", "0", "-s", strconv.Itoa(t.interval), "-stats", "pid,cpu,th,pstate,time,pageins,command")
	r, w, idle := timeoutPipe(400 * time.Millisecond)
	cmd.Stdout = w
	go func() {
		for {
			log.Println("run: starting copy to buffer")
			_, err := io.Copy(&t.buffer, r)
			if err != nil {
				log.Fatal(err)
			}
			log.Fatal("unexpected: run: copy to buffer ended")
		}
	}()
	t.idle = idle
	go t.watch()
	log.Println("run: starting command")
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	// log.Println("run: releasing lock")
	t.mutex.Unlock()
	// log.Println("run: waiting on command to finish")
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
	log.Fatal("unexpected: run: command ended")
}

func (t *Top) watch() {
	log.Println("watch: starting")
	counter := 0
	for {
		select {
		case <-t.idle:
			// log.Println("watch: idle, taking lock")
			t.mutex.RLock()
			// log.Println("watch: idle, taken lock")

			if counter > 0 {
				t.scanner = bufio.NewScanner(&t.buffer)
				t.results = t.scanResults()
			}
			t.buffer.Reset()

			// log.Println("watch: idle, releasing lock")
			t.mutex.RUnlock()
			counter += 1
		}
		if counter > 1 {
			select {
			case t.NextTick <- Tick{}:
				// log.Println("watch: sent next tick")
			default:
				// log.Println("watch: sent nothing")
			}
		}
	}
}

func (t *Top) nextLine() (line string) {
	if t.scanner.Scan() {
		line = t.scanner.Text()
	}
	if err := t.scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return
}

func (t *Top) nextFields() []string {
	return strings.Fields(t.nextLine())
}

func (t *Top) chompHeader() {
	t.nextLine() // "Processes: 306 total, 2 running, 2 stuck, 302 sleeping, 1772 threads"
	t.nextLine() // "2016/11/20 20:18:55"
	t.nextLine() // "Load Avg: 1.36, 1.41, 1.35"
	t.nextLine() // "CPU usage: 3.70% user, 22.22% sys, 74.7% idle"
	t.nextLine() // "SharedLibs: 150M resident, 19M data, 15M linkedit."
	t.nextLine() // "MemRegions: 83717 total, 3073M resident, 71M private, 868M shared."
	t.nextLine() // "PhysMem: 8688M used (3048M wired), 7694M unused."
	t.nextLine() // "VM: 2471G vsize, 533M framework vsize, 15758266(0) swapins, 17238545(0) swapouts."
	t.nextLine() // "Networks: packets: 26102141/14G in, 21138143/6128M out."
	t.nextLine() // "Disks: 6676021/171G read, 6960487/301G written."
	t.nextLine() // ""
}

var expectedHeaders = []string{"PID", "%CPU", "#TH", "STATE", "TIME", "PAGEINS", "COMMAND"}

func parseTopLine(line string) (p Process) {
	// top sometimes gives us junky output, like any of these:
	// "72846  0.0  1     sleeping00:00.02 86       postgres        "
	// "72846  0.0  1     sleeping0:00.02 86       postgres        "
	// "72846  0.0  1     sleeping0::00.02 86       postgres        "
	// "72846  0.0  1     sleeping0:00.02 86       postgres        "
	// "72846  0.0  1     sleeping0::00.02 86       postgres        "
	// "72846  0.0  1     sleeping00:00.02 86       postgres        "
	// "72846  0.0  1     sleeping000.02 86       postgres        "
	// "72846  0.0  1     sleeping00000.02 86       postgres        "
	// "72846  0.0  1     sleeping00:00.02 86       postgres        "
	defer func() {
		if err := recover(); err != nil {
			// log.Printf("error parsing line: %q\n", line)
		}
	}()

	fields := strings.Fields(line)
	l := len(fields)
	p = Process{
		Pid:     fields[0],
		Cpu:     fields[1],
		Threads: fields[2],
		State:   fields[3],
		Time:    fields[4],
		Pageins: fields[5],
		// command name may be split on space
		Command: strings.Join(fields[6:l], " "),
	}
	return
}

func (t *Top) scanResults() (results []Process) {
	t.chompHeader()
	if !reflect.DeepEqual(t.nextFields(), expectedHeaders) {
		log.Fatal("unexpected fields")
	}
	for t.scanner.Scan() {
		entry := parseTopLine(t.scanner.Text())
		results = append(results, entry)
	}
	if err := t.scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return results
}

// -----------------------------------------------------------------------------

type timeoutWriter struct {
	writer    *io.PipeWriter
	heartbeat chan Tick
}

func (t *timeoutWriter) Write(p []byte) (n int, err error) {
	// log.Println("tw: got a write")
	select {
	case t.heartbeat <- Tick{}:
		// log.Printf("tw: sent heartbeat")
	default:
		// log.Printf("tw: you don't care")
	}
	return t.writer.Write(p)
}

func (t *timeoutWriter) Close() error {
	return t.writer.Close()
}

func (t *timeoutWriter) notifyOnTimeout(d time.Duration, timeout chan Tick) {
	stalled := false
	for {
		if stalled {
			// log.Println("tw: waiting on heartbeat now")
			select {
			case <-t.heartbeat:
				// log.Println("tw: got a heartbeat (resurrected)")
				stalled = false
			}
		} else {
			select {
			case <-t.heartbeat:
				// log.Println("tw: got a heartbeat")
			case <-time.After(d):
				// log.Println("tw: stalled!")
				stalled = true
				select {
				case timeout <- Tick{}:
					// log.Println("tw: let you know :)")
				default:
					// log.Println("tw: you're not interested :(")
				}
			}
		}
	}
}

func timeoutPipe(d time.Duration) (io.ReadCloser, io.WriteCloser, chan Tick) {
	r, pw := io.Pipe()
	tw := &timeoutWriter{writer: pw, heartbeat: make(chan Tick)}
	timeout := make(chan Tick)
	go tw.notifyOnTimeout(d, timeout)
	return r, tw, timeout
}
