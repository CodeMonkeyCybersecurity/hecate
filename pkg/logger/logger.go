// pkg/logger/logger.go
package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

var Log *zap.Logger

// DefaultConfig returns a standard zap.Config object with custom settings.
func DefaultConfig() zap.Config {
	return zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.InfoLevel),                // Default log level: Info
		Development:      true,                                               // Development mode by default
		Encoding:         "json",                                             // JSON log format
		OutputPaths:      []string{"stdout", "/var/log/cyberMonkey/eos.log"}, // Log to console and file
		ErrorOutputPaths: []string{"stderr"},                                 // Log errors to stderr
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),                  // Use pre-configured development encoder

	}
}

// EnsureLogPermissions ensures the correct permissions for the log directory and file.
func EnsureLogPermissions(logFilePath string) error {
	dir := filepath.Dir(logFilePath)

	// Ensure the directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err // Return the error if directory creation fails
		}
	} else {
		// Set stricter permissions for the directory
		if err := os.Chmod(dir, 0700); err != nil {
			return err // Return the error if permission setting fails
		}
	}


	// Ensure the log file exists
	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		file, err := os.Create(logFilePath)
		if err != nil {
			return err // Return the error if file creation fails
		}
		file.Close()
	}

	// Set permissions for the log file (read/write for owner only)
	if err := os.Chmod(logFilePath, 0600); err != nil {
		return err // Return the error if permission setting fails
	}

	return nil
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
}

// GetLogger returns the global logger instance.
func GetLogger() *zap.Logger {
	if Log == nil {
		Initialize()
	}
	return Log
}

// LogCommandExecution logs when a command is executed
func LogCommandExecution(cmdName string, args []string) {
	Log.Info("Command executed", zap.String("command", cmdName), zap.Strings("args", args))
}

// Sync flushes any buffered log entries.
func Sync() {
	if Log != nil {
		err := Log.Sync()
		if err != nil {
			if _, ok := err.(*os.PathError); !ok {
				Log.Error("Failed to sync logger", zap.Error(err))
			}
		}
	}
}
