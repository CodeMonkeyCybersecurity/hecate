package main

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// File constants.
const (
	LastValuesFile      = ".hecate.conf"
	DockerComposeFile   = "docker-compose.yml"
)

// AppOption maps an option number (as string) to an app name and config file.
type AppOption struct {
	AppName  string
	ConfFile string
}

// APP_OPTIONS holds the mapping from option numbers to app options.
var APP_OPTIONS = map[string]AppOption{
	"1":  {"Static website", "base.conf"},
	"2":  {"Wazuh", "delphi.conf"},
	"3":  {"Mattermost", "collaborate.conf"},
	"4":  {"Nextcloud", "cloud.conf"},
	"5":  {"Mailcow", "mailcow.conf"},
	"6":  {"Jenkins", "jenkins.conf"},
	"7":  {"Grafana", "observe.conf"},
	"8":  {"Umami", "analytics.conf"},
	"9":  {"MinIO", "s3.conf"},
	"10": {"Wiki.js", "wiki.conf"},
	"11": {"ERPNext", "erp.conf"},
	"12": {"Jellyfin", "jellyfin.conf"},
	"13": {"Persephone", "persephone.conf"},
}

// SUPPORTED_APPS maps app keywords (in lowercase) to a list of port markers.
var SUPPORTED_APPS = map[string][]string{
	"wazuh":     {"1515", "1514", "55000"},
	"mailcow":   {"25", "587", "465", "110", "995", "143", "993"},
	"nextcloud": {"3478"},
}

// loadLastValues reads key="value" pairs from LastValuesFile.
func loadLastValues() (map[string]string, error) {
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
		value := strings.TrimSpace(parts[1])
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

// backupFile makes a backup of the given filepath with a timestamp prefix.
func backupFile(filepath string) error {
	info, err := os.Stat(filepath)
	if err != nil || info.IsDir() {
		return nil
	}
	timestamp := time.Now().Format("20060102-150405")
	dir := filepathDir(filepath)
	base := filepathBase(filepath)
	backupPath := filepath.Join(dir, fmt.Sprintf("%s_%s.bak", timestamp, base))
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

// filepathDir and filepathBase are helpers for getting directory and base name.
func filepathDir(path string) string {
	return filepath.Dir(path)
}

func filepathBase(path string) string {
	return filepath.Base(path)
}

// displayOptions prints the available options.
func displayOptions() {
	fmt.Println("Available EOS backend web apps:")
	// Sort option numbers numerically.
	var keys []int
	for k := range APP_OPTIONS {
		if num, err := strconv.Atoi(k); err == nil {
			keys = append(keys, num)
		}
	}
	sort.Ints(keys)
	for _, num := range keys {
		k := strconv.Itoa(num)
		option := APP_OPTIONS[k]
		fmt.Printf("  %s. %s  -> %s\n", k, option.AppName, option.ConfFile)
	}
}

// getUserSelection prompts the user for a comma-separated list of option numbers.
// Returns the set of supported app keywords (from SUPPORTED_APPS) and the raw selection string.
func getUserSelection(defaultSelection string) (map[string]struct{}, string) {
	reader := bufio.NewReader(os.Stdin)
	promptMsg := "Enter the numbers (comma-separated) of the apps you want enabled (or type 'all' for all supported)"
	if defaultSelection != "" {
		promptMsg += fmt.Sprintf(" [default: %s]", defaultSelection)
	}
	promptMsg += ": "
	fmt.Print(promptMsg)
	selection, _ := reader.ReadString('\n')
	selection = strings.TrimSpace(selection)
	if selection == "" && defaultSelection != "" {
		selection = defaultSelection
	}
	if strings.ToLower(selection) == "all" {
		// Return all supported app keywords.
		all := make(map[string]struct{})
		for k := range SUPPORTED_APPS {
			all[k] = struct{}{}
		}
		return all, "all"
	}

	chosenKeywords := make(map[string]struct{})
	valid := true
	parts := strings.Split(selection, ",")
	for _, token := range parts {
		token = strings.TrimSpace(token)
		option, exists := APP_OPTIONS[token]
		if !exists {
			fmt.Printf("Invalid option: %s\n", token)
			valid = false
			break
		}
		// Map the app name (lowercase) to a supported keyword if applicable.
		key := strings.ToLower(option.AppName)
		if _, ok := SUPPORTED_APPS[key]; ok {
			chosenKeywords[key] = struct{}{}
		}
	}
	if valid && len(chosenKeywords) > 0 {
		return chosenKeywords, selection
	}
	fmt.Println("Please enter a valid comma-separated list of options corresponding to supported apps.")
	return getUserSelection(defaultSelection)
}

// updateComposeFile reads the docker-compose file, and for lines containing a marker
// for any selected app, removes the leading '#' before a dash.
func updateComposeFile(selectedApps map[string]struct{}) error {
	content, err := os.ReadFile(DockerComposeFile)
	if err != nil {
		return fmt.Errorf("Error: %s not found", DockerComposeFile)
	}
	lines := strings.Split(string(content), "\n")
	var newLines []string
	// Precompile a regex to remove leading '#' before a hyphen.
	re := regexp.MustCompile(`^(\s*)#\s*(-)`)
	for _, line := range lines {
		modifiedLine := line
		// For each supported app in selectedApps, check if any marker exists in the line.
		for app, markers := range SUPPORTED_APPS {
			if _, selected := selectedApps[app]; selected {
				for _, marker := range markers {
					if strings.Contains(line, marker) {
						modifiedLine = re.ReplaceAllString(line, "$1$2")
						// If modified, no need to test other markers.
						break
					}
				}
			}
		}
		newLines = append(newLines, modifiedLine)
	}
	// Backup the original docker-compose file.
	if err := backupFile(DockerComposeFile); err != nil {
		return err
	}
	// Write out the updated content.
	outContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(DockerComposeFile, []byte(outContent), 0644); err != nil {
		return err
	}
	// Create a slice of selected app keys for printing.
	var selApps []string
	for app := range selectedApps {
		selApps = append(selApps, app)
	}
	fmt.Printf("Updated %s for apps: %s\n", DockerComposeFile, strings.Join(selApps, ", "))
	return nil
}

func main() {
	// Load last values from config file.
	lastValues, err := loadLastValues()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	var selectedApps map[string]struct{}
	var selectionStr string

	// If command-line arguments are provided, use non-interactive mode.
	if len(os.Args) > 1 {
		selectedApps = make(map[string]struct{})
		for _, arg := range os.Args[1:] {
			lowArg := strings.ToLower(arg)
			if _, ok := SUPPORTED_APPS[lowArg]; ok {
				selectedApps[lowArg] = struct{}{}
			}
		}
		if len(selectedApps) == 0 {
			fmt.Println("No supported apps found in the command-line arguments.")
			os.Exit(1)
		}
		// Create a comma-separated string for record.
		var appsList []string
		for app := range selectedApps {
			appsList = append(appsList, app)
		}
		selectionStr = strings.Join(appsList, ", ")
	} else {
		// Interactive mode.
		displayOptions()
		defaultSelection := lastValues["APPS_SELECTION"]
		selectedApps, selectionStr = getUserSelection(defaultSelection)
		// Save the selection as the new default.
		lastValues["APPS_SELECTION"] = selectionStr
		if err := saveLastValues(lastValues); err != nil {
			fmt.Printf("Error saving configuration: %v\n", err)
		}
	}

	if err := updateComposeFile(selectedApps); err != nil {
		fmt.Printf("Error updating docker compose file: %v\n", err)
		os.Exit(1)
	}
}
