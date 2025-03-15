// pkg/utils/utils.go
package utils

import (
    "bufio"
    "fmt"
    "io"
    "io/fs"
    "os"
    "path/filepath"
    "strings"
    "time"
)

// Constants that define your config file and directory to process.
const (
    LastValuesFile = ".hecate.conf"
    ConfDir        = "conf.d"
)

// restoreDir removes dstDir and copies backupDir -> dstDir
func RestoreDir(backupDir, dstDir string) {
	info, err := os.Stat(backupDir)
	if err != nil || !info.IsDir() {
		fmt.Printf("Error: Backup directory '%s' does not exist or is not a directory.\n", backupDir)
		os.Exit(1)
	}
	if err := utils.RemoveIfExists(dstDir); err != nil {
		fmt.Printf("Error removing %s: %v\n", dstDir, err)
		os.Exit(1)
	}
	if err := utils.CopyDir(backupDir, dstDir); err != nil {
		fmt.Printf("Error during restore of %s: %v\n", backupDir, err)
		os.Exit(1)
	}
	fmt.Printf("Restore complete: '%s' has been restored to '%s'.\n", backupDir, dstDir)
}

// restoreFile removes dstFile and copies backupFile -> dstFile
func RestoreFile(backupFile, dstFile string) {
	info, err := os.Stat(backupFile)
	if err != nil || info.IsDir() {
		fmt.Printf("Error: Backup file '%s' does not exist or is a directory.\n", backupFile)
		os.Exit(1)
	}
	if err := utils.RemoveIfExists(dstFile); err != nil {
		fmt.Printf("Error removing %s: %v\n", dstFile, err)
		os.Exit(1)
	}
	if err := utils.CopyFile(backupFile, dstFile); err != nil {
		fmt.Printf("Error during restore of %s: %v\n", backupFile, err)
		os.Exit(1)
	}
	fmt.Printf("Restore complete: '%s' has been restored to '%s'.\n", backupFile, dstFile)
}

// findLatestBackup finds the lexicographically greatest file whose name starts with `prefix` and ends with ".bak".
// For example, if prefix is "conf.d.", it might find "conf.d.20250325-101010.bak".
func FindLatestBackup(prefix string) (string, error) {
	entries, err := os.ReadDir(".")
	if err != nil {
		return "", err
	}
	var latest string
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, prefix) && strings.HasSuffix(name, ".bak") {
			// For purely numeric timestamps, a lexicographical comparison is enough to find the "largest" date/time.
			// e.g. conf.d.20250325-101010.bak > conf.d.20250325-100000.bak
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
        // Create backup first.
        if err := BackupFile(path); err != nil {
            fmt.Printf("Error creating backup for %s: %v\n", path, err)
            return
        }
        // Write new content.
        err = os.WriteFile(path, []byte(newContent), 0644)
        if err != nil {
            fmt.Printf("Error writing %s: %v\n", path, err)
            return
        }
        fmt.Printf("Updated %s\n", path)
    }
}

// processConfDirectory walks through directory recursively, updating each .conf file.
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

// promptInput prompts the user with a message and returns the input or default value.
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

// saveLastValues writes key="value" lines to LastValuesFile.
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

// loadLastValues reads key="value" lines from LastValuesFile.
func LoadLastValues() (map[string]string, error) {
    values := make(map[string]string)
    file, err := os.Open(LastValuesFile)
    if err != nil {
        // If file doesn't exist, return empty map.
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
        value := strings.TrimSpace(parts[1])
        // Remove surrounding quotes.
        value = strings.Trim(value, `"`)
        values[key] = value
    }
    return values, scanner.Err()
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
    defer func() {
        _ = out.Close()
    }()

    if _, err = io.Copy(out, in); err != nil {
        return err
    }

    // Ensure the copied file has the same permissions.
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

    // Create destination directory.
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
