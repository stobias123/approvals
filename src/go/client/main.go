package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

type RequestApprovalPayload struct {
	// Add fields as necessary
}

type ApprovalResponse struct {
	ID string `json:"id"`
}

type ApprovalStatus struct {
	Approval struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	} `json:"approval"`
}

func main() {
	orgID := flag.String("org_id", "example-org", "Organization ID")
	waitTime := flag.String("wait_time", "30s", "Total wait time for approval status (e.g., 30s, 1m)")
	baseURL := flag.String("base_url", "http://localhost:8080", "Base URL of the approval service")

	flag.Parse()

	totalWaitDuration, err := time.ParseDuration(*waitTime)
	if err != nil {
		log.Fatalf("Invalid wait time duration: %v", err)
	}

	approvalID, err := createApproval(*baseURL, *orgID)
	if err != nil {
		log.Fatalf("Failed to create approval: %v", err)
	}

	fmt.Printf("Approval created with ID: %s\n", approvalID)

	err = waitForApproval(*baseURL, *orgID, approvalID, totalWaitDuration)
	if err != nil {
		if err.Error() == fmt.Sprintf("timeout: approval %s not approved within %s", approvalID, totalWaitDuration) {
			log.Error(err)
			os.Exit(2)
		} else if err.Error() == fmt.Sprintf("denied: approval %s has been denied", approvalID) {
			log.Error(err)
			os.Exit(1)
		} else {
			log.Fatalf("Failed to wait for approval: %v", err)
		}
	}

	fmt.Printf("Approval %s has been approved\n", approvalID)
	os.Exit(0)
}

func createApproval(baseURL, orgID string) (string, error) {
	url := fmt.Sprintf("%s/%s/approvals", baseURL, orgID)

	payload := RequestApprovalPayload{
		// Initialize payload fields as necessary
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to create approval: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, body)
	}

	var approvalResp ApprovalResponse
	err = json.NewDecoder(resp.Body).Decode(&approvalResp)
	if err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return approvalResp.ID, nil
}

func waitForApproval(baseURL, orgID, approvalID string, totalWaitDuration time.Duration) error {
	url := fmt.Sprintf("%s/%s/approvals/%s", baseURL, orgID, approvalID)
	interval := 5 * time.Second // Check every 5 seconds
	timeout := time.After(totalWaitDuration)
	ticker := time.NewTicker(interval)

	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout: approval %s not approved within %s", approvalID, totalWaitDuration)
		case <-ticker.C:
			log.Infof("Checking approval status for approval ID: %s", approvalID)
			resp, err := http.Get(url)
			if err != nil {
				return fmt.Errorf("failed to get approval status: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, body)
			}

			var approvalStatus ApprovalStatus
			err = json.NewDecoder(resp.Body).Decode(&approvalStatus)
			if err != nil {
				return fmt.Errorf("failed to decode response: %w", err)
			}

			if approvalStatus.Approval.Status == "approved" {
				return nil
			} else if approvalStatus.Approval.Status == "denied" {
				return fmt.Errorf("denied: approval %s has been denied", approvalID)
			}
		}
	}
}
