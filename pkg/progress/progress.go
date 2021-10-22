// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

// Package progress prints the program progress on screen. It's similar to a logger, but
// with better formatting.
package progress

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
)

// Log messages as new Events
var (
	evError   = "[✗] %s"
	evRunning = "[ ] %v"
	evRunOk   = "\r[✓]"
	evRunFail = "\r[✗]"
)

const spinners = `/-\|`

const (
	colorReset = "\033[0m"

	colorRed = "\033[31m"
	// colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	// colorPurple = "\033[35m"
	colorCyan = "\033[36m"
	// colorWhite  = "\033[37m"
)

type Progress struct {
	out     io.Writer // destination for output, usually os.Stderr
	running []byte
	seq     int // sequence of spinners
	debug   bool
}

var p *Progress

func init() {
	p = &Progress{out: os.Stderr, debug: false}
}

func SetDebug(on bool) {
	p.debug = on
}

func Cursor(show bool) {
	if p.out != os.Stdout && p.out != os.Stderr {
		return
	}
	if show {
		output([]byte("\033[?25h")) // Show cursor
	} else {
		output([]byte("\033[?25l")) // Hide cursor
	}
}

func Status(format string, a ...interface{}) {
	if len(p.running) > 0 {
		clearLine()
		output([]byte(colorCyan))
	}

	outputln("[>] " + fmt.Sprintf(format, a...))

	if len(p.running) > 0 {
		output([]byte(colorReset))
		output(p.running)
	}
}

func Error(err error) {
	if len(p.running) > 0 {
		clearLine()
	}

	output([]byte(colorRed))
	outputln(fmt.Sprintf(evError, err))
	output([]byte(colorReset))

	if len(p.running) > 0 {
		output(p.running)
	}
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func ErrorStack(err error) {
	if len(p.running) > 0 {
		clearLine()
	}

	output([]byte(colorRed))

	outputln(fmt.Sprintf(evError, err))
	if err, ok := err.(stackTracer); ok {
		for _, x := range err.StackTrace() {
			outputln(fmt.Sprintf("    |- %#v", x))
		}
	}

	output([]byte(colorReset))

	if len(p.running) > 0 {
		output(p.running)
	}
}

func ErrorMsg(format string, a ...interface{}) {
	if len(p.running) > 0 {
		clearLine()
	}

	output([]byte(colorRed))
	outputln("[✗] " + fmt.Sprintf(format, a...))
	output([]byte(colorReset))

	if len(p.running) > 0 {
		output(p.running)
	}

}

func Warning(format string, a ...interface{}) {
	if len(p.running) > 0 {
		clearLine()
	}

	output([]byte(colorYellow))
	outputln("[!] " + fmt.Sprintf(format, a...))
	output([]byte(colorReset))

	if len(p.running) > 0 {
		output(p.running)
	}

}

func Debug(format string, a ...interface{}) {
	if !p.debug {
		return
	}
	if len(p.running) > 0 {
		clearLine()
	}

	output([]byte(colorBlue))
	outputln("--- " + fmt.Sprintf(format, a...))
	output([]byte(colorReset))

	if len(p.running) > 0 {
		output(p.running)
	}
}

func Running(msg string) {
	p.running = []byte(fmt.Sprintf(evRunning, msg))
	output(p.running)
}

func Spinner() {
	output([]byte{'\r', '[', spinners[p.seq], ']'})
	p.seq = (p.seq + 1) % len(spinners)
}

func RunOK() {
	outputln(evRunOk)
	p.running = p.running[:0]
}

func RunFail() {
	output([]byte(colorRed))

	if len(p.running) > 0 {
		clearLine()
		output(p.running)
	}

	outputln(evRunFail)
	p.running = p.running[:0]

	output([]byte(colorReset))
}

func Download(a string) {
	p.running = []byte(fmt.Sprintf("[          ] %s", a))
	output(p.running)
}

/* ------- output ------- */

func clearLine() {
	if len(p.running) == 0 {
		return
	}
	buf := bytes.Repeat([]byte(" "), len(p.running)+2)
	buf[0] = byte('\r')
	buf[len(buf)-1] = byte('\r')
	_, _ = p.out.Write(buf)
}

func output(buf []byte) {
	_, _ = p.out.Write(buf)
}

func outputln(s string) {
	if p.out == nil {
		return
	}
	var buf []byte
	buf = append(buf, s...)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		buf = append(buf, '\n')
	}
	output(buf)
}
