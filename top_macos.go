// +build darwin

package main

import (
	"bufio"
	"log"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func NewTop(interval int) *Top {
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
	cmd := exec.Command("top", "-l", "0", "-s", strconv.Itoa(interval), "-stats", "pid,cpu,th,pstate,time,pageins,command")
	top := &Top{cmd: RunIdleCmd(cmd, 400*time.Millisecond), NextTick: make(chan Tick)}
	go top.watch()
	return top
}

func (t *Top) ProcessList() []Process {
	return t.results
}

func (t *Top) watch() {
	counter := 0
	for {
		select {
		case <-t.cmd.Idle:
			if counter > 0 { // We go idle before the first batch of output is received
				t.scanner = bufio.NewScanner(&t.cmd.BufStdout)
				t.results = t.scanResults()
			}
			t.cmd.BufStdout.Reset()
			counter += 1
		}
		if counter > 1 { // macOS `top` has bullshit CPU results on the first tick
			select {
			case t.NextTick <- Tick{}:
			default:
				// Don't block on notifying about next tick.
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
