package main

import (
	"bufio"
	"bytes"
	"log"
	"os/exec"
	"strings"
)

type top struct {
	scanner *bufio.Scanner
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

func (t *top) run() {
	cmd := exec.Command("top", "-l", "1", "-stats", "pid,command,cpu,th,pstate,time,pageins")
	out := new(bytes.Buffer)
	cmd.Stdout = out
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
	t.scanner = bufio.NewScanner(out)

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
	// PID    COMMAND          %CPU #TH   STATE    TIME     PAGEINS
	// 99701  gosublime.margo_ 0.0  13    sleeping 02:54.09 3695+
	// 99156  printtool        0.0  2     sleeping 00:00.55 190+
	// 83615  Google Chrome He 0.0  14    sleeping 03:06.73 6089+
	// 80917  Google Chrome He 0.0  10    sleeping 00:10.14 1+
}

func (t *top) nextLine() (line string) {
	if t.scanner.Scan() {
		line = t.scanner.Text()
	}
	if err := t.scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return
}

func (t *top) nextFields() []string {
	return strings.Fields(t.nextLine())
}

func (t *top) chompHeader() {
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

func (t *top) scanAll() (results []Process) {
	t.nextFields() // []string{"PID", "COMMAND", "%CPU", "#TH", "STATE", "TIME", "PAGEINS"}
	for t.scanner.Scan() {
		fields := strings.Fields(t.scanner.Text())
		l := len(fields) - 1
		entry := Process{
			Pageins: fields[l],
			Time:    fields[l-1],
			State:   fields[l-2],
			Threads: fields[l-3],
			Cpu:     fields[l-4],
			// command name may have been split on space into many fields
			Command: strings.Join(fields[1:l-4], " "),
			Pid:     fields[0],
		}
		results = append(results, entry)
	}
	if err := t.scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return results
}

func ProcessList() []Process {
	t := new(top)
	t.run()
	t.chompHeader()
	return t.scanAll()
}
