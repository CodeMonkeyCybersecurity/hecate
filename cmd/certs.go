/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// certsCmd represents the certs command.
var certsCmd = &cobra.Command{
	Use:   "certs",
	Short: "Create certificates",
	Long:  `This command generates the necessary certificates for Hecate.`,
	Run: func(cmd *cobra.Command, args []string) {
		// All certificate creation logic goes here
		fmt.Println("Creating certificates...")

		// Constants
		const LAST_VALUES_FILE = ".hecate.conf"
		// 1. Check Docker processes and stop Hecate.
		fmt.Println("Checking Docker processes...")
		if err := runCommand([]string{"docker", "ps"}, false); err != nil {
			fmt.Printf("Error checking Docker processes: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Stopping Hecate...")
		if err := runCommand([]string{"docker", "compose", "down"}, false); err != nil {
			fmt.Println("Warning: Docker compose down failed (likely because there is no Hecate container up). Continuing...")
		}

		// 2. Load previous values if available.
		prevValues, err := loadLastValues(LAST_VALUES_FILE)
		if err != nil {
			fmt.Printf("Error loading previous values: %v\n", err)
			os.Exit(1)
		}
		baseDomain := promptInput("Enter the base domain (e.g. domain.com)", prevValues["BASE_DOMAIN"])
		subdomain := promptSubdomain()
		mailCert := promptInput("Enter the contact email (e.g. example@domain.com)", prevValues["EMAIL"])

		// Save the entered values for future runs.
		newValues := map[string]string{
			"BASE_DOMAIN": baseDomain,
			"EMAIL":       mailCert,
		}
		if err := saveLastValues(LAST_VALUES_FILE, newValues); err != nil {
			fmt.Printf("Error saving values: %v\n", err)
			os.Exit(1)
		}

		// 3. Combine to form the full domain.
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
		if err := runCommand(certbotCommand, false); err != nil {
			fmt.Printf("Error running certbot: %v\n", err)
			os.Exit(1)
		}

		// 5. Verify certificates are present.
		certPath := fmt.Sprintf("/etc/letsencrypt/live/%s/", fullDomain)
		fmt.Printf("Verifying that the certificates are in '%s'...\n", certPath)
		if err := runCommand([]string{"sudo", "ls", "-l", certPath}, false); err != nil {
			fmt.Printf("Error verifying certificates: %v\n", err)
			os.Exit(1)
		}

		// 6. Change directory to /opt/hecate and ensure certs/ exists.
		hecateDir := "/opt/hecate"
		if err := os.Chdir(hecateDir); err != nil {
			fmt.Printf("Error changing directory to %s: %v\n", hecateDir, err)
			os.Exit(1)
		}
		if err := os.MkdirAll("certs", 0755); err != nil {
			fmt.Printf("Error creating certs directory: %v\n", err)
			os.Exit(1)
		}

		// 7. Ask user to confirm certificate name.
		defaultCertName := baseDomain
		if subdomain != "" {
			defaultCertName = subdomain
		}
		certName := confirmCertName(defaultCertName)

		// 8. Copy certificate files.
		sourceFullchain := fmt.Sprintf("/etc/letsencrypt/live/%s/fullchain.pem", fullDomain)
		sourcePrivkey := fmt.Sprintf("/etc/letsencrypt/live/%s/privkey.pem", fullDomain)
		destFullchain := filepath.Join("certs", fmt.Sprintf("%s.fullchain.pem", certName))
		destPrivkey := filepath.Join("certs", fmt.Sprintf("%s.privkey.pem", certName))

		fmt.Println("Copying certificate files...")
		if err := runCommand([]string{"sudo", "cp", sourceFullchain, destFullchain}, false); err != nil {
			fmt.Printf("Error copying fullchain.pem: %v\n", err)
			os.Exit(1)
		}
		if err := runCommand([]string{"sudo", "cp", sourcePrivkey, destPrivkey}, false); err != nil {
			fmt.Printf("Error copying privkey.pem: %v\n", err)
			os.Exit(1)
		}

		// 9. Set appropriate permissions.
		fmt.Println("Setting appropriate permissions on the certificate files...")
		if err := runCommand([]string{"sudo", "chmod", "644", destFullchain}, false); err != nil {
			fmt.Printf("Error setting permissions on %s: %v\n", destFullchain, err)
			os.Exit(1)
		}
		if err := runCommand([]string{"sudo", "chmod", "600", destPrivkey}, false); err != nil {
			fmt.Printf("Error setting permissions on %s: %v\n", destPrivkey, err)
			os.Exit(1)
		}

		// 10. List the certs directory.
		fmt.Println("Listing the certs/ directory:")
		if err := runCommand([]string{"ls", "-lah", "certs/"}, false); err != nil {
			fmt.Printf("Error listing certs directory: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nYou should now have the appropriate certificates for https://%s\n", fullDomain)
		fmt.Println("Next, run ./updateConfigVariables.py and ./updateEosApps.py before (re)starting Hecate")
		fmt.Println("\nCertificates created successfully.")
	},
}

// Helper functions for the command

// runCommand runs a command and prints it. It exits on error.
func runCommand(command []string, shell bool) error {
	if shell {
		fmt.Printf("Running command: %s\n", strings.Join(command, " "))
		cmd := exec.Command("sh", "-c", strings.Join(command, " "))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	} else {
		fmt.Printf("Running command: %s\n", strings.Join(command, " "))
		cmd := exec.Command(command[0], command[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
}

// loadLastValues loads key=value pairs from a file.
func loadLastValues(filename string) (map[string]string, error) {
	values := make(map[string]string)
	file, err := os.Open(filename)
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
		// Remove surrounding quotes if any.
		value = strings.Trim(value, `"`)
		values[key] = value
	}
	return values, scanner.Err()
}

// saveLastValues writes key=value pairs to a file.
func saveLastValues(filename string, values map[string]string) error {
	file, err := os.Create(filename)
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

// promptInput prompts the user with a message. If defaultVal is provided, it is shown.
func promptInput(promptMessage, defaultVal string) string {
	reader := bufio.NewReader(os.Stdin)
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", promptMessage, defaultVal)
	} else {
		fmt.Printf("%s: ", promptMessage)
	}
	in, _ := reader.ReadString('\n')
	in = strings.TrimSpace(in)
	if in == "" && defaultVal != "" {
		return defaultVal
	}
	for in == "" {
		fmt.Println("Input cannot be empty. Please try again.")
		fmt.Printf("%s: ", promptMessage)
		in, _ = reader.ReadString('\n')
		in = strings.TrimSpace(in)
	}
	return in
}

// promptSubdomain prompts for the subdomain.
func promptSubdomain() string {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter the subdomain to configure (e.g. sub). Leave blank if none: ")
		subdomain, _ := reader.ReadString('\n')
		subdomain = strings.TrimSpace(subdomain)
		if subdomain == "" {
			fmt.Print("You entered a blank subdomain. Do you wish to continue with no subdomain? (yes/no): ")
			confirm, _ := reader.ReadString('\n')
			confirm = strings.ToLower(strings.TrimSpace(confirm))
			if confirm == "yes" || confirm == "y" {
				return ""
			}
			continue
		} else {
			return subdomain
		}
	}
}

// confirmCertName asks the user to confirm or change the default certificate name.
func confirmCertName(defaultName string) string {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("Use certificate name '%s'? (yes/no): ", defaultName)
		confirm, _ := reader.ReadString('\n')
		confirm = strings.ToLower(strings.TrimSpace(confirm))
		if confirm == "yes" || confirm == "y" {
			return defaultName
		} else if confirm == "no" || confirm == "n" {
			return promptInput("Enter the desired certificate name (for file naming)", "")
		} else {
			fmt.Println("Please answer yes or no.")
		}
	}
}

func init() {
	// Attach this subcommand to the parent create command.
	createCmd.AddCommand(certsCmd)
}
