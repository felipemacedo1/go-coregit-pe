package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

// Level represents log levels
type Level int

const (
	LevelInfo Level = iota
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger provides structured logging
type Logger struct {
	output io.Writer
	json   bool
	level  Level
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// NewLogger creates a new logger
func NewLogger(output io.Writer, jsonFormat bool) *Logger {
	if output == nil {
		output = os.Stderr
	}
	return &Logger{
		output: output,
		json:   jsonFormat,
		level:  LevelInfo,
	}
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level Level) {
	l.level = level
}

// Info logs an info message
func (l *Logger) Info(message string, fields ...map[string]interface{}) {
	l.log(LevelInfo, message, fields...)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, fields ...map[string]interface{}) {
	l.log(LevelWarn, message, fields...)
}

// Error logs an error message
func (l *Logger) Error(message string, fields ...map[string]interface{}) {
	l.log(LevelError, message, fields...)
}

// log writes a log entry
func (l *Logger) log(level Level, message string, fields ...map[string]interface{}) {
	if level < l.level {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level.String(),
		Message:   message,
	}

	if len(fields) > 0 {
		entry.Fields = fields[0]
	}

	if l.json {
		l.writeJSON(entry)
	} else {
		l.writeText(entry)
	}
}

// writeJSON writes log entry as JSON
func (l *Logger) writeJSON(entry LogEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Failed to marshal log entry: %v", err)
		return
	}
	fmt.Fprintln(l.output, string(data))
}

// writeText writes log entry as plain text
func (l *Logger) writeText(entry LogEntry) {
	output := fmt.Sprintf("[%s] %s: %s", entry.Timestamp, entry.Level, entry.Message)

	if entry.Fields != nil && len(entry.Fields) > 0 {
		output += " |"
		for k, v := range entry.Fields {
			output += fmt.Sprintf(" %s=%v", k, v)
		}
	}

	fmt.Fprintln(l.output, output)
}

// Default logger instance
var defaultLogger = NewLogger(os.Stderr, false)

// Info logs an info message using the default logger
func Info(message string, fields ...map[string]interface{}) {
	defaultLogger.Info(message, fields...)
}

// Warn logs a warning message using the default logger
func Warn(message string, fields ...map[string]interface{}) {
	defaultLogger.Warn(message, fields...)
}

// Error logs an error message using the default logger
func Error(message string, fields ...map[string]interface{}) {
	defaultLogger.Error(message, fields...)
}

// SetJSONFormat enables/disables JSON formatting for the default logger
func SetJSONFormat(enabled bool) {
	defaultLogger.json = enabled
}

// SetLevel sets the log level for the default logger
func SetLevel(level Level) {
	defaultLogger.SetLevel(level)
}
