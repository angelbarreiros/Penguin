package logger

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	White  = "\033[37m"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
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
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

func (l LogLevel) Color() string {
	switch l {
	case DEBUG:
		return Cyan
	case INFO:
		return Green
	case WARN:
		return Yellow
	case ERROR:
		return Red
	case FATAL:
		return Purple
	default:
		return White
	}
}

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Fatal(msg string, args ...any)
	SetLevel(level LogLevel)
}

type FileLogger struct {
	mu            sync.Mutex
	file          *os.File
	level         LogLevel
	logDir        string
	baseFileName  string
	maxFileSizeMB int
	maxAgeDays    int
}

var fileLoggerInstance *FileLogger
var fileLoggerOnce sync.Once

type ConsoleLogger struct {
	mu      sync.Mutex
	level   LogLevel
	colored bool
}

var consoleLoggerInstance *ConsoleLogger
var consoleLoggerOnce sync.Once

func GetFileLogger() *FileLogger {
	fileLoggerOnce.Do(func() {
		fileLoggerInstance = &FileLogger{
			level:         INFO,
			logDir:        "logs",
			baseFileName:  "app.log",
			maxFileSizeMB: 10,
			maxAgeDays:    30,
		}

		if err := os.MkdirAll(fileLoggerInstance.logDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create log directory: %v\n", err)
		}

		fileLoggerInstance.openLogFile()
	})
	return fileLoggerInstance
}

func GetConsoleLogger() *ConsoleLogger {
	consoleLoggerOnce.Do(func() {
		consoleLoggerInstance = &ConsoleLogger{
			level:   INFO,
			colored: true,
		}
	})
	return consoleLoggerInstance
}

func (l *FileLogger) Configure(logDir string, baseFileName string, maxFileSizeMB int, maxAgeDays int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	var changed bool = false

	if logDir != "" && logDir != l.logDir {
		l.logDir = logDir
		changed = true
	}

	if baseFileName != "" && baseFileName != l.baseFileName {
		l.baseFileName = baseFileName
		changed = true
	}

	if maxFileSizeMB > 0 && maxFileSizeMB != l.maxFileSizeMB {
		l.maxFileSizeMB = maxFileSizeMB
	}

	if maxAgeDays > 0 && maxAgeDays != l.maxAgeDays {
		l.maxAgeDays = maxAgeDays
	}

	if changed {
		if l.file != nil {
			l.file.Close()
			l.file = nil
		}

		if err := os.MkdirAll(l.logDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create log directory: %v\n", err)
		}

		l.openLogFile()
	}
}

// SetLevel sets the minimum log level for the file logger
func (l *FileLogger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// SetLevel sets the minimum log level for the console logger
func (c *ConsoleLogger) SetLevel(level LogLevel) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.level = level
}

// EnableColors enables or disables colored output for the console logger
func (c *ConsoleLogger) EnableColors(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.colored = enabled
}

// Debug logs a debug message
func (l *FileLogger) Debug(msg string, args ...any) {
	_, file, line, _ := runtime.Caller(1)
	l.logWithCaller(DEBUG, file, line, msg, args...)
}

// Info logs an info message
func (l *FileLogger) Info(msg string, args ...any) {
	_, file, line, _ := runtime.Caller(1)
	l.logWithCaller(INFO, file, line, msg, args...)
}

// Warn logs a warning message
func (l *FileLogger) Warn(msg string, args ...any) {
	_, file, line, _ := runtime.Caller(1)
	l.logWithCaller(WARN, file, line, msg, args...)
}

// Error logs an error message
func (l *FileLogger) Error(msg string, args ...any) {
	_, file, line, _ := runtime.Caller(1)
	l.logWithCaller(ERROR, file, line, msg, args...)
}

// Fatal logs a fatal message and exits the program
func (l *FileLogger) Fatal(msg string, args ...any) {
	_, file, line, _ := runtime.Caller(1)
	l.logWithCaller(FATAL, file, line, msg, args...)
	os.Exit(1)
}

// Debug logs a debug message to console
func (c *ConsoleLogger) Debug(msg string, args ...any) {
	_, file, line, _ := runtime.Caller(1)
	c.logWithCaller(DEBUG, file, line, msg, args...)
}

// Info logs an info message to console
func (c *ConsoleLogger) Info(msg string, args ...any) {
	_, file, line, _ := runtime.Caller(1)
	c.logWithCaller(INFO, file, line, msg, args...)
}

// Warn logs a warning message to console
func (c *ConsoleLogger) Warn(msg string, args ...any) {
	_, file, line, _ := runtime.Caller(1)
	c.logWithCaller(WARN, file, line, msg, args...)
}

// Error logs an error message to console
func (c *ConsoleLogger) Error(msg string, args ...any) {
	_, file, line, _ := runtime.Caller(1)
	c.logWithCaller(ERROR, file, line, msg, args...)
}

// Fatal logs a fatal message to console and exits
func (c *ConsoleLogger) Fatal(msg string, args ...any) {
	_, file, line, _ := runtime.Caller(1)
	c.logWithCaller(FATAL, file, line, msg, args...)
	os.Exit(1)
}

func getShortFilePath(file string) string {
	var short string = filepath.Base(file)
	var dir string = filepath.Base(filepath.Dir(file))
	if dir != "." && dir != "/" {
		return dir + "/" + short
	}
	return short
}

func (l *FileLogger) logWithCaller(level LogLevel, file string, line int, msg string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < l.level {
		return
	}

	var timestamp string = time.Now().Format("2006-01-02 15:04:05.000")
	var shortFile string = getShortFilePath(file)
	var caller string = fmt.Sprintf("%s:%d", shortFile, line)

	var logMsg strings.Builder
	logMsg.WriteString("[")
	logMsg.WriteString(timestamp)
	logMsg.WriteString("] [")
	logMsg.WriteString(level.String())
	logMsg.WriteString("] [")
	logMsg.WriteString(caller)
	logMsg.WriteString("] ")

	if len(args) > 0 {
		logMsg.WriteString(fmt.Sprintf(msg, args...))
	} else {
		logMsg.WriteString(msg)
	}
	logMsg.WriteString("\n")

	if l.file == nil {
		l.openLogFile()
	}

	if l.file != nil {
		if _, err := l.file.WriteString(logMsg.String()); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write to log file: %v\n", err)
		}

		if fi, err := l.file.Stat(); err == nil {
			if fi.Size() > int64(l.maxFileSizeMB*1024*1024) {
				l.rotateLog()
			}
		}
	}
}

func (c *ConsoleLogger) logWithCaller(level LogLevel, file string, line int, msg string, args ...any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if level < c.level {
		return
	}

	var timestamp string = time.Now().Format("2006-01-02 15:04:05.000")
	var shortFile string = getShortFilePath(file)
	var caller string = fmt.Sprintf("%s:%d", shortFile, line)

	var logMsg strings.Builder
	logMsg.WriteString("[")
	logMsg.WriteString(timestamp)
	logMsg.WriteString("] [")
	if c.colored {
		logMsg.WriteString(level.Color())
	}
	logMsg.WriteString(level.String())
	if c.colored {
		logMsg.WriteString(Reset)
	}
	logMsg.WriteString("] [")
	logMsg.WriteString(caller)
	logMsg.WriteString("] ")

	if len(args) > 0 {
		logMsg.WriteString(fmt.Sprintf(msg, args...))
	} else {
		logMsg.WriteString(msg)
	}
	logMsg.WriteString("\n")

	fmt.Print(logMsg.String())
}

func (l *FileLogger) openLogFile() {
	if l.file != nil {
		l.file.Close()
		l.file = nil
	}
	logPath := filepath.Join(l.logDir, l.baseFileName)
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file: %v\n", err)
		return
	}

	l.file = file
}

func (l *FileLogger) rotateLog() {
	if l.file == nil {
		return
	}

	l.file.Close()
	l.file = nil

	var currentPath string = filepath.Join(l.logDir, l.baseFileName)
	var timestamp string = time.Now().Format("20060102-150405")
	var newPath string = filepath.Join(l.logDir, fmt.Sprintf("%s.%s", l.baseFileName, timestamp))

	if err := os.Rename(currentPath, newPath); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to rotate log file: %v\n", err)
	}

	go func(filePath string) {
		compressedPath := filePath + ".gz"
		err := compressFile(filePath, compressedPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to compress log file: %v\n", err)
			return
		}

		os.Remove(filePath)
	}(newPath)

	l.openLogFile()
}

func compressFile(src, dst string) error {

	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	var gzipWriter *gzip.Writer = gzip.NewWriter(destFile)
	defer gzipWriter.Close()

	_, err = io.Copy(gzipWriter, sourceFile)
	if err != nil {
		return err
	}

	return nil
}
