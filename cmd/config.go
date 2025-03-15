/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
// config.go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/CodeMonkeyCybersecurity/hecate/pkg/config"
    "github.com/CodeMonkeyCybersecurity/hecate/pkg/utils"
	"github.com/spf13/cobra"
)

// configCmd is the unified command to run configuration tasks.
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Run unified configuration tasks (certificates, docker-compose update, Eos and Nginx configuration)",
	Long: `This command performs various configuration tasks for Hecate:
  
  1) Create certificates
  2) Update the docker-compose file
  3) Configure EOS backend web apps 
  4) Configure Nginx defaults

You can choose to run one or all of these tasks interactively.`,
	Run: func(cmd *cobra.Command, args []string) {
		runUnifiedConfig()
	},
}

func init() {
	// Attach this command to the parent "create" command.
	createCmd.AddCommand(configCmd)
}

// runUnifiedConfig presents an interactive menu for the user.
func runUnifiedConfig() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("=== Unified Configuration Command ===")
	fmt.Println("Select the task you want to perform:")
	fmt.Println("1) Create Certificates")
	fmt.Println("2) Update docker-compose file")
	fmt.Println("3) Configure Eos backend web apps")
    fmt.Println("4) Configure Nginx defaults")
	fmt.Println("5) Run all tasks")
	fmt.Print("Enter choice (1-5): ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		runCerts()
	case "2":
		utils.RunComposeInteractive()
	case "3":
		runEos()
	case "4":
        runHttp()
	case "5":
		runCerts()
		utils.RunComposeInteractive()
		runEos()
        runHttp()
	default:
		fmt.Println("Invalid choice. Exiting.")
		os.Exit(1)
	}
}

//
// ------------------ CERTIFICATES FUNCTIONALITY ------------------
//

func runCerts() {
	fmt.Println("\n=== Certificate Creation ===")
	const LAST_VALUES_FILE = ".hecate.conf"

	// 1. Check Docker processes and stop Hecate.
	fmt.Println("Checking Docker processes...")
	if err := utils.RunCommand([]string{"docker", "ps"}, false); err != nil {
		fmt.Printf("Error checking Docker processes: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Stopping Hecate...")
	if err := utils.RunCommand([]string{"docker", "compose", "down"}, false); err != nil {
		fmt.Println("Warning: Docker compose down failed. Continuing...")
	}

	// 2. Load previous values.
	prevValues, err := utils.LoadLastValues(LAST_VALUES_FILE)
	if err != nil {
		fmt.Printf("Error loading previous values: %v\n", err)
		os.Exit(1)
	}

	baseDomain := utils.PromptInput("BASE_DOMAIN", "Enter the base domain (e.g. domain.com)", prevValues["BASE_DOMAIN"])
	subdomain := utils.PromptSubdomain()
	mailCert := utils.PromptInput("Enter the contact email (e.g. example@domain.com)", prevValues["EMAIL"])

	// Save the entered values for future runs.
	newValues := map[string]string{
		"BASE_DOMAIN": baseDomain,
		"EMAIL":       mailCert,
	}
	if err := utils.SaveLastValues(LAST_VALUES_FILE, newValues); err != nil {
		fmt.Printf("Error saving values: %v\n", err)
		os.Exit(1)
	}

	// 3. Form the full domain.
	var fullDomain string
	if subdomain != "" {
		fullDomain = fmt.Sprintf("%s.%s", subdomain, baseDomain)
	} else {
		fullDomain = baseDomain
	}
	fmt.Printf("\nThe full domain for certificate generation will be: %s\n", fullDomain)

	// 4. Run certbot to obtain certificate.
	certbotCommand := []string{
		"sudo", "certbot", "certonly", "--standalone",
		"-d", fullDomain,
		"--email", mailCert,
		"--agree-tos",
	}
	if err := utils.RunCommand(certbotCommand, false); err != nil {
		fmt.Printf("Error running certbot: %v\n", err)
		os.Exit(1)
	}

	// 5. Verify certificates.
	certPath := fmt.Sprintf("/etc/letsencrypt/live/%s/", fullDomain)
	fmt.Printf("Verifying that the certificates are in '%s'...\n", certPath)
	if err := utils.RunCommand([]string{"sudo", "ls", "-l", certPath}, false); err != nil {
		fmt.Printf("Error verifying certificates: %v\n", err)
		os.Exit(1)
	}

	// 6. Prepare the local certs directory.
	hecateDir := "/opt/hecate"
	if err := os.Chdir(hecateDir); err != nil {
		fmt.Printf("Error changing directory to %s: %v\n", hecateDir, err)
		os.Exit(1)
	}
	if err := os.MkdirAll("certs", 0755); err != nil {
		fmt.Printf("Error creating certs directory: %v\n", err)
		os.Exit(1)
	}

	// 7. Confirm certificate name.
	defaultCertName := baseDomain
	if subdomain != "" {
		defaultCertName = subdomain
	}
	certName := utils.ConfirmCertName(defaultCertName)

	// 8. Copy certificate files.
	sourceFullchain := fmt.Sprintf("/etc/letsencrypt/live/%s/fullchain.pem", fullDomain)
	sourcePrivkey := fmt.Sprintf("/etc/letsencrypt/live/%s/privkey.pem", fullDomain)
	destFullchain := filepath.Join("certs", fmt.Sprintf("%s.fullchain.pem", certName))
	destPrivkey := filepath.Join("certs", fmt.Sprintf("%s.privkey.pem", certName))

	fmt.Println("Copying certificate files...")
	if err := utils.RunCommand([]string{"sudo", "cp", sourceFullchain, destFullchain}, false); err != nil {
		fmt.Printf("Error copying fullchain.pem: %v\n", err)
		os.Exit(1)
	}
	if err := utils.RunCommand([]string{"sudo", "cp", sourcePrivkey, destPrivkey}, false); err != nil {
		fmt.Printf("Error copying privkey.pem: %v\n", err)
		os.Exit(1)
	}

	// 9. Set file permissions.
	fmt.Println("Setting permissions on the certificate files...")
	if err := utils.RunCommand([]string{"sudo", "chmod", "644", destFullchain}, false); err != nil {
		fmt.Printf("Error setting permissions on %s: %v\n", destFullchain, err)
		os.Exit(1)
	}
	if err := utils.RunCommand([]string{"sudo", "chmod", "600", destPrivkey}, false); err != nil {
		fmt.Printf("Error setting permissions on %s: %v\n", destPrivkey, err)
		os.Exit(1)
	}

	// 10. List the certs directory.
	fmt.Println("Listing the certs/ directory:")
	if err := utils.RunCommand([]string{"ls", "-lah", "certs/"}, false); err != nil {
		fmt.Printf("Error listing certs directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nCertificates created successfully for https://%s\n", fullDomain)
	fmt.Println("Next, run ./updateConfigVariables.py and ./updateEosApps.py before (re)starting Hecate")
}

//
// ------------------ EOS FUNCTIONALITY ------------------
//

func runEos() {
	fmt.Println("\n=== EOS Backend Web Apps Selector ===")
	const LAST_VALUES_FILE = ".hecate.conf"
	lastValues, err := utils.LoadLastValues(LAST_VALUES_FILE)
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}
	defaultApps := lastValues["APPS_SELECTION"]
	config.DisplayOptions()
	selectedApps, selectionStr := config.GetUserSelection(defaultApps)
	allowedFiles := make(map[string]bool)
	// Essential files that are always kept.
	essential := []string{"http.conf", "stream.conf", "fallback.conf"}
	for _, fname := range essential {
		allowedFiles[fname] = true
	}
	for _, app := range selectedApps {
		allowedFiles[app.ConfFile] = true
	}
	fmt.Println("\nYou have selected the following configuration files to keep:")
	for file := range allowedFiles {
		if file == "http.conf" || file == "stream.conf" || file == "fallback.conf" {
			fmt.Printf(" - Essential file: %s\n", file)
		} else {
			for _, opt := range config.Apps {
				if opt.ConfFile == file {
					fmt.Printf(" - %s (%s)\n", opt.Name, file)
					break
				}
			}
		}
	}
	fmt.Println("\nNow scanning the conf.d directory and removing files not in your selection...")
	if err :=utils.RemoveUnwantedConfFiles(allowedFiles); err != nil {
		fmt.Printf("Error removing unwanted conf files: %v\n", err)
		os.Exit(1)
	}
	lastValues["APPS_SELECTION"] = selectionStr
	if err := utils.SaveLastValues(LAST_VALUES_FILE, lastValues); err != nil {
		fmt.Printf("Error saving configuration: %v\n", err)
	}
}


//
// ------------------ HTTP FUNCTIONALITY ------------------
//

func runHttp() {
		configFile := "http.conf"
		backupFile := "http.conf.bak"

		fmt.Println("Welcome to the HTTP block configuration updater for your http.conf file.")
		fmt.Println("Below you'll see a description for each setting along with its current default value.")
		fmt.Println("Press Enter to keep the default value or type a new one as desired.\n")

		serverTokens := utils.PromptWithDefault(
			"Hide NGINX version (server_tokens)",
			"off",
			"server_tokens: When set to 'off', NGINX will not display its version in error pages and headers.",
		)

		includeMime := utils.PromptWithDefault(
			"Path to MIME types file (include)",
			"mime.types",
			"include: This file defines MIME types for various file extensions.",
		)

		defaultType := utils.PromptWithDefault(
			"Default MIME type (default_type)",
			"application/octet-stream",
			"default_type: The default MIME type for files with unknown extensions.",
		)

		errorLogPath := utils.PromptWithDefault(
			"Error log path",
			"/var/log/nginx/error.log",
			"error_log: The file path where NGINX will log error messages.",
		)

		errorLogLevel := utils.PromptWithDefault(
			"Error log level (e.g., warn, debug)",
			"warn",
			"Error log level: Determines the minimum severity of messages to be logged (e.g., 'warn' or 'debug').",
		)

		accessLogPath := utils.PromptWithDefault(
			"Access log path",
			"/var/log/nginx/access.log",
			"access_log: The file path where NGINX will log access details for incoming requests.",
		)

		sendfile := utils.PromptWithDefault(
			"Sendfile (on/off)",
			"on",
			"sendfile: When 'on', NGINX uses the sendfile system call to transfer files efficiently.",
		)

		includeServers := utils.PromptWithDefault(
			"Include server blocks directory/pattern",
			"/etc/nginx/conf.d/servers/*.conf",
			"include: Specifies the path or pattern for including server block configuration files.",
		)

		// Build the configuration content.
		configContent := fmt.Sprintf(`###
# HTTP BLOCK
###

http {
    # Hide NGINX version
    server_tokens   %s;
    include         %s;
    default_type    %s;

    error_log       %s %s;    # change warn to debug if installing a development server
    access_log      %s;       # enable access logging 
    sendfile        %s;
    
    ###
    # SERVER BLOCKS
    ###
    include          %s;
}
`, serverTokens, includeMime, defaultType, errorLogPath, errorLogLevel, accessLogPath, sendfile, includeServers)

		// Check if http.conf exists and back it up if necessary.
		if _, err := os.Stat(configFile); err == nil {
			// Remove existing backup if present.
			if _, err := os.Stat(backupFile); err == nil {
				if err := os.Remove(backupFile); err != nil {
					fmt.Printf("Error removing existing backup %s: %v\n", backupFile, err)
					os.Exit(1)
				}
			}
			// Rename current config file to backup.
			if err := os.Rename(configFile, backupFile); err != nil {
				fmt.Printf("Error backing up %s to %s: %v\n", configFile, backupFile, err)
				os.Exit(1)
			}
			fmt.Printf("\nExisting %s has been backed up to %s\n", configFile, backupFile)
		}

		// Write the new configuration to http.conf.
		f, err := os.Create(configFile)
		if err != nil {
			fmt.Printf("Error creating %s: %v\n", configFile, err)
			os.Exit(1)
		}
		defer f.Close()

		_, err = f.WriteString(configContent)
		if err != nil {
			fmt.Printf("Error writing to %s: %v\n", configFile, err)
			os.Exit(1)
		}

		fmt.Printf("\nNew configuration has been written to %s\n", configFile)
	},
}
