package jsonlog

import (
	"encoding/json"
	"io"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

// Level is custom type for the severity level
type Level int8

// Enum of different levels of severity
const (
	LevelInfo Level = iota
	LevelError
	LevelFatal
	LevelOff
)

// String method makes it human readible
func (level Level) ToString() string {
	switch level {
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return ""
	}
}

// Logger is a custom struct to hold the output destination, where the logs will be written to,
// minLevel, the minimum level of severity to log
// mu to coordinate writes
type Logger struct {
	out      io.Writer
	minLevel Level
	mu       sync.Mutex
}

func New(out io.Writer, minLevel Level) *Logger {
	return &Logger{
		out:      out,
		minLevel: minLevel,
	}
}

func (logger *Logger) PrintInfo(message string, properties map[string]string) {
	logger.print(LevelInfo, message, properties)
}

func (logger *Logger) PrintError(err error, properties map[string]string) {
	logger.print(LevelError, err.Error(), properties)
}

func (logger *Logger) PrintFatal(err error, properties map[string]string) {
	logger.print(LevelFatal, err.Error(), properties)
	// stop the program incase of fatal error
	os.Exit(1)
}

// print is an internal method on the Logger struct to do the actual logging
func (logger *Logger) print(
	level Level, message string, properties map[string]string,
) (int, error) {
	// dont do anything if the level is less than the minimum level of severity to log
	if level < logger.minLevel {
		return -1, nil
	}

	// create an aux struct to hold the data to be marshalled and logged
	aux := struct {
		Level      string            `json:"level"`
		Message    string            `json:"message"`
		Properties map[string]string `json:"properties,omitempty"`
		Trace      string            `json:"trace,omitempty"`
		Time       string            `json:"time"`
	}{
		Level:      level.ToString(),
		Message:    message,
		Properties: properties,
		Time:       time.Now().UTC().Format(time.RFC3339),
	}

	// include the error stuck trace if its above or equal to the error level of severity
	if level >= LevelError {
		aux.Trace = string(debug.Stack())
	}

	// declare a line var which which will hold the data to be logged
	var line []byte

	// marshal the data
	line, err := json.Marshal(aux)
	// incase of error, log the following entry
	if err != nil {
		line = []byte(LevelError.ToString() + ": failed to marshal log entry" + err.Error())
	}

	// lock the mutext to prevent concurrent writting, if not done multiple log entries might be
	// intermangled
	logger.mu.Lock()
	defer logger.mu.Unlock()

	// write to the output destination
	return logger.out.Write(append(line, '\n'))
}

// Write method is here so that Logger satifies the io.Writer interface
func (logger *Logger) Write(p []byte) (n int, err error) {
	return logger.print(LevelError, string(p), nil)
}
