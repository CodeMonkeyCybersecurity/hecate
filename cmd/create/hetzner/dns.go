package hetzner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const hetznerAPIBase = "https://dns.hetzner.com/api/v1"

// CreateRecordRequest is the request body for creating or updating a DNS record.
type CreateRecordRequest struct {
	ZoneID string `json:"zone_id"`
	Type   string `json:"type"` // e.g. "A", "CNAME"
	Name   string `json:"name"`
	Value  string `json:"value"`
	TTL    int    `json:"ttl"`
}

// RecordResponse holds data for the record creation response.
type RecordResponse struct {
	Record struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"record"`
}

// ZonesResponse is used to decode the JSON containing a list of zones.
type ZonesResponse struct {
	Zones []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"zones"`
}

// NewCreateHetznerWildcardCmd returns a Cobra command for creating a Hetzner wildcard DNS record.
func NewCreateHetznerWildcardCmd() *cobra.Command {
	var (
		domain string
		ip     string
	)

	cmd := &cobra.Command{
		Use:   "hetzner-dns",
		Short: "Create a DNS record at Hetzner, ideally with a wildcard but with a fall back to a subdomain if wildcard fails",
		Run: func(cmd *cobra.Command, args []string) error {
			if domain == "" || ip == "" {
				return fmt.Errorf("domain and ip are required")
			}

			hetznerToken := os.Getenv("HETZNER_DNS_API_TOKEN")
			if hetznerToken == "" {
				return fmt.Errorf("missing Hetzner DNS API token (env HETZNER_DNS_API_TOKEN)")
			}

			// 1) Fetch the zone ID for the given domain from Hetzner
			zoneID, err := getZoneIDForDomain(hetznerToken, domain)
			if err != nil {
				return fmt.Errorf("failed to get zone for domain %q: %v", domain, err)
			}

			fmt.Printf("Using zone %s for domain %s\n", zoneID, domain)
			fmt.Println("Attempting to create wildcard record...")

			// 2) Attempt to create a wildcard record
			err = createRecord(hetznerToken, zoneID, "*", ip)
			if err != nil {
				fmt.Printf("Wildcard record creation failed: %v\n", err)
				fmt.Println("Falling back to normal subdomain: 'wildcard-fallback'")
				subdomain := "wildcard-fallback"

				fallbackErr := createRecord(hetznerToken, zoneID, subdomain, ip)
				if fallbackErr != nil {
					return fmt.Errorf("subdomain creation failed after wildcard failure: %v", fallbackErr)
				}

				fmt.Printf("Successfully created subdomain record: %s.%s -> %s\n", subdomain, domain, ip)
				return nil
			}

			// If we succeed with wildcard
			fmt.Printf("Successfully created wildcard record: *.%s -> %s\n", domain, ip)
			return nil
		},
	}

	cmd.Flags().StringVar(&domain, "domain", "", "Root domain name (e.g. example.com)")
	cmd.Flags().StringVar(&ip, "ip", "", "IP address for the A record")

	return cmd
}

// getZoneIDForDomain fetches all zones from Hetzner and attempts to match the given domain.
func getZoneIDForDomain(token, domain string) (string, error) {
	// Some users might store domain as "example.com" while the zone is "example.com."
	// For safety, we remove trailing dots, etc.
	domain = strings.TrimSuffix(domain, ".")

	req, err := http.NewRequest("GET", hetznerAPIBase+"/zones", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Auth-API-Token", token)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status from zones list: %s", resp.Status)
	}

	var zr ZonesResponse
	if err := json.NewDecoder(resp.Body).Decode(&zr); err != nil {
		return "", err
	}

	for _, z := range zr.Zones {
		// If the user’s domain is exactly the zone name or ends with it
		// (so “sub.example.com” can match zone “example.com”).
		zoneName := strings.TrimSuffix(z.Name, ".")
		if zoneName == domain || strings.HasSuffix(domain, zoneName) {
			return z.ID, nil
		}
	}
	return "", fmt.Errorf("zone not found for domain %q", domain)
}

// createRecord tries to create an A record in Hetzner DNS.
func createRecord(token, zoneID, name, ip string) error {
	reqBody := CreateRecordRequest{
		ZoneID: zoneID,
		Type:   "A",
		Name:   name, // "*" for wildcard, or "wildcard-fallback" for a normal subdomain
		Value:  ip,
		TTL:    300,  // Adjust as desired
	}

	bodyBytes, err := json.Marshal(&reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", hetznerAPIBase+"/records", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Auth-API-Token", token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var responseBody bytes.Buffer
		_, _ = responseBody.ReadFrom(resp.Body)
		return fmt.Errorf(
			"record creation failed (%d): %s",
			resp.StatusCode,
			responseBody.String(),
		)
	}

	var recordResp RecordResponse
	if err := json.NewDecoder(resp.Body).Decode(&recordResp); err != nil {
		return err
	}

	return nil
}
