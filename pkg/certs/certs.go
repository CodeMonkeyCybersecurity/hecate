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
	// Construct the fully qualified domain name.
	fqdn := fmt.Sprintf("%s.%s", appName, baseDomain)
	privKey := filepath.Join(certDir, fmt.Sprintf("%s.privkey.pem", fqdn))
	fullChain := filepath.Join(certDir, fmt.Sprintf("%s.fullchain.pem", fqdn))

	// Check if the private key exists.
	if _, err := os.Stat(privKey); os.IsNotExist(err) {
		log.Info("No certificate found; attempting to retrieve via certbot",
			zap.String("domain", fqdn))
		// Execute certbot to obtain a certificate.
		cmd := exec.Command("sudo", "certbot", "certonly", "--standalone", "--preferred-challenges", "http", "-d", fqdn, "-m", email, "--agree-tos", "--non-interactive")
		output, err := cmd.CombinedOutput()
		log.Info("Certbot output", zap.String("output", string(output)))
		if err != nil {
			log.Error("Failed to generate certificate", zap.String("domain", fqdn), zap.Error(err))
			return fmt.Errorf("failed to generate certificate: %w", err)
		}
		// In production, you would move or copy the generated certificates to certDir.
	} else if _, err := os.Stat(fullChain); os.IsNotExist(err) {
		log.Error("Certificate fullchain not found", zap.String("domain", fqdn))
		return fmt.Errorf("fullchain certificate missing")
	} else {
		log.Info("Certificate exists", zap.String("domain", fqdn))
	}
	return nil
}
