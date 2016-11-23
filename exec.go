package main

import (
	"bytes"
	"io"
	"log"
	"os/exec"
	"time"
)

type Tick struct{}

type IdleCmd struct {
	// Buffered stdout.
	BufStdout bytes.Buffer
	// Idle sends a Tick whenever the command output goes idle.
	Idle chan Tick
}

// RunIdleCmd runs an os/exec Command, buffering its output. If it goes idle, and
// stops printing to Stdout after a given duration, then a message is sent on the
// Idle channel. The timer is reset when writes resume.
// There is no guarantee that the buffer won't be written to at any time.
func RunIdleCmd(cmd *exec.Cmd, idleAfter time.Duration) *IdleCmd {
	c := &IdleCmd{Idle: make(chan Tick)}
	go c.startAndWait(cmd, idleAfter)
	return c
}

func (c *IdleCmd) startAndWait(cmd *exec.Cmd, idleAfter time.Duration) {
	r, w := IdlePipe(idleAfter, c.Idle)
	cmd.Stdout = w
	// Buffer stdout.
	go func() {
		if _, err := io.Copy(&c.BufStdout, r); err != nil {
			log.Fatal(err)
		}
	}()
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}

// -----------------------------------------------------------------------------

type timeoutWriter struct {
	writer    *io.PipeWriter
	heartbeat chan Tick
}

func (t *timeoutWriter) Write(p []byte) (n int, err error) {
	select {
	case t.heartbeat <- Tick{}:
	default:
		// Don't block on sending heartbeats.
	}
	return t.writer.Write(p)
}

func (t *timeoutWriter) Close() error {
	return t.writer.Close()
}

func (t *timeoutWriter) notifyOnIdle(d time.Duration, idle chan Tick) {
	stalled := false
	for {
		if stalled {
			select {
			case <-t.heartbeat:
				stalled = false
			}
		} else {
			select {
			case <-t.heartbeat:
			case <-time.After(d):
				stalled = true
				select {
				case idle <- Tick{}:
				default:
					// Don't block on sending idle notifications either.
				}
			}
		}
	}
}

// IdlePipe creates a synchronous in-memory pipe, the same as io.Pipe, with the
// additional feature that if nothing is written after a specified duration, then
// a message is sent over the channel to notify that the pipeline has stalled.
// When writes resume, then the idle timer is reset.
func IdlePipe(d time.Duration, idle chan Tick) (io.ReadCloser, io.WriteCloser) {
	r, pw := io.Pipe()
	tw := &timeoutWriter{writer: pw, heartbeat: make(chan Tick)}
	go tw.notifyOnIdle(d, idle)
	return r, tw
}
