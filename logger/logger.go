package logger

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

const (
	DEFAULT   = 0
	DEBUG     = 100
	INFO      = 200
	NOTICE    = 300
	WARNING   = 400
	ERROR     = 500
	CRITICAL  = 600
	ALERT     = 700
	EMERGENCY = 800
)

type Logger interface {
	Debug(msg string)
}

// LogEntry represents a log entry fields and values
type LogEntry struct {
	// Message payload, optional
	Msg string `json:"msg"`
	// Defaults to Default (0)
	Severity int
	// Include http request values, optional
	HTTPRequest *http.Request
}

func Debug(msg string) {
	le := LogEntry{Msg: msg, Severity: DEBUG}
	le.print()
}

func Info(msg string) {
	le := LogEntry{Msg: msg, Severity: INFO}
	le.print()
}
func Notice(msg string) {
	le := LogEntry{Msg: msg, Severity: NOTICE}
	le.print()
}
func Warning(msg string) {
	le := LogEntry{Msg: msg, Severity: WARNING}
	le.print()
}
func Error(msg string) {
	le := LogEntry{Msg: msg, Severity: ERROR}
	le.print()
}
func Critical(msg string) {
	le := LogEntry{Msg: msg, Severity: CRITICAL}
	le.print()
}
func Alert(msg string) {
	le := LogEntry{Msg: msg, Severity: ALERT}
	le.print()
}
func Emergency(msg string) {
	le := LogEntry{Msg: msg, Severity: EMERGENCY}
	le.print()
}

func Info2(entry interface{}) {
	switch v := entry.(type) {
	case string:
		logEntry := &LogEntry{
			Msg:         v,
			Severity:    INFO,
			HTTPRequest: nil,
		}
		logEntry.print()
	case *LogEntry:
		v.Severity = INFO
		v.print()
	default:

	}
}

func Log(le *LogEntry) {
	le.print()
}

func (le *LogEntry) print() {
	output, err := json.Marshal(le)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Fprintf(os.Stdout, "%v\n", string(output))
}
