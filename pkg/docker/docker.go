package docker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"hecate/pkg/logger"
)


//
//---------------------------- CONTAINER FUNCTIONS ---------------------------- //
//

// RunDockerComposeService starts a specific service from a docker-compose file
func RunDockerComposeService(composeFile, service string) error {
	logger.Info("Starting Docker service", zap.String("service", service), zap.String("composeFile", composeFile))
	cmd := exec.Command("docker-compose", "-f", composeFile, "up", "-d", service)
	err := cmd.Run()
	if err != nil {
		logger.Error("Failed to start Docker service", zap.String("service", service), zap.Error(err))
		return err
	}
	logger.Info("Docker service started successfully", zap.String("service", service))
	return nil
}

// RemoveVolumes removes the specified Docker volumes.
func RemoveVolumes(volumes []string) error {
    for _, volume := range volumes {
        // Execute the docker volume rm command.
        if err := Execute("docker", "volume", "rm", volume); err != nil {
            logger.Warn("failed to remove volume", zap.String("volume", volume), zap.Error(err))
        } else {
            logger.Info("Volume removed successfully", zap.String("volume", volume))
        }
    }
    return nil
}

// StopContainers stops the specified Docker containers.
func StopContainers(containers []string) error {
	// Build the arguments for "docker stop" command.
	args := append([]string{"stop"}, containers...)
	
	// Execute the command.
	if err := Execute("docker", args...); err != nil {
		return fmt.Errorf("failed to stop containers %v: %w", containers, err)
	}
	
	// Log the successful stopping of containers.
	logger.Info("Containers stopped successfully", zap.Any("containers", containers))
	return nil
}

// RemoveContainers removes the specified Docker containers.
func RemoveContainers(containers []string) error {
	args := append([]string{"rm"}, containers...)
	if err := Execute("docker", args...); err != nil {
		return fmt.Errorf("failed to remove containers %v: %w", containers, err)
	}
	logger.Info("Containers removed successfully", zap.Any("containers", containers))
	return nil
}

// RemoveImages removes the specified Docker images.
// It logs a warning if an image cannot be removed, but continues with the others.
func RemoveImages(images []string) error {
	for _, image := range images {
		if err := Execute("docker", "rmi", image); err != nil {
			logger.Warn("failed to remove image (it might be used elsewhere)",
				zap.String("image", image), zap.Error(err))
		} else {
			logger.Info("Image removed successfully", zap.String("image", image))
		}
	}
	return nil
}

// BackupVolume backs up a single Docker volume by running a temporary Alpine container.
// It returns the full path to the backup file.
func BackupVolume(volumeName, backupDir string) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	backupFile := fmt.Sprintf("%s_%s.tar.gz", timestamp, volumeName)
	cmd := []string{
		"run", "--rm",
		"-v", fmt.Sprintf("%s:/volume", volumeName),
		"-v", fmt.Sprintf("%s:/backup", backupDir),
		"alpine",
		"tar", "czf", fmt.Sprintf("/backup/%s", backupFile),
		"-C", "/volume", ".",
	}
	if err := Execute("docker", cmd...); err != nil {
		return "", fmt.Errorf("failed to backup volume %s: %w", volumeName, err)
	}
	return filepath.Join(backupDir, backupFile), nil
}

// BackupVolumes backs up all provided volumes to the specified backupDir.
// It returns a map with volume names as keys and their backup file paths as values.
// If any backup fails, it logs the error but continues processing the remaining volumes.
func BackupVolumes(volumes []string, backupDir string) (map[string]string, error) {
	backupResults := make(map[string]string)

	// Ensure the backup directory exists.
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return backupResults, fmt.Errorf("failed to create backup directory %s: %w", backupDir, err)
	}

	// Loop through each volume and back it up.
	for _, vol := range volumes {
		logger.Info("Backing up volume", zap.String("volume", vol))
		backupFile, err := BackupVolume(vol, backupDir)
		if err != nil {
			logger.Error("Error backing up volume", zap.String("volume", vol), zap.Error(err))
			// Continue processing other volumes even if one fails.
		} else {
			logger.Info("Volume backup completed", zap.String("volume", vol), zap.String("backupFile", backupFile))
			backupResults[vol] = backupFile
		}
	}
	return backupResults, nil
}

// ParseComposeFile reads a docker-compose file and returns container names, images, and volumes.
func ParseComposeFile(composePath string) (containers []string, images []string, volumes []string, err error) {
	data, err := os.ReadFile(composePath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to read compose file: %w", err)
	}

	var compose ComposeFile
	if err := yaml.Unmarshal(data, &compose); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to unmarshal compose file: %w", err)
	}

	// Extract container names and images from services.
	for key, svc := range compose.Services {
		// If ContainerName is not provided, you can decide to use the service key
		if svc.ContainerName != "" {
			containers = append(containers, svc.ContainerName)
		} else {
			containers = append(containers, key)
		}
		if svc.Image != "" {
			images = append(images, svc.Image)
		}
	}

	// Extract volume names.
	for volName := range compose.Volumes {
		volumes = append(volumes, volName)
	}

	logger.Info("Parsed compose file successfully", zap.String("path", composePath),
		zap.Any("containers", containers), zap.Any("images", images), zap.Any("volumes", volumes))

	return containers, images, volumes, nil
}


// EnsureArachneNetwork checks if the Docker network "arachne-net" exists.
// If it does not exist, it creates it with the desired IPv4 and IPv6 subnets.
func EnsureArachneNetwork() error {
	networkName := "arachne-net"
	desiredIPv4 := "10.1.0.0/16"
	desiredIPv6 := "fd42:1a2b:3c4d:5e6f::/64"

	// Check if the network exists by running: docker network inspect arachne-net
	cmd := exec.Command("docker", "network", "inspect", networkName)
	if err := cmd.Run(); err == nil {
		// Network exists, so just return
		return nil
	}

	// If the network does not exist, create it with the specified subnets.
	createCmd := exec.Command("docker", "network", "create",
		"--driver", "bridge",
		"--subnet", desiredIPv4,
		"--ipv6",
		"--subnet", desiredIPv6,
		networkName,
	)
	output, err := createCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create network %s: %v, output: %s", networkName, err, output)
	}

	return nil
}

// CheckDockerContainers runs "docker ps" and logs its output.
// It returns an error if the command fails.
func CheckDockerContainers() error {
	cmd := exec.Command("docker", "ps")
	output, err := cmd.CombinedOutput()
	// Print output to terminal
	fmt.Println(string(output))
	if err != nil {
		return fmt.Errorf("failed to run docker ps: %v, output: %s", err, output)
	}
	logger.Info("Docker ps output", zap.String("output", string(output)))
	return nil
}
