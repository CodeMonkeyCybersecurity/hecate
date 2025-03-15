// pkg/utils/utils.go
package utils

import (
    "bufio"
    "fmt"
    "io"
    "io/fs"
    "os"
    "path/filepath"
    "sort"
    "strconv"
    "strings"
    "time"
)

// Constants that define your config file and directory to process.
const (
    LastValuesFile    = ".hecate.conf"
    ConfDir           = "conf.d"
    DockerComposeFile = "docker-compose.yml"
)

// App represents an application option.
type App struct {
	Option   string // the option number as a string
	Name     string
	ConfFile string
	Markers  []string
}

// defaultMarkers holds the default port markers that apply to all apps.
var DefaultMarkers = []string{"80", "443"}

// combineMarkers returns a new slice containing the default markers
// plus any additional markers passed in.
func CombineMarkers(additional ...string) []string {
	markers := make([]string, len(defaultMarkers))
	copy(markers, defaultMarkers)
	markers = append(markers, additional...)
	return markers
}

// APPS_SELECTION maps option numbers (as strings) to their app name and configuration file.
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

// updateComposeFile reads the docker-compose file and, for each line that contains any marker
// from a selected app, removes the leading '#' so that the line becomes active.
func UpdateComposeFile(selectedApps map[string]App) error {
	content, err := os.ReadFile(DockerComposeFile)
	if err != nil {
		return fmt.Errorf("Error: %s not found", DockerComposeFile)
	}
	lines := strings.Split(string(content), "\n")
	// Regex to remove leading '#' and any spaces following it.
	uncommentRegex := regexp.MustCompile(`^(\s*)#\s*`)
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

// DisplayOptions prints the available options from APPS_SELECTION.
func DisplayOptions() {
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
        fmt.Printf("  %s. %s -> %s\n", k, option.AppName, option.ConfFile)
    }
}

// GetUserSelection prompts the user for a comma-separated list of options.
// Returns a set (map[string]bool) of allowed configuration filenames and the raw selection string.
func GetUserSelection(defaultSelection string) (map[string]bool, string) {
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
    return GetUserSelection(defaultSelection)
}

// RemoveUnwantedConfFiles walks through ConfDir and deletes any .conf file whose base name is not in allowedFiles.
func RemoveUnwantedConfFiles(allowedFiles map[string]bool) {
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

// SaveLastValues writes key="value" lines to LastValuesFile.
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

// LoadLastValues reads key="value" lines from LastValuesFile.
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
        value := strings.TrimSpace(parts[1])
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
