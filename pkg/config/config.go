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

// Constants for file and directory names.
const (
	LastValuesFile    = ".hecate.conf"
	DockerComposeFile = "docker-compose.yml"
	assetsPath := "assets"
	nginxConfPath := "/etc/nginx/conf.d/"
	nginxStreamPath := "/etc/nginx/stream.d/"
)

// DefaultMarkers holds the default port markers that apply to all apps.
var DefaultMarkers = []string{"80", "443"}

// CombineMarkers returns a new slice containing the default markers plus any additional markers.
func CombineMarkers(additional ...string) []string {
	markers := make([]string, len(DefaultMarkers))
	copy(markers, DefaultMarkers)
	markers = append(markers, additional...)
	return markers
}

// App represents an application option.
type App struct {
	Option   string // Option number as a string.
	Name     string
	ConfFile string
	Markers  []string
}

// Apps holds all available application options.
var Apps = []App{
	{Option: "1", Name: "Static website", ConfFile: "base.conf", Markers: DefaultMarkers},
	{Option: "2", Name: "Wazuh", ConfFile: "delphi.conf", Markers: CombineMarkers("1515", "1514", "55000")},
	{Option: "3", Name: "Mattermost", ConfFile: "collaborate.conf", Markers: DefaultMarkers},
	{Option: "4", Name: "Nextcloud", ConfFile: "cloud.conf", Markers: CombineMarkers("3478", "coturn:")},
	{Option: "5", Name: "Mailcow", ConfFile: "mailcow.conf", Markers: CombineMarkers("25", "587", "465", "110", "995", "143", "993")},
	{Option: "6", Name: "Jenkins", ConfFile: "jenkins.conf", Markers: DefaultMarkers},
	{Option: "7", Name: "Grafana", ConfFile: "observe.conf", Markers: DefaultMarkers},
	{Option: "8", Name: "Umami", ConfFile: "analytics.conf", Markers: DefaultMarkers},
	{Option: "9", Name: "MinIO", ConfFile: "s3.conf", Markers: DefaultMarkers},
	{Option: "10", Name: "Wiki.js", ConfFile: "wiki.conf", Markers: DefaultMarkers},
	{Option: "11", Name: "ERPNext", ConfFile: "erp.conf", Markers: DefaultMarkers},
	{Option: "12", Name: "Jellyfin", ConfFile: "jellyfin.conf", Markers: DefaultMarkers},
	{Option: "13", Name: "Persephone", ConfFile: "persephone.conf", Markers: DefaultMarkers},
}

// DisplayOptions prints the available application options.
func DisplayOptions() {
	fmt.Println("Available EOS backend web apps:")
	// Sort by Option number.
	var keys []int
	for _, app := range Apps {
		if num, err := strconv.Atoi(app.Option); err == nil {
			keys = append(keys, num)
		}
	}
	sort.Ints(keys)
	for _, num := range keys {
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
