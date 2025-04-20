package logger

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const days = 30
const hours = 24
const hours_on_month = days * hours

const default_level = 0
const default_prefix = "LOG: "
const default_size = 1024 * 1024 * 10 // 10MB
const default_rotation_time = time.Duration(hours_on_month * time.Hour)
const default_compress = false
const default_compress_suffix = ".gz"
const (
	FlagDate         = 1 << iota // Include the date in the log (e.g., 2009/01/23)
	FlagTime                     // Include the time in the log (e.g., 01:23:23)
	FlagMicroseconds             // Include microsecond resolution (e.g., 01:23:23.123123)
	FlagLongFile                 // Include full file path and line number (e.g., /a/b/c/d.go:23)
	FlagShortFile                // Include short file name and line number (e.g., d.go:23)
	FlagUTC                      // Use UTC instead of local time
	FlagMsgPrefix                // Include a message prefix
)
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

type Logger struct {
	logger         *log.Logger
	flag           int
	prefix         string
	outputFile     *os.File
	rotationTime   time.Duration
	rotationSize   int64
	compress       bool
	compressSuffix string
	useColors      bool
}

var loggerInstance *Logger = nil
var once sync.Once

type Options struct {
	Flag           int           // log Flag
	Prefix         string        // log prefix
	RotationTime   time.Duration // log rotation time
	RotationSize   int64         // log rotation size
	FilePath       string        // log output file
	Compress       bool          // log compression
	CompressSuffix string        // log compression suffix
}

func DefaultLogger() *Logger {
	once.Do(initLogger)
	return loggerInstance
}

func initLogger() {
	var logger = log.Default()
	loggerInstance = &Logger{
		logger:         logger,
		flag:           default_level,
		prefix:         default_prefix,
		outputFile:     nil,
		rotationTime:   default_rotation_time,
		rotationSize:   default_size,
		compress:       default_compress,
		compressSuffix: default_compress_suffix,
		useColors:      true,
	}

}

func LoggerWithOptions(options Options) *Logger {
	var logger *log.Logger = log.Default()
	flags := options.Flag
	if flags == 0 {
		flags = default_level
	}
	logger.SetFlags(flags)

	prefix := options.Prefix
	if prefix == "" {
		prefix = default_prefix
	}
	logger.SetPrefix(prefix)

	rotationTime := options.RotationTime
	if rotationTime == 0 {
		rotationTime = default_rotation_time
	}

	rotationSize := options.RotationSize
	if rotationSize == 0 {
		rotationSize = default_size
	}

	compressSuffix := options.CompressSuffix
	if compressSuffix == "" {
		compressSuffix = default_compress_suffix
	}
	var outputFile, err = os.OpenFile(options.FilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer outputFile.Close()
	if err != nil {
		panic(err)
	}

	useColors := outputFile == nil

	if outputFile != nil {
		logger.SetOutput(outputFile)
	}

	var newLoggerWithOptions = Logger{
		logger:         logger,
		flag:           flags,
		prefix:         prefix,
		outputFile:     outputFile,
		rotationTime:   rotationTime,
		rotationSize:   rotationSize,
		compress:       options.Compress,
		compressSuffix: compressSuffix,
		useColors:      useColors,
	}

	return &newLoggerWithOptions
}
func (l *Logger) String() string {
	var builder strings.Builder

	builder.WriteString("Logger{level: ")
	builder.WriteString(strconv.Itoa(l.flag))
	builder.WriteString(", prefix: ")
	builder.WriteString(l.prefix)
	builder.WriteString(", rotationTime: ")
	builder.WriteString(l.rotationTime.String())
	builder.WriteString(", rotationSize: ")
	builder.WriteString(strconv.FormatInt(l.rotationSize, 10))
	builder.WriteString(", compress: ")
	builder.WriteString(strconv.FormatBool(l.compress))
	builder.WriteString(", compressSuffix: ")
	builder.WriteString(l.compressSuffix)
	builder.WriteString("}")

	return builder.String()
}

func (l Logger) Panic(s string) {
	l.logger.Panicln(s)
}
func (l Logger) Printl(s string) {
	l.logger.Println(s)
}
func (l Logger) Printf(s string, v ...any) {
	l.logger.Printf(s, v...)
}
func (l Logger) Info(s string) {
	l.logger.SetPrefix(prefixString(l.useColors, "Info: ", colorGreen))
	l.logger.Output(2, s)
	l.logger.SetPrefix(l.prefix)
}

func (l Logger) InfoF(s string, v ...any) {
	l.logger.SetPrefix(prefixString(l.useColors, "Info: ", colorGreen))
	l.logger.Output(2, fmt.Sprintf(s, v...))
	l.logger.SetPrefix(l.prefix)
}
func (l Logger) Warn(s string) {
	l.logger.SetPrefix(prefixString(l.useColors, "Warning: ", colorYellow))
	l.logger.Output(2, s)
	l.logger.SetPrefix(l.prefix)
}

func (l Logger) WarnF(s string, v ...any) {
	l.logger.SetPrefix(prefixString(l.useColors, "Warning: ", colorYellow))
	l.logger.Output(2, fmt.Sprintf(s, v...))
	l.logger.SetPrefix(l.prefix)

}
func (l Logger) Fatal(s string) {
	l.logger.SetPrefix(prefixString(l.useColors, "Fatal: ", colorRed))
	l.logger.Output(2, s)
	os.Exit(1)
	l.logger.SetPrefix(l.prefix)

}
func (l Logger) FatalF(s string, v ...any) {
	l.logger.SetPrefix(prefixString(l.useColors, "Fatal: ", colorRed))
	l.logger.Output(2, fmt.Sprintf(s, v...))
	os.Exit(1)
	l.logger.SetPrefix(l.prefix)

}

func prefixString(uc bool, prefix string, color string) string {
	var builder strings.Builder
	if uc {
		builder.WriteString(color)
	}

	builder.WriteString(prefix)
	if uc {
		builder.WriteString(colorReset)
	}
	return builder.String()

}
