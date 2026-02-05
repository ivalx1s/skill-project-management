package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogLevel represents logging severity
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger handles session-based logging for the TUI
type Logger struct {
	mu          sync.Mutex
	file        *os.File
	sessionID   string
	sessionStart time.Time
	logsDir     string
	minLevel    LogLevel
	actionCount int
}

// NewLogger creates a new logger instance
func NewLogger(logsDir string) *Logger {
	return &Logger{
		logsDir:  logsDir,
		minLevel: DEBUG, // Log everything by default
	}
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.minLevel = level
}

// Open starts a new logging session
func (l *Logger) Open() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Generate session ID from timestamp
	l.sessionStart = time.Now()
	l.sessionID = l.sessionStart.Format("2006-01-02_15-04-05")

	// Ensure logs directory exists
	if err := os.MkdirAll(l.logsDir, 0755); err != nil {
		return fmt.Errorf("failed to create logs dir: %w", err)
	}

	// Open log file
	logPath := filepath.Join(l.logsDir, fmt.Sprintf("session_%s.log", l.sessionID))
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	l.file = file

	// Write session header
	l.writeHeader()

	return nil
}

// Close ends the logging session
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file == nil {
		return nil
	}

	// Write session footer
	duration := time.Since(l.sessionStart)
	l.writeRaw(fmt.Sprintf("\n=== SESSION END ===\n"))
	l.writeRaw(fmt.Sprintf("Duration: %s\n", duration.Round(time.Second)))
	l.writeRaw(fmt.Sprintf("Total actions: %d\n", l.actionCount))
	l.writeRaw(fmt.Sprintf("=====================================\n"))

	err := l.file.Close()
	l.file = nil
	return err
}

// writeHeader writes the session start info
func (l *Logger) writeHeader() {
	cwd, _ := os.Getwd()
	l.writeRaw("=====================================\n")
	l.writeRaw(fmt.Sprintf("=== SESSION START: %s ===\n", l.sessionID))
	l.writeRaw(fmt.Sprintf("Time: %s\n", l.sessionStart.Format(time.RFC3339)))
	l.writeRaw(fmt.Sprintf("Working dir: %s\n", cwd))
	l.writeRaw("=====================================\n\n")
}

// writeRaw writes raw text without formatting
func (l *Logger) writeRaw(text string) {
	if l.file != nil {
		l.file.WriteString(text)
	}
}

// log writes a formatted log entry
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file == nil || level < l.minLevel {
		return
	}

	timestamp := time.Now().Format("15:04:05.000")
	msg := fmt.Sprintf(format, args...)
	entry := fmt.Sprintf("[%s] [%s] %s\n", timestamp, level, msg)
	l.file.WriteString(entry)
	l.actionCount++
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Action logs a user action (convenience method)
func (l *Logger) Action(action string, details string) {
	if details != "" {
		l.Info("[ACTION] %s: %s", action, details)
	} else {
		l.Info("[ACTION] %s", action)
	}
}

// Screen logs a screen transition
func (l *Logger) Screen(from, to string) {
	l.Info("[SCREEN] %s → %s", from, to)
}

// Command logs a command execution
func (l *Logger) Command(cmd, args, result string) {
	if args != "" {
		l.Info("[CMD] /%s %s → %s", cmd, args, result)
	} else {
		l.Info("[CMD] /%s → %s", cmd, result)
	}
}

// Key logs a key press (for debugging)
func (l *Logger) Key(key string) {
	l.Debug("[KEY] %s", key)
}

// State logs current state (for debugging)
func (l *Logger) State(format string, args ...interface{}) {
	l.Debug("[STATE] "+format, args...)
}

// CLIError logs a CLI command failure
func (l *Logger) CLIError(command string, err error, stderr string) {
	l.Error("[CLI] %s failed: %v", command, err)
	if stderr != "" {
		l.Error("[CLI] stderr: %s", stderr)
	}
}

// SessionID returns the current session ID
func (l *Logger) SessionID() string {
	return l.sessionID
}
