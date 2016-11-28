package main

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os/exec"
	"time"
)

type Tick struct{}

type IdleCmd struct {
	Stdout io.Reader
	buf    *bufio.Writer

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

func (c *IdleCmd) Buffer() *bytes.Buffer {
	var bb bytes.Buffer

	// Copy exactly what we've buffered.
	go func(n int) {
		if _, err := io.CopyN(&bb, c.Stdout, int64(n)); err != nil {
			log.Fatal(err)
		}
	}(c.buf.Buffered())

	// Flush blocks on the underlying writer (the write end of the pipe).
	if err := c.buf.Flush(); err != nil {
		log.Fatal(err)
	}

	return &bb
}

func (c *IdleCmd) startAndWait(cmd *exec.Cmd, idleAfter time.Duration) {
	r, w := io.Pipe()
	c.Stdout = r
	c.buf = bufio.NewWriterSize(w, 4069*1024)

	// Buffer command output.
	cmd.Stdout = IdleWriter(c.buf, idleAfter, c.Idle)

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}

// -----------------------------------------------------------------------------

type idleWriter struct {
	writer    io.Writer
	heartbeat chan Tick
}

func (w *idleWriter) Write(p []byte) (n int, err error) {
	w.heartbeat <- Tick{}
	return w.writer.Write(p)
}

func (w *idleWriter) notifyOnIdle(d time.Duration, idle chan Tick) {
	stalled := false
	for {
		if stalled {
			<-w.heartbeat // Wait on heartbeat.
			stalled = false
		} else {
			select {
			case <-w.heartbeat:
			case <-time.After(d):
				stalled = true
				select {
				case idle <- Tick{}:
				default:
					// Don't block on sending idle notifications.
				}
			}
		}
	}
}

// IdleWriter wraps an io.Writer so that if nothing is written after a specified
// duration, then a message is sent over the channel to notify that it has gone idle.
// When writes resume, then the idle timer is reset.
func IdleWriter(w io.Writer, d time.Duration, idle chan Tick) io.Writer {
	iw := &idleWriter{writer: w, heartbeat: make(chan Tick)}
	go iw.notifyOnIdle(d, idle)
	return iw
}

// IdlePipe creates a synchronous in-memory pipe, the same as io.Pipe, with the
// additional feature that if nothing is written after a specified duration, then
// a message is sent over the channel to notify that the pipeline has stalled.
// When writes resume, then the idle timer is reset.
func IdlePipe(d time.Duration, idle chan Tick) (*io.PipeReader, io.Writer) {
	r, w := io.Pipe()
	return r, IdleWriter(w, d, idle)
}
