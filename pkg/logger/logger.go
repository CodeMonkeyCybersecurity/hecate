// pkg/logger/logger.go
package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

var Log *zap.Logger

// DefaultConfig returns a standard zap.Config object with custom settings.
func DefaultConfig() zap.Config {
	level := zap.InfoLevel
	switch os.Getenv("LOG_LEVEL") {
	case "trace":
		level = zap.DebugLevel
	case "debug":
		level = zap.DebugLevel
	case "dpanic":
		level = zap.DPanicLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	case "fatal":
		level = zap.FatalLevel
	}

	return zap.Config{
		Level:            zap.NewAtomicLevelAt(level),                        // Default log level: Info
		Development:      true,                                               // Development mode by default
		Encoding:         "json",                                             // JSON log format
		OutputPaths:      []string{"stdout", "/var/log/cyberMonkey/eos.log"}, // Log to console and file
		ErrorOutputPaths: []string{"stderr"},                                 // Log errors to stderr
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),                  // Use pre-configured development encoder
	}
}

// EnsureLogPermissions ensures correct permissions for log directory & file.
func EnsureLogPermissions(logFilePath string) error {
	dir := filepath.Dir(logFilePath)

	// Ensure the directory exists
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	// Set strict file permissions first
	if err := os.Chmod(dir, 0700); err != nil {
		return err
	}

	// Ensure the log file exists
	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		file, err := os.Create(logFilePath)
		if err != nil {
			return err
		}
		file.Close()
	}

	// Set strict file permissions
	return os.Chmod(logFilePath, 0600)
}

func InitializeWithConfig(cfg zap.Config) {
	if Log != nil {
		return // Prevent re-initialization
	}

	// Ensure permissions for each log output path BEFORE initializing logger
	for _, path := range cfg.OutputPaths {
		if path != "stdout" && path != "stderr" {
			if err := EnsureLogPermissions(path); err != nil {
				println("Permission error:", err.Error())
				panic("Failed to ensure permissions for log file: " + err.Error())
			}
		}
	}

	// Now safely build logger
	var err error
	Log, err = cfg.Build()
	if err != nil {
		cfg.OutputPaths = []string{"stdout"} // Fallback to stdout
		Log, err = cfg.Build()
		if err != nil {
			panic("Failed to initialize logger with fallback config: " + err.Error())
		}
	}
}

func Initialize() {
	InitializeWithConfig(DefaultConfig())
	zap.ReplaceGlobals(Log) // Ensures zap.L() always uses this logger
	Log.Info("Logger successfully initialized", zap.String("log_level", DefaultConfig().Level.String()))
}

// GetLogger returns the global logger instance.
func GetLogger() *zap.Logger {
	if Log == nil {
		Initialize()
	}
	return Log
}

// -------------------------- HELPER FUNCTIONS FOR CLEANER LOGGING --------------------------

// Info logs an informational message.
func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

// Warn logs a warning message.
func Warn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}

// Error logs an error message.
func Error(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}

// Debug logs a debug message.
func Debug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

// Fatal logs a fatal error and exits.
func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
}

// Panic logs a message & panics the application
func Panic(msg string, fields ...zap.Field) {
	GetLogger().Panic(msg, fields...)
}

// Sync flushes any buffered log entries & returns error if it fails.
func Sync() error {
	if Log != nil {
		if err := Log.Sync(); err != nil {
			if _, ok := err.(*os.PathError); !ok {
				Log.Error("Failed to sync logger", zap.Error(err))
			}
			return err
		}
	}
	return nil
}

// ✅ Add this helper to guarantee a logger (fallback to console logger if needed)
func GetSafeLogger() *zap.Logger {
	if Log == nil {
		fmt.Println("⚠️ Logger not initialized, using fallback zap logger")
		fallback, _ := zap.NewDevelopment()
		return fallback
	}
	return Log
}
