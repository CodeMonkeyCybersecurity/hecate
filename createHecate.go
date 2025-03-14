package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Utility represents a menu item
type Utility struct {
	key    string
	desc   string
	script string
}

func printMenu(utilities []Utility) {
	fmt.Println("\n--- Deploy Hecate Utility Wrapper ---\n")
	for _, util := range utilities {
		fmt.Printf("%s. %s\n", util.key, util.desc)
	}
	fmt.Println()
}

func runScript(scriptPath string) {
	// Check if the script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		fmt.Printf("Error: %s not found.\n", scriptPath)
		return
	}

	// Execute the script using Python3
	cmd := exec.Command("python3", scriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// If the error is from the command exiting with a non-zero status
		if exitError, ok := err.(*exec.ExitError); ok {
			fmt.Printf("Error: Script %s exited with error code %d\n", scriptPath, exitError.ExitCode())
		} else {
			fmt.Printf("An error occurred while running %s: %v\n", scriptPath, err)
		}
	}
}

func main() {
	utilities := []Utility{
		{"1", "Create Backup", "utilities/createBackup.go"},
		{"2", "Create Config Variables", "utilities/creteConfigVariables.go"},
		{"3", "Create EOS Apps", "utilities/createEosApps.go"},
		{"4", "Create http.conf", "utilities/createHttpConf.go"},
		{"5", "Create Docker Compose", "utilities/createDockerCompose.go"},
		{"6", "Create Certificates", "utilities/createCerts.go"},
		{"7", "Restore Config", "utilities/createRestore.go"},
		{"q", "Quit", ""},
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		printMenu(utilities)
		fmt.Print("Enter the number of the utility to run (or 'q' to quit): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input, please try again.")
			continue
		}

		choice := strings.TrimSpace(input)
		if strings.ToLower(choice) == "q" {
			fmt.Println("Exiting deployHecate. Goodbye!")
			break
		}

		var found bool
		for _, util := range utilities {
			if util.key == choice {
				found = true
				fmt.Printf("\nRunning '%s' from %s...\n\n", util.desc, util.script)
				runScript(util.script)
				break
			}
		}

		if !found {
			fmt.Println("Invalid selection. Please try again.")
		}
	}
}
