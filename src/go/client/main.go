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
	Message string `json:"message"`
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

type OutputStatus struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func main() {
	orgID := flag.String("org-id", "example-org", "Organization ID")
	waitTime := flag.String("wait-time", "30s", "Total wait time for approval status (e.g., 30s, 1m)")
	baseURL := flag.String("base-url", "http://localhost:8080", "Base URL of the approval service")
	message := flag.String("message", "Please approve this request", "Message for the approval request")

	flag.Parse()

	exitCode := 0

	totalWaitDuration, err := time.ParseDuration(*waitTime)
	if err != nil {
		log.Fatalf("Invalid wait time duration: %v", err)
	}

	approvalID, err := createApproval(*baseURL, *orgID, *message)
	if err != nil {
		log.Fatalf("Failed to create approval: %v", err)
	}

	status, err := waitForApproval(*baseURL, *orgID, approvalID, totalWaitDuration)
	outputStatus := OutputStatus{
		ID:     approvalID,
		Status: status,
	}
	if err != nil {
		exitCode = 1
	}
	output, err := json.Marshal(outputStatus)
	if err != nil {
		log.Fatalf("Failed to marshal output status: %v", err)
	}
	fmt.Println(string(output))
	os.Exit(exitCode)
}

func createApproval(baseURL, orgID, message string) (string, error) {
	url := fmt.Sprintf("%s/%s/approvals", baseURL, orgID)

	payload := RequestApprovalPayload{
		Message: message,
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

func waitForApproval(baseURL, orgID, approvalID string, totalWaitDuration time.Duration) (string, error) {
	url := fmt.Sprintf("%s/%s/approvals/%s", baseURL, orgID, approvalID)
	interval := 5 * time.Second // Check every 5 seconds
	timeout := time.After(totalWaitDuration)
	ticker := time.NewTicker(interval)

	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return "timed out", fmt.Errorf("timeout: approval %s not approved within %s", approvalID, totalWaitDuration)
		case <-ticker.C:
			log.Infof("Checking approval status for approval ID: %s", approvalID)
			resp, err := http.Get(url)
			if err != nil {
				return "", fmt.Errorf("failed to get approval status: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return "", fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, body)
			}

			var approvalStatus ApprovalStatus
			err = json.NewDecoder(resp.Body).Decode(&approvalStatus)
			if err != nil {
				return "", fmt.Errorf("failed to decode response: %w", err)
			}

			log.Info(approvalStatus.Approval.Status)
			if approvalStatus.Approval.Status == "approved" {
				return "approved", nil
			} else if approvalStatus.Approval.Status == "rejected" {
				return "rejected", fmt.Errorf("denied: approval %s has been denied", approvalID)
			}
		}
	}
}
