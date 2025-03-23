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


var log = logger.GetLogger()

//
//---------------------------- RESTORE ---------------------------- //
//

// RestoreDir restores a backup directory by copying it to the target location.
func RestoreDir(src, dest string) error {
	
	log.Info("Restoring directory", zap.String("src", src), zap.String("dest", dest))

	err := CopyDir(src, dest)
	if err != nil {
		log.Error("Failed to restore directory", zap.Error(err))
		return fmt.Errorf("failed to restore directory: %w", err)
	}
	log.Info("Directory restored successfully", zap.String("dest", dest))
	return nil
}

// RestoreFile restores a backup file to the original location.
func RestoreFile(src, dest string) error {
	
	log.Info("Restoring file", zap.String("src", src), zap.String("dest", dest))

	err := CopyFile(src, dest)
	if err != nil {
		log.Error("Failed to restore file", zap.Error(err))
		return fmt.Errorf("failed to restore file: %w", err)
	}
	log.Info("File restored successfully", zap.String("dest", dest))
	return nil
}

// FindLatestBackup finds the latest backup file with a given prefix.
func FindLatestBackup(prefix string) (string, error) {
	
	log.Info("Searching for latest backup with prefix", zap.String("prefix", prefix))

	dir := "." // or the directory where your backups are stored
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Error("Failed to read directory", zap.Error(err))
		return "", err
	}

	var latest string
	var latestModTime time.Time
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			if info.ModTime().After(latestModTime) {
				latest = entry.Name()
				latestModTime = info.ModTime()
			}
		}
	}

	if latest == "" {
		return "", fmt.Errorf("no backups found with prefix %s", prefix)
	}

	log.Info("Latest backup found", zap.String("filename", latest))
	return latest, nil
}

//
//---------------------------- DEPLOY ---------------------------- //
//

// DeployApp deploys the application by copying necessary config files and restarting services
func DeployApp(app string, force bool) error {
	log.Info("🚀 Starting deployment", zap.String("app", app), zap.Bool("force", force))

	if err := ValidateConfigPaths(app); err != nil {
		return fmt.Errorf("failed to validate config paths: %w", err)
	}
	
	httpSrc := filepath.Join("assets/servers", app+".conf")
	httpDest := filepath.Join("/etc/nginx/sites-available", app)
	streamSrc := filepath.Join("assets/stream", app+".conf")
	streamDest := filepath.Join("/etc/nginx/stream.d", app+".conf")
	symlinkPath := filepath.Join("/etc/nginx/sites-enabled", app)

	// Check if config already exists
	if _, err := os.Stat(httpDest); err == nil && !force {
	    if !isContainerRunning(app) {
	        log.Warn("No active container detected, cleaning up stale deployment", zap.String("app", app))
	        if err := RemoveApp(app); err != nil {
	            return fmt.Errorf("failed to remove stale deployment: %w", err)
	        }
	    } else {
	        log.Warn("❌ Application already deployed. Use --force to overwrite.", zap.String("app", app))
	        return fmt.Errorf("application %s already deployed. Use --force to overwrite", app)
	    }
	}

	// Clean up existing files if force is enabled
	if force {
		log.Warn("⚠️ Overwriting existing deployment", zap.String("app", app))
		if err := RemoveApp(app); err != nil {
			return fmt.Errorf("failed to overwrite existing deployments: %w", err)
		}
	}	

	// Copy HTTP config
	if err := CopyFile(httpSrc, httpDest); err != nil {
		log.Error("❌ Failed to copy HTTP config", zap.String("src", httpSrc), zap.Error(err))
		return fmt.Errorf("failed to copy HTTP config: %w", err)
	}
	log.Info("✅ HTTP config copied", zap.String("dest", httpDest))

	// Copy Stream config (if present)
	if FileExists(streamSrc) {
		if err := CopyFile(streamSrc, streamDest); err != nil {
			log.Error("❌ Failed to copy stream config", zap.String("src", streamSrc), zap.Error(err))
			return fmt.Errorf("failed to copy stream config: %w", err)
		}
		log.Info("✅ Stream config copied", zap.String("dest", streamDest))
	}

	// Symlink into sites-enabled
	if err := os.Symlink(httpDest, symlinkPath); err != nil && !os.IsExist(err) {
		log.Error("❌ Failed to create symlink", zap.String("link", symlinkPath), zap.Error(err))
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	log.Info("🔗 Symlink created", zap.String("from", symlinkPath), zap.String("to", httpDest))

	// Test Nginx configuration
	cmdTest := exec.Command("nginx", "-t")
	if output, err := cmdTest.CombinedOutput(); err != nil {
		log.Error("❌ Nginx config test failed", zap.String("output", string(output)), zap.Error(err))
		return fmt.Errorf("nginx config test failed: %s", string(output))
	}

	// Restart Nginx
	cmdRestart := exec.Command("systemctl", "restart", "nginx")
	if err := cmdRestart.Run(); err != nil {
		log.Error("❌ Failed to restart Nginx", zap.Error(err))
		return fmt.Errorf("failed to restart nginx: %w", err)
	}

	log.Info("✅ Deployment successful", zap.String("app", app))
	return nil
}

// isContainerRunning checks whether a container with the given app name is running.
func isContainerRunning(app string) bool {
	// Adjust the filter if your container names differ.
	out, err := exec.Command("docker", "ps", "--filter", "name="+app, "--format", "{{.Names}}").Output()
	if err != nil {
		log.Error("Error checking container status", zap.Error(err))
		return false
	}
	containerNames := strings.TrimSpace(string(out))
	return containerNames != ""
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
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		log.Warn("Unexpected error checking file", zap.String("file", filename), zap.Error(err))
		return false
	}
	return true
}


func CopyFile(src, dst string) error {
	

	srcFile, err := os.Open(src)
	if err != nil {
		if os.IsNotExist(err) {
			log.Error("Source file not found", zap.String("file", src))
			return fmt.Errorf("source file does not exist: %s", src)
		}
		log.Error("Error opening source file", zap.String("file", src), zap.Error(err))
		return fmt.Errorf("error opening source file: %w", err)
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		log.Error("Error creating destination file", zap.String("file", dst), zap.Error(err))
		return fmt.Errorf("error creating destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy contents
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		log.Error("Error copying file contents", zap.String("src", src), zap.String("dst", dst), zap.Error(err))
		return fmt.Errorf("error copying file: %w", err)
	}

	return nil
}



// RemoveIfExists removes a file or directory if it exists.
func RemoveIfExists(path string) error {
	

	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No need to remove if it doesn't exist
		}
		log.Error("Error checking path before removal", zap.String("path", path), zap.Error(err))
		return fmt.Errorf("error checking path before removal: %w", err)

	}
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to remove path %s: %w", path, err)
	}
	return nil
}


// CopyDir copies the contents of a directory from src to dst.
func CopyDir(src string, dst string) error {
	if _, err := os.Stat(src); err != nil {
		if os.IsNotExist(err) {
			log.Error("Source directory not found", zap.String("dir", src))
			return fmt.Errorf("source directory does not exist: %s", src)
		}
		return fmt.Errorf("error checking source directory: %w", err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

    if err := os.MkdirAll(dst, 0755); err != nil {
        return fmt.Errorf("failed to create destination directory %s: %w", dst, err)
	}
    
    for _, entry := range entries {
        srcPath := filepath.Join(src, entry.Name())
        dstPath := filepath.Join(dst, entry.Name())

        if entry.IsDir() {
            if err := CopyDir(srcPath, dstPath); err != nil {
                return fmt.Errorf("failed to copy subdirectory %s: %w", srcPath, err)
            }
        } else {
            input, err := os.ReadFile(srcPath)
            if err != nil {
                return fmt.Errorf("failed to read file %s: %w", srcPath, err)
            }
            if err := os.WriteFile(dstPath, input, 0644); err != nil {
                return fmt.Errorf("failed to write file %s: %w", dstPath, err)
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
	log.Debug("Computing SHA256 hash", zap.String("input", s))
	hash := sha256.Sum256([]byte(s))
	hashStr := hex.EncodeToString(hash[:])
	log.Debug("Computed SHA256 hash", zap.String("hash", hashStr))
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
		log.Error("Failed to open log file for monitoring", zap.String("logFilePath", logFilePath), zap.Error(err))
		return fmt.Errorf("failed to open log file for monitoring: %w", err)
	}
	defer file.Close()

	// Seek to the end of the file so we only see new log lines.
	_, err = file.Seek(0, io.SeekEnd)
	if err != nil {
		log.Error("Failed to seek log file", zap.String("logFilePath", logFilePath), zap.Error(err))
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
				log.Debug("Vault Log Line", zap.String("logLine", line))
				if strings.Contains(line, marker) {
					log.Info("Vault marker found, exiting log monitor", zap.String("marker", marker))
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
	log.Info("Retrieving internal hostname")
	hostname, err := os.Hostname()
	if err != nil {
		log.Error("Unable to retrieve hostname, defaulting to localhost", zap.Error(err))
		return "localhost"
	}
	log.Info("Retrieved hostname", zap.String("hostname", hostname))
	return hostname
}


//
//---------------------------- ERROR HANDLING ---------------------------- //
//

// HandleError logs an error and optionally exits the program
func HandleError(err error, message string, exit bool) {
	if err != nil {
		log.Error(message, zap.Error(err))
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
	log.Debug("Processing YAML map")
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
	log.Debug("Completed processing YAML map")
}


//
//---------------------------- DEPLOY HELPERS ---------------------------- //
//

// quote adds quotes around a string for cleaner logging
func quote(s string) string {
	return fmt.Sprintf("%q", s)
}

// RemoveApp deletes existing deployment files safely (used with --force)
func RemoveApp(app string) error {
	
	httpDest := filepath.Join("/etc/nginx/sites-available", app)
	streamDest := filepath.Join("/etc/nginx/stream.d", app+".conf")
	symlinkPath := filepath.Join("/etc/nginx/sites-enabled", app)

	// Remove main config
	if err := os.RemoveAll(httpDest); err != nil && !os.IsNotExist(err) {
		log.Error("❌ Failed to remove HTTP config", zap.String("file", httpDest), zap.Error(err))
		return fmt.Errorf("failed to remove HTTP config: %w", err)
	}

	// Remove stream config if it exists
	if err := os.RemoveAll(streamDest); err != nil && !os.IsNotExist(err) {
		log.Error("❌ Failed to remove stream config", zap.String("file", streamDest), zap.Error(err))
		return fmt.Errorf("failed to remove stream config: %w", err)
	}

	// Remove symlink
	if err := os.RemoveAll(symlinkPath); err != nil && !os.IsNotExist(err) {
		log.Error("❌ Failed to remove symlink", zap.String("link", symlinkPath), zap.Error(err))
		return fmt.Errorf("failed to remove symlink: %w", err)
	}

	log.Info("✅ Existing deployment cleaned up", zap.String("app", app))
	return nil
}

// ValidateConfigPaths checks that the app’s Nginx source config files exist
func ValidateConfigPaths(app string) error {
	
	httpSrc := filepath.Join("assets/servers", app+".conf")

	if _, err := os.Stat(httpSrc); err != nil {
		if os.IsNotExist(err) {
			log.Error("❌ Required config file not found", zap.String("file", httpSrc))
			return fmt.Errorf("missing HTTP config: %s", httpSrc)
		}
		return fmt.Errorf("error checking config file: %w", err)
	}	

	// Stream config is optional — no error if missing
	return nil
}
