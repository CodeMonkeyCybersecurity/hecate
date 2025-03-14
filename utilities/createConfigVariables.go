package main

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

const (
	LastValuesFile = ".hecate.conf"
	ConfDir        = "conf.d"
)

// loadLastValues reads key="value" lines from LastValuesFile.
func loadLastValues() (map[string]string, error) {
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

// saveLastValues writes key="value" lines to LastValuesFile.
func saveLastValues(values map[string]string) error {
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

// promptInput prompts the user with a message. If defaultVal is provided,
// it is displayed and returned if the user enters nothing.
func promptInput(varName, promptMessage, defaultVal string) string {
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

// backupFile creates a backup of the given file by appending a timestamp.
func backupFile(filepath string) error {
	info, err := os.Stat(filepath)
	if err != nil || info.IsDir() {
		return nil
	}
	timestamp := time.Now().Format("20060102-150405")
	dir := filepath[:strings.LastIndex(filepath, string(os.PathSeparator))]
	base := filepath[strings.LastIndex(filepath, string(os.PathSeparator))+1:]
	backupPath := filepath.Join(dir, fmt.Sprintf("%s_%s.bak", timestamp, base))
	
	// Copy file content to backupPath.
	in, err := os.Open(filepath)
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

	fmt.Printf("Backup of '%s' created as '%s'.\n", filepath, backupPath)
	return nil
}

// updateFile reads a file, replaces placeholders, and if changes occur, creates a backup and writes the new content.
func updateFile(filepath, BACKEND_IP, PERS_BACKEND_IP, DELPHI_BACKEND_IP, BASE_DOMAIN string) {
	original, err := os.ReadFile(filepath)
	if err != nil {
		fmt.Printf("Error reading %s: %v\n", filepath, err)
		return
	}
	content := string(original)
	newContent := strings.ReplaceAll(content, "${BACKEND_IP}", BACKEND_IP)
	newContent = strings.ReplaceAll(newContent, "${PERS_BACKEND_IP}", PERS_BACKEND_IP)
	newContent = strings.ReplaceAll(newContent, "${DELPHI_BACKEND_IP}", DELPHI_BACKEND_IP)
	newContent = strings.ReplaceAll(newContent, "${BASE_DOMAIN}", BASE_DOMAIN)

	if newContent != content {
		// Create backup first.
		if err := backupFile(filepath); err != nil {
			fmt.Printf("Error creating backup for %s: %v\n", filepath, err)
			return
		}
		// Write new content.
		err = os.WriteFile(filepath, []byte(newContent), 0644)
		if err != nil {
			fmt.Printf("Error writing %s: %v\n", filepath, err)
			return
		}
		fmt.Printf("Updated %s\n", filepath)
	}
}

// processConfDirectory walks through directory recursively, updating each .conf file.
func processConfDirectory(directory, BACKEND_IP, PERS_BACKEND_IP, DELPHI_BACKEND_IP, BASE_DOMAIN string) error {
	return filepath.WalkDir(directory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".conf") {
			updateFile(path, BACKEND_IP, PERS_BACKEND_IP, DELPHI_BACKEND_IP, BASE_DOMAIN)
		}
		return nil
	})
}

func main() {
	fmt.Println("=== Recursive conf.d Variable Updater ===\n")

	// Load previous values if available.
	lastValues, err := loadLastValues()
	if err != nil {
		fmt.Printf("Error loading previous values: %v\n", err)
		os.Exit(1)
	}

	// Prompt user for variables.
	BACKEND_IP := promptInput("BACKEND_IP", "Enter the backend IP address", lastValues["BACKEND_IP"])
	PERS_BACKEND_IP := promptInput("PERS_BACKEND_IP", "Enter the backend IP address for your Persephone backups", lastValues["PERS_BACKEND_IP"])
	DELPHI_BACKEND_IP := promptInput("DELPHI_BACKEND_IP", "Enter the backend IP address for your Delphi install", lastValues["DELPHI_BACKEND_IP"])
	BASE_DOMAIN := promptInput("BASE_DOMAIN", "Enter the base domain for your services", lastValues["BASE_DOMAIN"])

	// Save the values for future runs.
	newValues := map[string]string{
		"BACKEND_IP":       BACKEND_IP,
		"PERS_BACKEND_IP":  PERS_BACKEND_IP,
		"DELPHI_BACKEND_IP": DELPHI_BACKEND_IP,
		"BASE_DOMAIN":      BASE_DOMAIN,
	}
	if err := saveLastValues(newValues); err != nil {
		fmt.Printf("Error saving values: %v\n", err)
		os.Exit(1)
	}

	// Check that the conf.d directory exists.
	if info, err := os.Stat(ConfDir); err != nil || !info.IsDir() {
		fmt.Printf("Error: Directory '%s' not found in the current directory.\n", ConfDir)
		os.Exit(1)
	}

	// Process all .conf files in conf.d.
	if err := processConfDirectory(ConfDir, BACKEND_IP, PERS_BACKEND_IP, DELPHI_BACKEND_IP, BASE_DOMAIN); err != nil {
		fmt.Printf("Error processing configuration files: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nDone updating configuration files in the conf.d directory.")
}
