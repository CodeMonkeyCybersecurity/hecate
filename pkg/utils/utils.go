// pkg/utils/utils.go
package utils

import (
	"bufio"
	"context"
	"crypto/sha256"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"io"

	"gopkg.in/yaml.v2"
	"go.uber.org/zap"
	"github.com/spf13/cobra"

	"hecate/pkg/logger"
	"hecate/pkg/config"
)

//
//---------------------------- HECATE FUNCTIONS ---------------------------- //
//

// DeployApp deploys an application by copying necessary configs and restarting services
func DeployApp(app string, cmd *cobra.Command) error {
	logger.Info("Starting deployment", zap.String("app", app))  // ✅ Use logger.Info directly

	// Check if the required HTTP config exists
	httpConfig := filepath.Join(assetsPath, "servers", app+".conf")
	if !FileExists(httpConfig) {
		logger.Error("Missing HTTP config file", zap.String("file", httpConfig))
		return fmt.Errorf("missing Nginx HTTP config for %s", app)
	}

	// Copy HTTP config
	if err := CopyFile(httpConfig, filepath.Join(nginxConfPath, app+".conf")); err != nil {
		return fmt.Errorf("failed to copy HTTP config: %w", err)
	}

	// Copy Stream config if available
	streamConfig := filepath.Join(assetsPath, "stream", app+".conf")
	if FileExists(streamConfig) {
		if err := CopyFile(streamConfig, filepath.Join(nginxStreamPath, app+".conf")); err != nil {
			return fmt.Errorf("failed to copy Stream config: %w", err)
		}
	}

	// Handle NextCloud Coturn deployment
	if app == "nextcloud" {
		noTalk, _ := cmd.Flags().GetBool("without-talk")
		if !noTalk {
			logger.Info("Deploying Coturn for NextCloud Talk")
			if err := RunDockerComposeService(dockerComposeFile, "coturn"); err != nil {
				return fmt.Errorf("failed to deploy Coturn: %w", err)
			}
		} else {
			logger.Info("Skipping Coturn deployment")
		}
	}

	// Validate and restart Nginx
	if err := ValidateNginx(); err != nil {
		return fmt.Errorf("invalid Nginx configuration: %w", err)
	}

	if err := RestartNginx(); err != nil {
		return fmt.Errorf("failed to restart Nginx: %w", err)
	}

	logger.Info("Deployment successful", zap.String("app", app))
	fmt.Printf("Successfully deployed %s!\n", app)
	return nil
}

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

// ValidateNginx runs `nginx -t` to check configuration validity
func ValidateNginx() error {
	logger.Info("Validating Nginx configuration...")
	cmd := exec.Command("nginx", "-t")
	err := cmd.Run()
	if err != nil {
		logger.Error("Nginx configuration validation failed", zap.Error(err))
	}
	return err
}

// RestartNginx reloads the Nginx service
func RestartNginx() error {
	logger.Info("Restarting Nginx...")
	cmd := exec.Command("systemctl", "reload", "nginx")
	err := cmd.Run()
	if err != nil {
		logger.Error("Failed to restart Nginx", zap.Error(err))
	}
	return err
}


//
//---------------------------- COMMAND EXECUTION ---------------------------- //
//

// Execute runs a command with separate arguments.
func Execute(command string, args ...string) error {
	logger.Debug("Executing command", zap.String("command", command), zap.Strings("args", args))
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		logger.Error("Command execution failed", zap.String("command", command), zap.Strings("args", args), zap.Error(err))
	} else {
		logger.Info("Command executed successfully", zap.String("command", command))
	}
	return err
}

// ExecuteShell runs a shell command with pipes (`| grep`).
func ExecuteShell(command string) error {
	logger.Debug("Executing shell command", zap.String("command", command))
	cmd := exec.Command("bash", "-c", command) // Runs in shell mode
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		logger.Error("Shell command execution failed", zap.String("command", command), zap.Error(err))
	} else {
		logger.Info("Shell command executed successfully", zap.String("command", command))
	}
	return err
}

func ExecuteInDir(dir, command string, args ...string) error {
	logger.Debug("Executing command in directory", zap.String("directory", dir), zap.String("command", command), zap.Strings("args", args))
	cmd := exec.Command(command, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		logger.Error("Command execution failed in directory", zap.String("directory", dir), zap.String("command", command), zap.Strings("args", args), zap.Error(err))
	} else {
		logger.Info("Command executed successfully in directory", zap.String("directory", dir), zap.String("command", command))
	}
	return err
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
