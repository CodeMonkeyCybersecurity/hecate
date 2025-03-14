package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// File constants.
const (
	LastValuesFile = ".hecate.conf"
	ConfDir        = "conf.d"
)

// AppOption holds an app's name and corresponding configuration file.
type AppOption struct {
	AppName  string
	ConfFile string
}

// APPS_SELECTION maps option numbers (as strings) to their app name and config file.
var APPS_SELECTION = map[string]AppOption{
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

// backupFile creates a backup of a file by copying it with a timestamp prefix.
func backupFile(filepath string) error {
	info, err := os.Stat(filepath)
	if err != nil || info.IsDir() {
		return nil
	}
	timestamp := time.Now().Format("20060102-150405")
	dir := filepath.Dir(filepath)
	base := filepath.Base(filepath)
	backupPath := filepath.Join(dir, fmt.Sprintf("%s_%s.bak", timestamp, base))
	input, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(backupPath, input, 0644)
	if err != nil {
		return err
	}
	fmt.Printf("Backup of '%s' created as '%s'.\n", filepath, backupPath)
	return nil
}

// displayOptions prints the available options from APPS_SELECTION.
func displayOptions() {
	fmt.Println("Available EOS backend web apps:")
	// Sort option numbers numerically.
	var keys []int
	for k := range APPS_SELECTION {
		if num, err := strconv.Atoi(k); err == nil {
			keys = append(keys, num)
		}
	}
	sort.Ints(keys)
	for _, num := range keys {
		k := strconv.Itoa(num)
		option := APPS_SELECTION[k]
		fmt.Printf("  %s. %s  -> %s\n", k, option.AppName, option.ConfFile)
	}
}

// getUserSelection prompts the user for a comma-separated list of options.
// Returns a set (as map[string]bool) of allowed configuration filenames and the raw selection string.
func getUserSelection(defaultSelection string) (map[string]bool, string) {
	reader := bufio.NewReader(os.Stdin)
	promptMsg := "Enter the numbers (comma-separated) of the apps you want enabled (or type 'all' for all)"
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
		allowed := make(map[string]bool)
		for _, option := range APPS_SELECTION {
			allowed[option.ConfFile] = true
		}
		return allowed, "all"
	}
	allowed := make(map[string]bool)
	valid := true
	parts := strings.Split(selection, ",")
	for _, token := range parts {
		token = strings.TrimSpace(token)
		option, exists := APPS_SELECTION[token]
		if !exists {
			fmt.Printf("Invalid option: %s\n", token)
			valid = false
			break
		}
		allowed[option.ConfFile] = true
	}
	if valid && len(allowed) > 0 {
		return allowed, selection
	}
	fmt.Println("Please enter a valid comma-separated list of options.")
	return getUserSelection(defaultSelection)
}

// removeUnwantedConfFiles walks through CONF_DIR and deletes any .conf file whose base name is not in allowedFiles.
func removeUnwantedConfFiles(allowedFiles map[string]bool) {
	info, err := os.Stat(ConfDir)
	if err != nil || !info.IsDir() {
		fmt.Printf("Error: Directory '%s' not found.\n", ConfDir)
		os.Exit(1)
	}
	var removedFiles []string
	// Walk the directory recursively.
	err = filepath.Walk(ConfDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".conf") {
			if !allowedFiles[info.Name()] {
				err := os.Remove(path)
				if err != nil {
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

func main() {
	fmt.Println("=== EOS Backend Web Apps Selector ===\n")
	lastValues, err := loadLastValues()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}
	// Use the previously saved APPS_SELECTION as default, if present.
	defaultApps := lastValues["APPS_SELECTION"]
	displayOptions()
	allowedFiles, selectionStr := getUserSelection(defaultApps)
	// Always preserve essential files.
	essential := []string{"http.conf", "stream.conf", "fallback.conf"}
	for _, fname := range essential {
		allowedFiles[fname] = true
	}
	fmt.Println("\nYou have selected the following configuration files to keep:")
	for file := range allowedFiles {
		if file == "http.conf" || file == "stream.conf" || file == "fallback.conf" {
			fmt.Printf(" - Essential file: %s\n", file)
		} else {
			// Look up the app name.
			for _, opt := range APPS_SELECTION {
				if opt.ConfFile == file {
					fmt.Printf(" - %s (%s)\n", opt.AppName, file)
					break
				}
			}
		}
	}
	fmt.Println("\nNow scanning the conf.d directory and removing files not in your selection...")
	removeUnwantedConfFiles(allowedFiles)
	// Save the selection back into .hecate.conf.
	lastValues["APPS_SELECTION"] = selectionStr
	if err := saveLastValues(lastValues); err != nil {
		fmt.Printf("Error saving configuration: %v\n", err)
	}
}
