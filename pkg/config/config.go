// pkg/config/config.go
package config

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

//
// ---------------------------- CONSTANTS ---------------------------- //
//

// Constants for file and directory names.
const (
	LastValuesFile    	= ".hecate.conf"
	DefaultComposeYML 	= "docker-compose.yml"
	DefaultCertsDir   	= "certs"
	DefaultConfDir    	= "conf.d"
	AssetsPath        	= "assets"
	NginxConfPath     	= "/etc/nginx/conf.d/"
	NginxStreamPath   	= "/etc/nginx/stream.d/"
	DockerNetworkName 	= "arachne-net"
	DockerIPv4Subnet  	= "10.1.0.0/16"
	DockerIPv6Subnet  	= "fd42:1a2b:3c4d:5e6f::/64"
	DefaultConfigPath 	= "./config/default.yaml"
	AssetServerPath 	= "assets/servers"
    	AssetStreamPath 	= "assets/stream"
)

// DefaultMarkers holds the default port markers that apply to all apps.
var DefaultMarkers = []string{"80", "443"}

// CombineMarkers merges DefaultMarkers with additional markers.
func CombineMarkers(additional ...string) []string {
	return append(DefaultMarkers, additional...)
}

//
// ---------------------------- APPLICATION CONFIGURATION ---------------------------- //
//

// App represents an application option.
type App struct {
	Option   string // Option number as a string.
	Name     string
	ConfFile string
	Markers  []string
}

// GetSupportedAppNames returns a list of supported application names.
func GetSupportedAppNames() []string {
	var names []string
	for _, app := range Apps {
		names = append(names, strings.ToLower(app.Name)) // Normalize names to lowercase
	}
	return names
}

// Apps holds all available application options.
var Apps = []App{
	{"1", "Static website", "base.conf", DefaultMarkers},
	{"2", "Wazuh", "delphi.conf", CombineMarkers("1515", "1514", "55000")},
	{"3", "Mattermost", "collaborate.conf", DefaultMarkers},
	{"4", "Nextcloud", "cloud.conf", CombineMarkers("3478", "coturn:")},
	{"5", "Mailcow", "mailcow.conf", CombineMarkers("25", "587", "465", "110", "995", "143", "993")},
	{"6", "Jenkins", "jenkins.conf", DefaultMarkers},
	{"7", "Grafana", "observe.conf", DefaultMarkers},
	{"8", "Umami", "analytics.conf", DefaultMarkers},
	{"9", "MinIO", "s3.conf", DefaultMarkers},
	{"10", "Wiki.js", "wiki.conf", DefaultMarkers},
	{"11", "ERPNext", "erp.conf", DefaultMarkers},
	{"12", "Jellyfin", "jellyfin.conf", DefaultMarkers},
	{"13", "Persephone", "persephone.conf", DefaultMarkers},
}

//
// ---------------------------- FUNCTIONS ---------------------------- //
//

// DisplayOptions prints the available application options.
func DisplayOptions() {
	fmt.Println("Available Hecate backend web apps:")
	var sortedApps []int
	for _, app := range Apps {
		if num, err := strconv.Atoi(app.Option); err == nil {
			sortedApps = append(sortedApps, num)
		}
	}
	sort.Ints(sortedApps)
	for _, num := range sortedApps {
		for _, app := range Apps {
			if app.Option == strconv.Itoa(num) {
				fmt.Printf("  %s. %s -> %s\n", app.Option, app.Name, app.ConfFile)
				break
			}
		}
	}
}

// GetAppByOption returns the App corresponding to a given option string.
func GetAppByOption(option string) (App, bool) {
	for _, app := range Apps {
		if app.Option == option {
			return app, true
		}
	}
	return App{}, false
}

// GetUserSelection prompts the user for a comma-separated list of option numbers.
// It returns a map (keyed by lowercase app name) of the selected Apps and the raw selection string.
func GetUserSelection(defaultSelection string) (map[string]App, string) {
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

	selectedApps := make(map[string]App)
	if strings.ToLower(selection) == "all" {
		for _, app := range Apps {
			selectedApps[strings.ToLower(app.Name)] = app
		}
		return selectedApps, "all"
	}

	parts := strings.Split(selection, ",")
	for _, token := range parts {
		token = strings.TrimSpace(token)
		app, ok := GetAppByOption(token)
		if !ok {
			fmt.Printf("Invalid option: %s\n", token)
			return GetUserSelection(defaultSelection)
		}
		selectedApps[strings.ToLower(app.Name)] = app
	}
	if len(selectedApps) == 0 {
		fmt.Println("No valid options selected.")
		return GetUserSelection(defaultSelection)
	}
	return selectedApps, selection
}


//
// ---------------------------- HECATE CONFIGURATION MANAGEMENT ---------------------------- //
//

// HecateConfig holds the primary configuration values.
type HecateConfig struct {
	BaseDomain string
	BackendIP  string
}

// LoadConfig reads LastValuesFile (.hecate.conf) and returns the configuration.
// If the file does not exist or the values need to be updated, it prompts the user.
func LoadConfig() (*HecateConfig, error) {
	configPath := LastValuesFile
	cfg := &HecateConfig{}

	// Check if file exists.
	if _, err := os.Stat(configPath); err == nil {
		// Read existing config.
		f, err := os.Open(configPath)
		if err != nil {
			return nil, fmt.Errorf("unable to open %s: %w", configPath, err)
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "BASE_DOMAIN=") {
				cfg.BaseDomain = strings.TrimSpace(strings.TrimPrefix(line, "BASE_DOMAIN="))
			} else if strings.HasPrefix(line, "backendIP=") {
				cfg.BackendIP = strings.TrimSpace(strings.TrimPrefix(line, "backendIP="))
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("error reading %s: %w", configPath, err)
		}
	}

	// Show current configuration.
	fmt.Printf("Current configuration:\n  BASE_DOMAIN: %s\n  backendIP: %s\n", cfg.BaseDomain, cfg.BackendIP)
	// Ask user whether to keep these values. If the file is missing or user declines, prompt for new values.
	if !yesOrNo("Do you want to keep these values? (Y/n): ") || cfg.BaseDomain == "" || cfg.BackendIP == "" {
		cfg.BaseDomain = prompt("Enter new BASE_DOMAIN: ")
		cfg.BackendIP = prompt("Enter new backendIP: ")
	}

	// Write (or overwrite) configuration.
	content := fmt.Sprintf("BASE_DOMAIN=%s\nbackendIP=%s\n", cfg.BaseDomain, cfg.BackendIP)
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write %s: %w", configPath, err)
	}

	return cfg, nil
}

// prompt reads a line from standard input after displaying the provided message.
func prompt(message string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(message)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

// yesOrNo asks the user a yes/no question and returns true if the answer is yes (default yes).
func yesOrNo(message string) bool {
	response := prompt(message)
	if response == "" {
		return true // default yes
	}
	response = strings.ToLower(response)
	return response == "y" || response == "yes"
}
