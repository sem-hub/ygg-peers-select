package mlog

import (
	"fmt"
	"os"
)

const (
	DEBUG = iota
	INFO
	WARN
	ERROR
)

type Logger struct {
	level int
}

var (
	logHandler Logger
)

func GetLogger() *Logger {
	return &logHandler
}

func (log *Logger) SetLevel(level int) {
	log.level = level
}

func (log *Logger) Fatal(msg string) {
	log.Error(msg)
	os.Exit(1)
}

func (log *Logger) Debug(msg string) {
	if log.level == DEBUG {
		fmt.Println(msg)
	}
}

func (log *Logger) Info(msg string) {
	if log.level <= INFO {
		fmt.Println(msg)
	}
}

func (log *Logger) Warning(msg string) {
	if log.level <= WARN {
		fmt.Println(msg)
	}
}

func (log *Logger) Error(msg string) {
	if log.level <= ERROR {
		fmt.Println(msg)
	}
}
