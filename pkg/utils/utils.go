// pkg/utils/utils.go

package utils

import (
	"bufio"
	"context"
	"crypto/sha256"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
	"io"

	"go.uber.org/zap"

	"hecate/pkg/logger"
	"hecate/pkg/config"
)

// DeployApp deploys the application by copying necessary config files and restarting services
func DeployApp(app string, force bool) error {
	log := logger.GetLogger()
	if log == nil {
		fmt.Println("⚠️ Warning: Logger is nil. Defaulting to console output.")
	}

	log.Info("🚀 Starting deployment", zap.String("app", app), zap.Bool("force", force))

	appConfigPath := filepath.Join("/opt/hecate/configs", app)
	appDeploymentPath := filepath.Join("/etc/nginx/sites-available", app)
	appEnabledPath := filepath.Join("/etc/nginx/sites-enabled", app)

	// Check if config already exists
	if _, err := os.Stat(appDeploymentPath); err == nil {
		if !force {
			errMsg := fmt.Sprintf("❌ Application %s is already deployed. Use --force to overwrite.", app)
			log.Warn(errMsg)
			return fmt.Errorf(errMsg)
		}
		log.Warn("⚠️ Overwriting existing deployment", zap.String("app", app))
		
		// Remove existing deployment
		if err := os.Remove(appDeploymentPath); err != nil {
			log.Error("❌ Failed to remove existing config", zap.String("app", app), zap.Error(err))
			return err
		}

		// Remove existing symlink in sites-enabled
		if err := os.RemoveAll(appEnabledPath); err != nil {
			log.Error("❌ Failed to remove existing symlink", zap.String("app", app), zap.Error(err))
			return err
		}
	}

	// Copy new config file
	err := CopyFile(appConfigPath, appDeploymentPath)
	if err != nil {
		log.Error("❌ Failed to copy config file", zap.String("app", app), zap.Error(err))
		return err
	}

	// Create a symlink in `sites-enabled`
	if err := os.Symlink(appDeploymentPath, appEnabledPath); err != nil {
		log.Error("❌ Failed to create symlink", zap.String("app", app), zap.Error(err))
		return err
	}

	// Test Nginx configuration before restarting
	cmdTest := exec.Command("nginx", "-t")
	if output, err := cmdTest.CombinedOutput(); err != nil {
		log.Error("❌ Nginx config test failed", zap.String("output", string(output)), zap.Error(err))
		return fmt.Errorf("Nginx config test failed: %s", string(output))
	}

	// Restart Nginx
	cmdRestart := exec.Command("systemctl", "restart", "nginx")
	if err := cmdRestart.Run(); err != nil {
		log.Error("❌ Failed to restart Nginx", zap.Error(err))
		return err
	}

	log.Info("✅ Deployment successful", zap.String("app", app))
	return nil
}


//
//---------------------------- FORCE ---------------------------- //
//

// ✅ Global force flag
var (
	force bool
	mu    sync.Mutex
)

// SetForce sets the force flag value.
func SetForce(value bool) {
	mu.Lock()
	defer mu.Unlock()
	force = value
}

// GetForce retrieves the force flag value.
func GetForce() bool {
	mu.Lock()
	defer mu.Unlock()
	return force
}


//
//---------------------------- FACT CHECKING ---------------------------- //
//

// ✅ Moved here since it may be used in multiple commands
func IsValidApp(app string) bool {
	for _, validApp := range config.GetSupportedAppNames() {
		if app == validApp {
			return true
		}
	}
	return false
}

//
//---------------------------- FILE CRUD ---------------------------- //
//


// FileExists checks if a file exists
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0644)
}


// RemoveIfExists removes a file or directory if it exists.
func RemoveIfExists(path string) error {
    if _, err := os.Stat(path); err == nil {
        return os.RemoveAll(path)
    }
    return nil
}

// CopyDir copies the contents of a directory from src to dst.
func CopyDir(src string, dst string) error {
    // This is a simple example; production code may need more robust error handling.
    entries, err := os.ReadDir(src)
    if err != nil {
        return err
    }
    if err := os.MkdirAll(dst, 0755); err != nil {
        return err
    }
    for _, entry := range entries {
        srcPath := filepath.Join(src, entry.Name())
        dstPath := filepath.Join(dst, entry.Name())
        if entry.IsDir() {
            if err := CopyDir(srcPath, dstPath); err != nil {
                return err
            }
        } else {
            input, err := os.ReadFile(srcPath)
            if err != nil {
                return err
            }
            if err := os.WriteFile(dstPath, input, 0644); err != nil {
                return err
            }
        }
    }
    return nil
}


//
//---------------------------- CRYPTO, HASHING, SECRETS ---------------------------- //
//

// HashString computes and returns the SHA256 hash of the provided string.
func HashString(s string) string {
	logger.Debug("Computing SHA256 hash", zap.String("input", s))
	hash := sha256.Sum256([]byte(s))
	hashStr := hex.EncodeToString(hash[:])
	logger.Debug("Computed SHA256 hash", zap.String("hash", hashStr))
	return hashStr
}

// generatePassword creates a random alphanumeric password of the given length.
func GeneratePassword(length int) (string, error) {
	// Generate random bytes. Since hex encoding doubles the length, we need length/2 bytes.
	bytes := make([]byte, length/2)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	// Encode to hex and trim to required length.
	return hex.EncodeToString(bytes)[:length], nil
}


//
//---------------------------- LOGGING ---------------------------- //
//

// monitorVaultLogs tails the log file and prints new lines to STDOUT.
// It returns when it sees a line containing the specified marker or when the context is done.
func MonitorVaultLogs(ctx context.Context, logFilePath, marker string) error {
	file, err := os.Open(logFilePath)
	if err != nil {
		logger.Error("Failed to open log file for monitoring", zap.String("logFilePath", logFilePath), zap.Error(err))
		return fmt.Errorf("failed to open log file for monitoring: %w", err)
	}
	defer file.Close()

	// Seek to the end of the file so we only see new log lines.
	_, err = file.Seek(0, io.SeekEnd)
	if err != nil {
		logger.Error("Failed to seek log file", zap.String("logFilePath", logFilePath), zap.Error(err))
		return fmt.Errorf("failed to seek log file: %w", err)
	}

	scanner := bufio.NewScanner(file)
	for {
		select {
		case <-ctx.Done():
			logger.Warn("Timeout reached while waiting for Vault to start")
			return fmt.Errorf("timeout reached while waiting for Vault to start")
		default:
			if scanner.Scan() {
				line := scanner.Text()
				fmt.Println(line) // Print the log line to terminal
				logger.Debug("Vault Log Line", zap.String("logLine", line))
				if strings.Contains(line, marker) {
					logger.Info("Vault marker found, exiting log monitor", zap.String("marker", marker))
					return nil
				}
			} else {
				time.Sleep(500 * time.Millisecond) // No new line, wait and try again
			}
		}
	}
}

// ✅ Logs safely, even if logger is nil
func SafeLog(log *zap.Logger, msg string, fields ...zap.Field) {
	if log == nil {
		fmt.Println(msg)
	} else {
		log.Info(msg, fields...)
	}
}

// ✅ Logs an error safely, even if logger is nil
func PrintError(log *zap.Logger, msg string) {
	if log != nil {
		log.Error(msg)
	}
	fmt.Println(msg)
}

//
//---------------------------- HOSTNAME ---------------------------- //
//

// GetInternalHostname returns the machine's hostname.
// If os.Hostname() fails, it logs the error and returns "localhost".
func GetInternalHostname() string {
	logger.Info("Retrieving internal hostname")
	hostname, err := os.Hostname()
	if err != nil {
		logger.Error("Unable to retrieve hostname, defaulting to localhost", zap.Error(err))
		return "localhost"
	}
	logger.Info("Retrieved hostname", zap.String("hostname", hostname))
	return hostname
}


//
//---------------------------- ERROR HANDLING ---------------------------- //
//

// HandleError logs an error and optionally exits the program
func HandleError(err error, message string, exit bool) {
	if err != nil {
		logger.Error(message, zap.Error(err))
		if exit {
			logger.Fatal("Exiting program due to error", zap.String("message", message))
		}
	}
}

// WithErrorHandling wraps a function with error handling
func WithErrorHandling(fn func() error) {
	err := fn()
	if err != nil {
		HandleError(err, "An error occurred", true)
	}
}


//
//---------------------------- PERMISSIONS ---------------------------- //
//

// CheckSudo checks if the current user has sudo privileges
func CheckSudo() bool {
	cmd := exec.Command("sudo", "-n", "true") // Non-interactive sudo check
	if err := cmd.Run(); err != nil {
		logger.Warn("User does not have sudo privileges", zap.Error(err))
		return false
	}
	return true
}


//
//---------------------------- YAML ---------------------------- //
//


// Recursive function to process and print nested YAML structures
func ProcessMap(data map[string]interface{}, indent string) {
	logger.Debug("Processing YAML map")
	for key, value := range data {
		switch v := value.(type) {
		case map[string]interface{}:
			fmt.Printf("%s%s:\n", indent, key)
			ProcessMap(v, indent+"  ")
		case []interface{}:
			fmt.Printf("%s%s:\n", indent, key)
			for _, item := range v {
				if itemMap, ok := item.(map[string]interface{}); ok {
					ProcessMap(itemMap, indent+"  ")
				} else {
					fmt.Printf("%s  - %v\n", indent, item)
				}
			}
		default:
			fmt.Printf("%s%s: %v\n", indent, key, v)
		}
	}
	logger.Debug("Completed processing YAML map")
}
