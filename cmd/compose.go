/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/CodeMonkeyCybersecurity/hecate/pkg/utils"
	"github.com/spf13/cobra"
)

// defaultMarkers holds the default port markers that apply to all apps.
var defaultMarkers = []string{"80", "443"}

// combineMarkers returns a new slice containing the default markers
// plus any additional markers passed in.
func combineMarkers(additional ...string) []string {
	markers := make([]string, len(defaultMarkers))
	copy(markers, defaultMarkers)
	markers = append(markers, additional...)
	return markers
}

// App represents an application option.
type App struct {
	Option   string // the option number as a string
	Name     string
	ConfFile string
	Markers  []string
}

// Apps holds all available app options.
var Apps = []App{
	{Option: "1", Name: "Static website", ConfFile: "base.conf", Markers: defaultMarkers},
	{Option: "2", Name: "Wazuh", ConfFile: "delphi.conf", Markers: combineMarkers("1515", "1514", "55000")},
	{Option: "3", Name: "Mattermost", ConfFile: "collaborate.conf", Markers: defaultMarkers},
	{Option: "4", Name: "Nextcloud", ConfFile: "cloud.conf", Markers: combineMarkers("3478", "coturn:")},
	{Option: "5", Name: "Mailcow", ConfFile: "mailcow.conf", Markers: combineMarkers("25", "587", "465", "110", "995", "143", "993")},
	{Option: "6", Name: "Jenkins", ConfFile: "jenkins.conf", Markers: defaultMarkers},
	{Option: "7", Name: "Grafana", ConfFile: "observe.conf", Markers: defaultMarkers},
	{Option: "8", Name: "Umami", ConfFile: "analytics.conf", Markers: defaultMarkers},
	{Option: "9", Name: "MinIO", ConfFile: "s3.conf", Markers: defaultMarkers},
	{Option: "10", Name: "Wiki.js", ConfFile: "wiki.conf", Markers: defaultMarkers},
	{Option: "11", Name: "ERPNext", ConfFile: "erp.conf", Markers: defaultMarkers},
	{Option: "12", Name: "Jellyfin", ConfFile: "jellyfin.conf", Markers: defaultMarkers},
	{Option: "13", Name: "Persephone", ConfFile: "persephone.conf", Markers: defaultMarkers},
}

// displayOptions prints the available app options.
func displayOptions() {
	fmt.Println("Available EOS backend web apps:")
	// Sort apps by Option
	sort.Slice(Apps, func(i, j int) bool {
		a, _ := strconv.Atoi(Apps[i].Option)
		b, _ := strconv.Atoi(Apps[j].Option)
		return a < b
	})
	for _, app := range Apps {
		fmt.Printf("  %s. %s -> %s\n", app.Option, app.Name, app.ConfFile)
	}
}

// getAppByOption returns the App corresponding to a given option string.
func getAppByOption(option string) (App, bool) {
	for _, app := range Apps {
		if app.Option == option {
			return app, true
		}
	}
	return App{}, false
}

// getUserSelection prompts the user for a comma-separated list of option numbers.
// It returns a map (keyed by lowercase app name) of the selected Apps and the raw selection.
func getUserSelection(defaultSelection string) (map[string]App, string) {
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
		app, ok := getAppByOption(token)
		if !ok {
			fmt.Printf("Invalid option: %s\n", token)
			return getUserSelection(defaultSelection)
		}
		selectedApps[strings.ToLower(app.Name)] = app
	}
	if len(selectedApps) == 0 {
		fmt.Println("No valid options selected.")
		return getUserSelection(defaultSelection)
	}
	return selectedApps, selection
}

// updateComposeFile reads the docker-compose file and, for each line that contains any marker
// from a selected app, removes the leading '#' so that the line becomes active.
func updateComposeFile(selectedApps map[string]App) error {
	content, err := os.ReadFile(utils.DockerComposeFile)
	if err != nil {
		return fmt.Errorf("Error: %s not found", utils.DockerComposeFile)
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
	if err := utils.BackupFile(utils.DockerComposeFile); err != nil {
		return err
	}
	outContent := strings.Join(lines, "\n")
	if err := os.WriteFile(utils.DockerComposeFile, []byte(outContent), 0644); err != nil {
		return err
	}
	var selApps []string
	for _, app := range selectedApps {
		selApps = append(selApps, app.Name)
	}
	fmt.Printf("Updated %s for apps: %s\n", utils.DockerComposeFile, strings.Join(selApps, ", "))
	return nil
}

// composeCmd represents the "compose" subcommand.
var composeCmd = &cobra.Command{
	Use:   "compose [app ...]",
	Short: "Update the docker-compose file",
	Long: `Update the docker-compose file by uncommenting configuration lines 
associated with selected applications.

You can run this command in two modes:

1. Non-interactive mode:
   Supply one or more supported app option numbers as arguments.
   Example: 
       hecate create compose 4 5

2. Interactive mode:
   Run the command without valid app arguments, and you'll be prompted to choose.
   Example: 
       hecate create compose

Supported App Options:
  1. Static website    -> base.conf
  2. Wazuh             -> delphi.conf
  3. Mattermost        -> collaborate.conf
  4. Nextcloud         -> cloud.conf   (uncomments coturn service for Nextcloud)
  5. Mailcow           -> mailcow.conf
  6. Jenkins           -> jenkins.conf
  7. Grafana           -> observe.conf
  8. Umami             -> analytics.conf
  9. MinIO             -> s3.conf
  10. Wiki.js          -> wiki.conf
  11. ERPNext          -> erp.conf
  12. Jellyfin         -> jellyfin.conf
  13. Persephone       -> persephone.conf

When a valid app option is selected, the command will update the docker-compose file 
by removing the leading '#' on lines that contain specific markers.
If no valid app options are provided, the command will exit with an error.`,
	Run: func(cmd *cobra.Command, args []string) {
		runCompose(args)
	},
}

func init() {
	createCmd.AddCommand(composeCmd)
}

func runCompose(args []string) {
	lastValues, err := utils.LoadLastValues()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	var selectedApps map[string]App
	var selectionStr string

	// Non-interactive mode: Use provided arguments.
	if len(args) > 0 {
		selectedApps = make(map[string]App)
		for _, arg := range args {
			app, ok := getAppByOption(arg)
			if ok {
				selectedApps[strings.ToLower(app.Name)] = app
			}
		}
		if len(selectedApps) == 0 {
			fmt.Println("No supported apps found in the command-line arguments.")
			os.Exit(1)
		}
		var appsList []string
		for _, app := range selectedApps {
			appsList = append(appsList, app.Name)
		}
		selectionStr = strings.Join(appsList, ", ")
	} else {
		// Interactive mode.
		displayOptions()
		defaultSelection := lastValues["APPS_SELECTION"]
		selectedApps, selectionStr = getUserSelection(defaultSelection)
		lastValues["APPS_SELECTION"] = selectionStr
		if err := utils.SaveLastValues(lastValues); err != nil {
			fmt.Printf("Error saving configuration: %v\n", err)
		}
	}

	if err := updateComposeFile(selectedApps); err != nil {
		fmt.Printf("Error updating docker-compose file: %v\n", err)
		os.Exit(1)
	}
}
