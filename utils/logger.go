package utils

import (
	"ai-commons/config"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

type contextKey string

const LoggerContextKey contextKey = "logger"

var logger *log.Logger

// GetLoggerFromContext retrieves the logger from the context.
func GetLoggerFromContext(ctx context.Context) (*log.Entry, error) {
	logger, ok := ctx.Value(LoggerContextKey).(*log.Entry)
	if !ok {
		return nil, fmt.Errorf("logger not found in context")
	}
	return logger, nil
}

// InitLogger initializes the global logger instance for the utils package.
// It sets up output, level, and formatter based on provided parameters.
// This function should ideally be called once at application startup.
func InitLogger() error {
	logger = log.New() // Initialize the logger instance

	// set multiple writers for logging output
	var writers []io.Writer

	// Set output: Try to open file, fallback to stdout
	logFilePath := config.GetConfig().Logging.File
	if logFilePath != "" {
		file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			writers = append(writers, file)
			// logging using standard fmt, as logger isn't fully configured yet
			fmt.Printf("Logging to file: %s\n", logFilePath)
		} else {
			fmt.Fprintf(os.Stderr, "WARNING: Failed to open log file %s: %v. Logging to stdout only.\n", logFilePath, err)
		}
	}

	logToStdout := *config.GetConfig().Logging.Stdout
	if logToStdout || len(writers) == 0 {
		// If logToStdout is true or no file writer was added, log to stdout
		writers = append(writers, os.Stdout)
		if logFilePath == "" {
			fmt.Println("Logging to stdout.")
		} else if logToStdout {
			fmt.Println("Logging to stdout and file.")
		}
	}

	// Create a MultiWriter to write to all configured destinations
	mw := io.MultiWriter(writers...)
	logger.SetOutput(mw)

	// Set the log level
	levelStr := config.GetConfig().Logging.Level
	level, err := log.ParseLevel(levelStr)
	if err != nil {
		return fmt.Errorf("invalid log level '%s': %w", levelStr, err)
	}

	isJSON := config.GetConfig().Logging.Json
	if *isJSON {
		fmt.Println("Logging in JSON format.")
	} else {
		fmt.Println("Logging in text format.")
	}

	// Set the log level
	logger.SetLevel(level)

	// Set the formatter
	if *isJSON {
		logger.SetFormatter(&log.JSONFormatter{
			TimestampFormat: time.RFC3339Nano, // High precision timestamp
		})
	} else {
		logger.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
			ForceColors:   true, // Force colors in terminal output
		})
	}

	// Add a default field for the application name
	logger.WithField("app", "ai-commons").Info("Logger initialized")

	return nil
}

// GetBaseLogger returns the base initialized logger.
func GetBaseLogger() *log.Logger {
	if logger == nil {
		fmt.Fprintf(os.Stderr, "WARNING: GetBaseLogger called before InitLogger. Initializing default logger.\n")
		tempLogger := log.New()
		tempLogger.SetOutput(os.Stderr)
		tempLogger.SetLevel(log.InfoLevel)
		tempLogger.SetFormatter(&log.TextFormatter{FullTimestamp: true})
		logger = tempLogger
	}
	return logger
}
