// pkg/certs/certs.go

package certs

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"go.uber.org/zap"

	"hecate/pkg/logger"
)

var log = logger.GetLogger()

// EnsureCertificates checks if certificate files exist for the given domain,
// and if not, calls an external command (like certbot) to obtain them.
// It now accepts appName, baseDomain, and email as parameters.

func EnsureCertificates(appName, baseDomain, email string) error {
	certDir := "certs"
	// Ensure the local certs directory exists.
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return fmt.Errorf("failed to create local certs directory: %w", err)
	}

	// Construct the fully qualified domain name for certbot.
	fqdn := fmt.Sprintf("%s.%s", appName, baseDomain)
	// Local destination file names use only the subdomain.
	destPrivKey := filepath.Join(certDir, fmt.Sprintf("%s.privkey.pem", appName))
	destFullChain := filepath.Join(certDir, fmt.Sprintf("%s.fullchain.pem", appName))

	// Check if the local certificate files exist.
	if _, err := os.Stat(destPrivKey); os.IsNotExist(err) || os.IsNotExist(func() error {
		_, err := os.Stat(destFullChain)
		return err
	}()) {
		log.Info("No certificate found locally; attempting to retrieve via certbot", zap.String("domain", fqdn))
		
		// Execute certbot to obtain a certificate.
		cmd := exec.Command("sudo", "certbot", "certonly", "--standalone", "--preferred-challenges", "http", "-d", fqdn, "-m", email, "--agree-tos", "--non-interactive")
		output, err := cmd.CombinedOutput()
		log.Info("Certbot output", zap.String("output", string(output)))
		if err != nil {
			log.Error("Failed to generate certificate", zap.String("domain", fqdn), zap.Error(err))
			return fmt.Errorf("failed to generate certificate: %w", err)
		}

		// After certbot runs successfully, copy the generated certificate files.
		sourceDir := filepath.Join("/etc/letsencrypt/live", fqdn)
		
		privKeyData, err := ioutil.ReadFile(filepath.Join(sourceDir, "privkey.pem"))
		if err != nil {
			return fmt.Errorf("failed to read privkey from certbot directory: %w", err)
		}
		if err := os.WriteFile(destPrivKey, privKeyData, 0644); err != nil {
			return fmt.Errorf("failed to write privkey to local certs directory: %w", err)
		}

		fullChainData, err := ioutil.ReadFile(filepath.Join(sourceDir, "fullchain.pem"))
		if err != nil {
			return fmt.Errorf("failed to read fullchain from certbot directory: %w", err)
		}
		if err := os.WriteFile(destFullChain, fullChainData, 0644); err != nil {
			return fmt.Errorf("failed to write fullchain to local certs directory: %w", err)
		}

		log.Info("Certificate retrieved and copied successfully", zap.String("domain", fqdn))
	} else {
		log.Info("Certificate exists locally", zap.String("domain", fqdn))
	}
	
	return nil
}
