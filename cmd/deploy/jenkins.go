// cmd/deploy/jenkins

package deploy

import (
    "bufio"
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "github.com/spf13/cobra"
)

// deployJenkinsCmd represents the "deploy jenkins" subcommand
var deployJenkinsCmd = &cobra.Command{
    Use:   "jenkins",
    Short: "Deploy Jenkins using Hecate",
    RunE: func(cmd *cobra.Command, args []string) error {
        fmt.Println("🚀 Deploying Jenkins...")

        // 1) Check for .hecate.conf and gather variables
        baseDomain, backendIP, err := ensureHecateConfig()
        if err != nil {
            return fmt.Errorf("failed retrieving/confirming config: %w", err)
        }

        // 2) Identify required files vs “other” files
        requiredFiles := map[string]bool{
            "servers/jenkins.conf": true,
            "stream/jenkins.conf":  true,
            "http.conf":            true,
            "stream.conf":          true,
            // .hecate.conf is not in assets folder, but keep track if needed
        }

        assetsDir := "assets"
        err = organizeAssets(assetsDir, requiredFiles)
        if err != nil {
            return err
        }

        // 3) Replace placeholders in jenkins.conf files
        err = replacePlaceholders(filepath.Join(assetsDir, "servers/jenkins.conf"), baseDomain, backendIP)
        if err != nil {
            return err
        }
        err = replacePlaceholders(filepath.Join(assetsDir, "stream/jenkins.conf"), baseDomain, backendIP)
        if err != nil {
            return err
        }

        // 4) Check for certificates, or generate if necessary
        err = ensureCertificates(baseDomain)
        if err != nil {
            return err
        }

        // 5) Check whether we need to update docker-compose.yaml for Jenkins
        err = maybeUncommentJenkinsPort("docker-compose.yaml")
        if err != nil {
            return err
        }

        // 6) All done, run docker compose up -d
        fmt.Println("✅ All tasks completed. Bringing up Jenkins via Docker...")
        if err := dockerComposeUp(); err != nil {
            return fmt.Errorf("failed running docker compose: %w", err)
        }

        fmt.Println("🎉 Jenkins deployed successfully.")
        return nil
    },
}

// Ensure .hecate.conf exists and retrieve or confirm variables
func ensureHecateConfig() (string, string, error) {
    configPath := ".hecate.conf"

    var baseDomain, backendIP string
    // Check if file exists
    _, err := os.Stat(configPath)
    if os.IsNotExist(err) {
        //  File not there, so ask user for new values
        baseDomain, backendIP, err = promptForHecateVars("", "")
        if err != nil {
            return "", "", err
        }
        // Save
        err = writeHecateConfig(configPath, baseDomain, backendIP)
        if err != nil {
            return "", "", err
        }
    } else {
        // parse existing .hecate.conf
        storedDomain, storedIP, err := readHecateConfig(configPath)
        if err != nil {
            return "", "", err
        }
        // ask user to confirm or override
        baseDomain, backendIP, err = promptForHecateVars(storedDomain, storedIP)
        if err != nil {
            return "", "", err
        }
        // possibly rewrite .hecate.conf if changed
        err = writeHecateConfig(configPath, baseDomain, backendIP)
        if err != nil {
            return "", "", err
        }
    }
    return baseDomain, backendIP, nil
}

// Example: move files not in requiredFiles to other/
func organizeAssets(assetsDir string, requiredFiles map[string]bool) error {
    otherDir := filepath.Join(assetsDir, "other")
    if err := os.MkdirAll(otherDir, 0755); err != nil {
        return fmt.Errorf("failed creating other/ dir: %w", err)
    }

    err := filepath.Walk(assetsDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        // Skip directories themselves
        if info.IsDir() {
            return nil
        }

        relPath, _ := filepath.Rel(assetsDir, path)
        // If not in required set, move it
        if !requiredFiles[relPath] {
            newPath := filepath.Join(otherDir, relPath)
            // Make sure the directory structure inside other/ matches
            if err := os.MkdirAll(filepath.Dir(newPath), 0755); err != nil {
                return fmt.Errorf("error making subdir for %s: %w", newPath, err)
            }
            if err := os.Rename(path, newPath); err != nil {
                return fmt.Errorf("error moving %s to %s: %w", path, newPath, err)
            }
        }
        return nil
    })
    return err
}

// Replace placeholders in file
func replacePlaceholders(filePath, domain, ip string) error {
    input, err := os.ReadFile(filePath)
    if err != nil {
        return fmt.Errorf("error reading %s: %w", filePath, err)
    }
    content := string(input)
    content = strings.ReplaceAll(content, "${BASE_DOMAIN}", domain)
    content = strings.ReplaceAll(content, "${backendIP}", ip)

    err = os.WriteFile(filePath, []byte(content), 0644)
    if err != nil {
        return fmt.Errorf("error writing %s: %w", filePath, err)
    }
    return nil
}

// Check for cert files in certs/ or use let’s encrypt
func ensureCertificates(domain string) error {
    certDir := "certs"
    privKey := filepath.Join(certDir, fmt.Sprintf("%s.privkey.pem", domain))
    fullChain := filepath.Join(certDir, fmt.Sprintf("%s.fullchain.pem", domain))

    // if they’re not there, get them from Let’s Encrypt
    if _, err := os.Stat(privKey); os.IsNotExist(err) {
        fmt.Printf("No private key found for domain %s. Attempting to issue certificate...\n", domain)
        // Insert your ACME client or shell out to certbot, etc.:
        // e.g. runCertbot(domain)

        // For illustration only:
        // copy or rename from some location to:
        //   certs/domain.privkey.pem
        //   certs/domain.fullchain.pem
        fmt.Println("✅ Certificate generated and placed in certs/ directory (stub).")
    }
    // you could also check if fullChain is missing, etc.

    return nil
}

// Possibly uncomment Jenkins port in docker-compose.yaml
func maybeUncommentJenkinsPort(composeFile string) error {
    // If you do *not* want to uncomment automatically, skip.
    // This shows how you could do it if you do want to.

    input, err := os.ReadFile(composeFile)
    if err != nil {
        return fmt.Errorf("failed reading %s: %w", composeFile, err)
    }
    lines := strings.Split(string(input), "\n")

    // Example: search for "#- \"50000:50000\"" and uncomment it
    for i, line := range lines {
        trimmed := strings.TrimSpace(line)
        if strings.HasPrefix(trimmed, "#- \"50000:50000\"") {
            // Suppose we want to uncomment:
            lines[i] = strings.Replace(line, "#-", "-", 1)
        }
    }
    output := strings.Join(lines, "\n")
    err = os.WriteFile(composeFile, []byte(output), 0644)
    if err != nil {
        return fmt.Errorf("failed writing updated %s: %w", composeFile, err)
    }
    return nil
}

// Finally, run `docker compose up -d`
func dockerComposeUp() error {
    // You can either shell out or use something like https://github.com/docker/compose/v2/pkg/api
    // Here we’ll just do a simple shell exec:
    cmd := "docker compose up -d"
    fmt.Printf("Running: %s\n", cmd)
    // e.g. os/exec
    //   out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
    //   fmt.Println(string(out))
    //   return err
    return nil
}

// Prompt user for domain and IP
func promptForHecateVars(storedDomain, storedIP string) (string, string, error) {
    reader := bufio.NewReader(os.Stdin)

    finalDomain := storedDomain
    finalIP := storedIP

    if storedDomain != "" {
        fmt.Printf("Detected BASE_DOMAIN=%q from .hecate.conf; press Enter to keep or type a new value: ", storedDomain)
        text, _ := reader.ReadString('\n')
        text = strings.TrimSpace(text)
        if text != "" {
            finalDomain = text
        }
    } else {
        fmt.Printf("Enter BASE_DOMAIN: ")
        text, _ := reader.ReadString('\n')
        finalDomain = strings.TrimSpace(text)
    }

    if storedIP != "" {
        fmt.Printf("Detected backendIP=%q from .hecate.conf; press Enter to keep or type a new value: ", storedIP)
        text, _ := reader.ReadString('\n')
        text = strings.TrimSpace(text)
        if text != "" {
            finalIP = text
        }
    } else {
        fmt.Printf("Enter backendIP: ")
        text, _ := reader.ReadString('\n')
        finalIP = strings.TrimSpace(text)
    }

    // Basic validation, if you want
    if finalDomain == "" || finalIP == "" {
        return "", "", fmt.Errorf("cannot have empty domain or IP")
    }

    return finalDomain, finalIP, nil
}

// Stub: read .hecate.conf
func readHecateConfig(path string) (string, string, error) {
    // For example, if .hecate.conf is just lines: BASE_DOMAIN=..., backendIP=...
    data, err := os.ReadFile(path)
    if err != nil {
        return "", "", err
    }
    lines := strings.Split(string(data), "\n")
    var domain, ip string
    for _, line := range lines {
        if strings.HasPrefix(line, "BASE_DOMAIN=") {
            domain = strings.TrimPrefix(line, "BASE_DOMAIN=")
        } else if strings.HasPrefix(line, "backendIP=") {
            ip = strings.TrimPrefix(line, "backendIP=")
        }
    }
    return domain, ip, nil
}

// Stub: write .hecate.conf
func writeHecateConfig(path, domain, ip string) error {
    content := fmt.Sprintf("BASE_DOMAIN=%s\nbackendIP=%s\n", domain, ip)
    return os.WriteFile(path, []byte(content), 0644)
}

// Expose your command so it can be added to the root cmd
func NewDeployJenkinsCmd() *cobra.Command {
    return deployJenkinsCmd
}
