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

// AppOption maps an option number to an app name and its configuration file.
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

var SUPPORTED_APPS = map[string][]string{
	"static website": {"80", "443"},
	"mattermost":     {"80", "443"},
	"jenkins":        {"80", "443"},
	"grafana":        {"80", "443"},
	"umami":          {"80", "443"},
	"minio":          {"80", "443"},
	"wiki.js":        {"80", "443"},
	"jellyfin":       {"80", "443"},
	"persephone":     {"80", "443"},
	"wazuh":          {"1515", "1514", "55000"},
	"mailcow":        {"25", "587", "465", "110", "995", "143", "993"},
	"nextcloud":      {"3478", "5349", "49160-49200"},
}

// displayOptions prints the available options.
func displayOptions() {
	fmt.Println("Available EOS backend web apps:")
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
// It returns a set of supported app keywords and the raw selection string.
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
		// Map the app name (lowercase) to a supported keyword.
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

// updateComposeFile reads the docker-compose file and, for each line containing a marker
// corresponding to a selected app, removes the leading '#' before a dash.
func updateComposeFile(selectedApps map[string]struct{}) error {
	content, err := os.ReadFile(utils.DockerComposeFile)
	if err != nil {
		return fmt.Errorf("Error: %s not found", utils.DockerComposeFile)
	}
	lines := strings.Split(string(content), "\n")
	var newLines []string
	re := regexp.MustCompile(`^(\s*)#\s*(-)`)
	for _, line := range lines {
		modifiedLine := line
		for app, markers := range SUPPORTED_APPS {
			if _, selected := selectedApps[app]; selected {
				for _, marker := range markers {
					if strings.Contains(line, marker) {
						modifiedLine = re.ReplaceAllString(line, "$1$2")
						break
					}
				}
			}
		}
		newLines = append(newLines, modifiedLine)
	}
	// Backup the original docker-compose file.
	if err := utils.BackupFile(utils.DockerComposeFile); err != nil {
		return err
	}
	outContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(utils.DockerComposeFile, []byte(outContent), 0644); err != nil {
		return err
	}
	var selApps []string
	for app := range selectedApps {
		selApps = append(selApps, app)
	}
	fmt.Printf("Updated %s for apps: %s\n", utils.DockerComposeFile, strings.Join(selApps, ", "))
	return nil
}

// composeCmd represents the "compose" subcommand.
var composeCmd = &cobra.Command{
	Use:   "compose [app ...]",
	Short: "Update the docker-compose file",
	Long: `Update the docker-compose file by uncommenting port configuration lines 
associated with selected applications.

You can run this command in two modes:

1. Non-interactive mode:
   Supply one or more supported app keywords as arguments.
   Example: 
       hecate create compose nextcloud mailcow

2. Interactive mode:
   Run the command without valid app arguments, and you'll be prompted to choose.
   Example: 
       hecate create compose

Supported App Options:
  1. Static website    -> base.conf
  2. Wazuh             -> delphi.conf
  3. Mattermost        -> collaborate.conf
  4. Nextcloud         -> cloud.conf
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
by removing the leading '#' on the associated port lines. If no valid app options are 
provided, the command will exit with an error.`,
	Run: func(cmd *cobra.Command, args []string) {
		runCompose(args)
	},
}

func init() {
	createCmd.AddCommand(composeCmd)
}

func runCompose(args []string) {
	// Load previous values from configuration.
	lastValues, err := utils.LoadLastValues()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	var selectedApps map[string]struct{}
	var selectionStr string

	// Non-interactive mode: Use provided arguments.
	if len(args) > 0 {
		selectedApps = make(map[string]struct{})
		for _, arg := range args {
			lowArg := strings.ToLower(arg)
			if _, ok := SUPPORTED_APPS[lowArg]; ok {
				selectedApps[lowArg] = struct{}{}
			}
		}
		if len(selectedApps) == 0 {
			fmt.Println("No supported apps found in the command-line arguments.")
			os.Exit(1)
		}
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
		if err := utils.SaveLastValues(lastValues); err != nil {
			fmt.Printf("Error saving configuration: %v\n", err)
		}
	}

	if err := updateComposeFile(selectedApps); err != nil {
		fmt.Printf("Error updating docker-compose file: %v\n", err)
		os.Exit(1)
	}
}
