// pkg/certs/certs.go
package certs

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnsureCertificates checks if certificate files exist for the given domain,
// and if not, calls an external command (like certbot) to obtain them.
func EnsureCertificates(appName, baseDomain string) error {
	certDir := "certs"
	fqdn := fmt.Sprintf("%s.%s", appName, baseDomain)
	privKey := filepath.Join(certDir, fmt.Sprintf("%s.privkey.pem", fqdn))
	fullChain := filepath.Join(certDir, fmt.Sprintf("%s.fullchain.pem", fqdn))

	if _, err := os.Stat(privKey); os.IsNotExist(err) {
		fmt.Printf("No certificate found for %s. Attempting to retrieve via certbot...\n", fqdn)
		// Stub: Replace with your actual certbot call.
		// Example: exec.Command("certbot", "certonly", "--standalone", "--non-interactive", ...)
		// For now, we simply simulate success.
		fmt.Printf("✅ Simulating certificate generation for %s\n", fqdn)
		// In production, copy or symlink the generated certs into your certDir.
	} else if _, err := os.Stat(fullChain); os.IsNotExist(err) {
		fmt.Printf("Certificate fullchain not found for %s.\n", fqdn)
		return fmt.Errorf("fullchain certificate missing")
	} else {
		fmt.Printf("✅ Certificate for %s exists.\n", fqdn)
	}
	return nil
}
