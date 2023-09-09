package log

import (
	"log"
	"sync"
)

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
}

type Logger interface {
	Fatal(v ...any)
	Error(v ...any)
	Warn(v ...any)
	Info(v ...any)
	Debug(v ...any)

	Level(l Level) Logger
}

// DefaultLogger defaults to noopLogger.
var DefaultLogger Logger = &noopLogger{}
var mut sync.Mutex

func SetDefaultLogger(l Logger) {
	mut.Lock()
	defer mut.Unlock()
	DefaultLogger = l
}

func SetLevel(l Level) {
	mut.Lock()
	defer mut.Unlock()
	DefaultLogger = DefaultLogger.Level(l)
}

// SetStdLogger use std 's log package.
// Default is InfoLevel.
func SetStdLogger() {
	SetDefaultLogger(&stdLogger{l: InfoLevel})
}

type stdLogger struct {
	l Level
}

func (s *stdLogger) Level(l Level) Logger {
	nl := *s
	nl.l = l
	return &nl
}

func (s *stdLogger) Fatal(v ...any) {
	if !s.l.Allow(FatalLevel) {
		return
	}
	log.Fatalln(v)
}

func (s *stdLogger) Error(v ...any) {
	if !s.l.Allow(ErrorLevel) {
		return
	}
	log.SetPrefix("error: ")
	log.Println(v)
}

func (s *stdLogger) Info(v ...any) {
	if !s.l.Allow(InfoLevel) {
		return
	}
	log.SetPrefix("info: ")
	log.Println(v)
}

func (s *stdLogger) Warn(v ...any) {
	if !s.l.Allow(WarnLevel) {
		return
	}
	log.SetPrefix("warn: ")
	log.Println(v)
}

func (s *stdLogger) Debug(v ...any) {
	if !s.l.Allow(DebugLevel) {
		return
	}
	log.SetPrefix("debug: ")
	log.Println(v)
}

func Info(v ...any) {
	DefaultLogger.Info(v)
}

func Error(v ...any) {
	DefaultLogger.Error(v)
}

func Warn(v ...any) {
	DefaultLogger.Warn(v)
}

func Debug(v ...any) {
	DefaultLogger.Debug(v)
}

func Fatal(v ...any) {
	DefaultLogger.Fatal(v)
}

var _ Logger = &stdLogger{}

type noopLogger struct{}

func (e *noopLogger) Level(l Level) Logger {
	return e
}
func (e *noopLogger) Fatal(v ...any) {}
func (e *noopLogger) Error(v ...any) {}
func (e *noopLogger) Warn(v ...any)  {}
func (e *noopLogger) Info(v ...any)  {}
func (e *noopLogger) Debug(v ...any) {}

var _ Logger = &noopLogger{}

type Level int

const (
	FatalLevel Level = iota
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
)

func (l Level) Allow(level Level) bool {
	return l >= level
}
