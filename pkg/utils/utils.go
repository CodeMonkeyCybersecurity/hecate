// pkg/utils/utils.go
package utils

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/CodeMonkeyCybersecurity/hecate/pkg/config"
)

// Constants for file and directory names.
const (
	LastValuesFile    = ".hecate.conf"
	ConfDir           = "conf.d"
	DockerComposeFile = "docker-compose.yml"
)

//
// -------------------- Generic Functions --------------------
//

func RunCommand(command []string, shell bool) error {
	if shell {
		fmt.Printf("Running command: %s\n", strings.Join(command, " "))
		cmd := exec.Command("sh", "-c", strings.Join(command, " "))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	} else {
		fmt.Printf("Running command: %s\n", strings.Join(command, " "))
		cmd := exec.Command(command[0], command[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
}

// PromptWithDefault displays a prompt with a default value and returns the user's input or the default.
func PromptWithDefault(prompt, def, description string) string {
	fmt.Println()
	fmt.Println(description)
	fmt.Printf("%s [%s]: ", prompt, def)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return def
	}
	input = strings.TrimSpace(input)
	if input == "" {
		return def
	}
	return input
}

// CombineMarkers returns a new slice containing the provided defaultMarkers plus any additional markers.
func CombineMarkers(defaultMarkers []string, additional ...string) []string {
	markers := make([]string, len(defaultMarkers))
	copy(markers, defaultMarkers)
	markers = append(markers, additional...)
	return markers
}

// SaveLastValues writes key="value" pairs to LastValuesFile.
func SaveLastValues(values map[string]string) error {
    file, err := os.Create(LastValuesFile)
    if err != nil {
        return err
    }
    defer file.Close()

    for key, value := range values {
        _, err := file.WriteString(fmt.Sprintf("%s=\"%s\"\n", key, value))
        if err != nil {
            return err
        }
    }
    return nil
}

// LoadLastValues reads key="value" pairs from LastValuesFile.
func LoadLastValues() (map[string]string, error) {
    values := make(map[string]string)
    file, err := os.Open(LastValuesFile)
    if err != nil {
        if os.IsNotExist(err) {
            return values, nil
        }
        return nil, err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line == "" || !strings.Contains(line, "=") {
            continue
        }
        parts := strings.SplitN(line, "=", 2)
        key := strings.TrimSpace(parts[0])
        value := strings.Trim(strings.TrimSpace(parts[1]), `"`)
        values[key] = value
    }
    return values, scanner.Err()
}

//
// -------------------- Docker Compose Functions --------------------
//

// RunComposeInteractive updates the docker-compose file interactively.
func RunComposeInteractive() {
    fmt.Println("\n=== Docker Compose Update ===")
    const LAST_VALUES_FILE = ".hecate.conf"
    lastValues, err := LoadLastValues()
    if err != nil {
        fmt.Printf("Error loading configuration: %v\n", err)
        os.Exit(1)
    }
    defaultSelection := lastValues["APPS_SELECTION"]

    // Display the options before prompting for selection.
    config.DisplayOptions()
    
    selectedApps, selectionStr := config.GetUserSelection(defaultSelection)
    lastValues["APPS_SELECTION"] = selectionStr
    if err := SaveLastValues(lastValues); err != nil {
        fmt.Printf("Error saving configuration: %v\n", err)
    }
    if err := UpdateComposeFile(selectedApps); err != nil {
        fmt.Printf("Error updating docker-compose file: %v\n", err)
        os.Exit(1)
    }
    // Output updated docker-compose file.
    composeFile := "docker-compose.yml" // adjust as needed
    data, err := os.ReadFile(composeFile)
    if err != nil {
        fmt.Printf("Error reading %s: %v\n", composeFile, err)
    } else {
        fmt.Println("\n---- Updated docker-compose.yml ----")
        fmt.Println(string(data))
    }
}


// UpdateComposeFile reads the docker-compose file and, for each line that contains any marker
// from a selected app, removes the leading '#' so that the line becomes active.
// Additionally, for certain apps (Nextcloud, Mailcow, and Wazuh), it will process entire blocks
// delimited by a designated start marker and finish marker.
func UpdateComposeFile(selectedApps map[string]config.App) error {
	content, err := os.ReadFile(DockerComposeFile)
	if err != nil {
		return fmt.Errorf("Error: %s not found", DockerComposeFile)
	}
	lines := strings.Split(string(content), "\n")
	uncommentRegex := regexp.MustCompile(`^(\s*)#\s*`)

	// First pass: Uncomment individual lines that contain any marker from any selected app.
	for i, line := range lines {
		for _, app := range selectedApps {
			for _, marker := range app.Markers {
				if strings.Contains(line, marker) {
					lines[i] = uncommentRegex.ReplaceAllString(line, "$1")
					goto NextLine
				}
			}
		}
	NextLine:
	}

	// Define block markers for services that need full block uncommenting.
	blockMarkers := map[string]struct {
		start  string
		finish string
	}{
		"wazuh":     {start: "# <- uncomment if using Wazuh behind Hecate", finish: "# <- finish"},
		"mailcow":   {start: "# <- uncomment if using Mailcow behind Hecate", finish: "# <- finish"},
		"nextcloud": {start: "# <- uncomment if using Nextcloud behind Hecate", finish: "# <- finish"},
	}

	// Second pass: Process block uncommenting for each selected app that has block markers.
	for appName, markers := range blockMarkers {
		// Check if this service is selected.
		if _, ok := selectedApps[strings.ToLower(appName)]; !ok {
			continue
		}
		inBlock := false
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.Contains(trimmed, markers.start) {
				inBlock = true
				// Uncomment the start marker line.
				lines[i] = uncommentRegex.ReplaceAllString(line, "$1")
				continue
			}
			if inBlock {
				lines[i] = uncommentRegex.ReplaceAllString(line, "$1")
				if strings.Contains(trimmed, markers.finish) {
					inBlock = false
				}
			}
		}
	}

	// Backup the original docker-compose file.
	if err := BackupFile(DockerComposeFile); err != nil {
		return err
	}
	outContent := strings.Join(lines, "\n")
	if err := os.WriteFile(DockerComposeFile, []byte(outContent), 0644); err != nil {
		return err
	}
	var selApps []string
	for _, app := range selectedApps {
		selApps = append(selApps, app.Name)
	}
	fmt.Printf("Updated %s for apps: %s\n", DockerComposeFile, strings.Join(selApps, ", "))
	return nil
}


// -------------------- Certificate Functions --------------------

func PromptSubdomain() string {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter the subdomain to configure (e.g. sub). Leave blank if none: ")
		subdomain, _ := reader.ReadString('\n')
		subdomain = strings.TrimSpace(subdomain)
		if subdomain == "" {
			fmt.Print("You entered a blank subdomain. Do you wish to continue with no subdomain? (Y/n): ")
			confirm, _ := reader.ReadString('\n')
			confirm = strings.TrimSpace(confirm)
			if confirm == "" || strings.EqualFold(confirm, "y") || strings.EqualFold(confirm, "yes") {
				return ""
			}
			continue
		} else {
			return subdomain
		}
	}
}

func ConfirmCertName(defaultName string) string {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("Use certificate name '%s'? (Y/n): ", defaultName)
		subdomain, _ := reader.ReadString('\n')
		subdomain = strings.TrimSpace(defaultName)
		if subdomain == "" {
			fmt.Print("You entered no name. Do you wish to continue with no certificate name? (Y/n): ")
			confirm, _ := reader.ReadString('\n')
			confirm = strings.TrimSpace(confirm)
			if confirm == "" || strings.EqualFold(confirm, "y") || strings.EqualFold(confirm, "yes") {
				return ""
			}
			continue
		} else {
			return subdomain
		}
	}
}


// -------------------- File Backup and Copy Functions --------------------

// BackupFile creates a backup of the given file by appending a timestamp.
func BackupFile(path string) error {
    info, err := os.Stat(path)
    if err != nil || info.IsDir() {
        return nil
    }
    timestamp := time.Now().Format("20060102-150405")
    dir := filepath.Dir(path)
    base := filepath.Base(path)
    backupPath := filepath.Join(dir, fmt.Sprintf("%s_%s.bak", timestamp, base))

    in, err := os.Open(path)
    if err != nil {
        return err
    }
    defer in.Close()

    out, err := os.Create(backupPath)
    if err != nil {
        return err
    }
    defer out.Close()

    if _, err := io.Copy(out, in); err != nil {
        return err
    }

    fmt.Printf("Backup of '%s' created as '%s'.\n", path, backupPath)
    return nil
}

// RemoveIfExists deletes a file or directory if it exists.
func RemoveIfExists(path string) error {
    if _, err := os.Stat(path); err == nil {
        info, err := os.Stat(path)
        if err != nil {
            return err
        }
        if info.IsDir() {
            fmt.Printf("Removing existing directory '%s'...\n", path)
            return os.RemoveAll(path)
        }
        fmt.Printf("Removing existing file '%s'...\n", path)
        return os.Remove(path)
    }
    return nil
}

// CopyFile copies a single file from src to dst.
func CopyFile(src, dst string) error {
    in, err := os.Open(src)
    if err != nil {
        return err
    }
    defer in.Close()

    out, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer func() { _ = out.Close() }()

    if _, err = io.Copy(out, in); err != nil {
        return err
    }

    if info, err := os.Stat(src); err == nil {
        return os.Chmod(dst, info.Mode())
    }
    return nil
}

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
func CopyDir(src, dst string) error {
    src = filepath.Clean(src)
    dst = filepath.Clean(dst)

    srcInfo, err := os.Stat(src)
    if err != nil {
        return err
    }
    if !srcInfo.IsDir() {
        return fmt.Errorf("source %s is not a directory", src)
    }

    if err = os.MkdirAll(dst, srcInfo.Mode()); err != nil {
        return err
    }

    entries, err := os.ReadDir(src)
    if err != nil {
        return err
    }
    for _, entry := range entries {
        srcPath := filepath.Join(src, entry.Name())
        dstPath := filepath.Join(dst, entry.Name())

        entryInfo, err := entry.Info()
        if err != nil {
            return err
        }

        if entryInfo.IsDir() {
            if err = CopyDir(srcPath, dstPath); err != nil {
                return err
            }
        } else {
            if err = CopyFile(srcPath, dstPath); err != nil {
                return err
            }
        }
    }
    return nil
}

// RestoreDir removes dstDir and copies backupDir -> dstDir.
func RestoreDir(backupDir, dstDir string) {
    info, err := os.Stat(backupDir)
    if err != nil || !info.IsDir() {
        fmt.Printf("Error: Backup directory '%s' does not exist or is not a directory.\n", backupDir)
        os.Exit(1)
    }
    if err := RemoveIfExists(dstDir); err != nil {
        fmt.Printf("Error removing %s: %v\n", dstDir, err)
        os.Exit(1)
    }
    if err := CopyDir(backupDir, dstDir); err != nil {
        fmt.Printf("Error during restore of %s: %v\n", backupDir, err)
        os.Exit(1)
    }
    fmt.Printf("Restore complete: '%s' has been restored to '%s'.\n", backupDir, dstDir)
}

// RestoreFile removes dstFile and copies backupFile -> dstFile.
func RestoreFile(backupFile, dstFile string) {
    info, err := os.Stat(backupFile)
    if err != nil || info.IsDir() {
        fmt.Printf("Error: Backup file '%s' does not exist or is a directory.\n", backupFile)
        os.Exit(1)
    }
    if err := RemoveIfExists(dstFile); err != nil {
        fmt.Printf("Error removing %s: %v\n", dstFile, err)
        os.Exit(1)
    }
    if err := CopyFile(backupFile, dstFile); err != nil {
        fmt.Printf("Error during restore of %s: %v\n", backupFile, err)
        os.Exit(1)
    }
    fmt.Printf("Restore complete: '%s' has been restored to '%s'.\n", backupFile, dstFile)
}

// FindLatestBackup finds the lexicographically greatest file whose name starts with `prefix` and ends with ".bak".
func FindLatestBackup(prefix string) (string, error) {
    entries, err := os.ReadDir(".")
    if err != nil {
        return "", err
    }
    var latest string
    for _, e := range entries {
        name := e.Name()
        if strings.HasPrefix(name, prefix) && strings.HasSuffix(name, ".bak") {
            if name > latest {
                latest = name
            }
        }
    }
    if latest == "" {
        return "", fmt.Errorf("No .bak files found for prefix %q", prefix)
    }
    return latest, nil
}

// UpdateFile reads a file, replaces placeholders, creates a backup if changes occur, then writes the new content.
func UpdateFile(path, BACKEND_IP, PERS_BACKEND_IP, DELPHI_BACKEND_IP, BASE_DOMAIN string) {
    original, err := os.ReadFile(path)
    if err != nil {
        fmt.Printf("Error reading %s: %v\n", path, err)
        return
    }
    content := string(original)
    newContent := strings.ReplaceAll(content, "${BACKEND_IP}", BACKEND_IP)
    newContent = strings.ReplaceAll(newContent, "${PERS_BACKEND_IP}", PERS_BACKEND_IP)
    newContent = strings.ReplaceAll(newContent, "${DELPHI_BACKEND_IP}", DELPHI_BACKEND_IP)
    newContent = strings.ReplaceAll(newContent, "${BASE_DOMAIN}", BASE_DOMAIN)

    if newContent != content {
        if err := BackupFile(path); err != nil {
            fmt.Printf("Error creating backup for %s: %v\n", path, err)
            return
        }
        err = os.WriteFile(path, []byte(newContent), 0644)
        if err != nil {
            fmt.Printf("Error writing %s: %v\n", path, err)
            return
        }
        fmt.Printf("Updated %s\n", path)
    }
}

// RemoveUnwantedConfFiles walks through ConfDir and deletes any .conf file whose base name is not in allowedFiles.
func RemoveUnwantedConfFiles(allowedFiles map[string]bool) {
	info, err := os.Stat(ConfDir)
	if err != nil || !info.IsDir() {
		fmt.Printf("Error: Directory '%s' not found.\n", ConfDir)
		os.Exit(1)
	}
	var removedFiles []string
	err = filepath.Walk(ConfDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".conf") {
			if !allowedFiles[info.Name()] {
				if err := os.Remove(path); err != nil {
					fmt.Printf("Error removing %s: %v\n", path, err)
				} else {
					removedFiles = append(removedFiles, path)
					fmt.Printf("Removed: %s\n", path)
				}
			}
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error walking through '%s': %v\n", ConfDir, err)
		os.Exit(1)
	}
	if len(removedFiles) == 0 {
		fmt.Println("No configuration files were removed.")
	} else {
		fmt.Println("\nCleanup complete. The following files were removed:")
		for _, f := range removedFiles {
			fmt.Printf(" - %s\n", f)
		}
	}
}

// ProcessConfDirectory walks through a directory recursively, updating each .conf file.
func ProcessConfDirectory(directory, BACKEND_IP, PERS_BACKEND_IP, DELPHI_BACKEND_IP, BASE_DOMAIN string) error {
    return filepath.WalkDir(directory, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }
        if !d.IsDir() && strings.HasSuffix(d.Name(), ".conf") {
            UpdateFile(path, BACKEND_IP, PERS_BACKEND_IP, DELPHI_BACKEND_IP, BASE_DOMAIN)
        }
        return nil
    })
}

//
// ------------------ USER INTERACTION ------------------
//

// PromptInput prompts the user with a message and returns the input or default value.
func PromptInput(varName, promptMessage, defaultVal string) string {
    reader := bufio.NewReader(os.Stdin)
    for {
        if defaultVal != "" {
            fmt.Printf("%s [%s]: ", promptMessage, defaultVal)
        } else {
            fmt.Printf("%s: ", promptMessage)
        }
        in, _ := reader.ReadString('\n')
        in = strings.TrimSpace(in)
        if in == "" && defaultVal != "" {
            return defaultVal
        } else if in != "" {
            return in
        }
        fmt.Printf("Error: %s cannot be empty. Please enter a valid value.\n", varName)
    }
}
