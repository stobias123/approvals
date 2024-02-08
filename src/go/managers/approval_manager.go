package managers

import (
	"encoding/json"
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/stobias123/slack_button/repositories"
)

type ApprovalManager struct {
	approvalRepository repositories.ApprovalRepository
	mqttClient         mqtt.Client
}

func NewApprovalManager(ar repositories.ApprovalRepository) *ApprovalManager {
	opts := mqtt.NewClientOptions().AddBroker("tcp://mqtturl:1883").SetClientID("go_mqtt_client")
	opts.SetUsername("mqttuser")
	opts.SetPassword("mqttPass")
	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Error connecting to MQTT broker: %v", token.Error())
	}
	return &ApprovalManager{
		approvalRepository: ar,
		mqttClient:         mqttClient,
	}
}

func (am *ApprovalManager) RequestApproval(orgID string, approvalID string) error {
	approval := &repositories.Approval{
		ID:     approvalID,
		OrgID:  orgID,
		Status: repositories.ApprovalStatusPending,
	}
	return am.approvalRepository.CreateApproval(approval)
}

type ApprovalPayload struct {
	Approved   bool   `json:"approved"`
	ApprovalID string `json:"approval_id"`
}

func (am *ApprovalManager) Approve(orgID string, approvalID string) error {
	topicName := fmt.Sprintf("%s/approvals/response", orgID)
	approvalPayload := &ApprovalPayload{
		Approved:   true,
		ApprovalID: approvalID,
	}
	am.mqttClient.Publish(topicName, 0, false, toJson(approvalPayload))
	return am.approvalRepository.UpdateApprovalStatus(orgID, approvalID, repositories.ApprovalStatusApproved)
}

func (am *ApprovalManager) Reject(orgID string, approvalID string) error {
	topicName := fmt.Sprintf("%s/approvals/response", orgID)
	rejectPayload := &ApprovalPayload{
		Approved:   false,
		ApprovalID: approvalID,
	}
	am.mqttClient.Publish(topicName, 0, false, toJson(rejectPayload))
	return am.approvalRepository.UpdateApprovalStatus(orgID, approvalID, repositories.ApprovalStatusRejected)
}

func (am *ApprovalManager) GetApprovals(orgID string) ([]repositories.Approval, error) {
	return am.approvalRepository.GetApprovals(orgID)
}

func toJson(value interface{}) string {
	bytes, err := json.Marshal(value)
	if err != nil {
		log.Printf("Error marshalling to JSON: %v", err)
		return ""
	}
	return string(bytes)
}
