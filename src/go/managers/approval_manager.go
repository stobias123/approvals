package managers

import (
	"encoding/json"
	"log"

	"github.com/stobias123/slack_button/repositories"
)

type ApprovalManager struct {
	approvalRepository repositories.ApprovalRepository
}

func NewApprovalManager(ar repositories.ApprovalRepository) *ApprovalManager {
	return &ApprovalManager{
		approvalRepository: ar,
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
	return am.approvalRepository.UpdateApprovalStatus(orgID, approvalID, repositories.ApprovalStatusApproved)
}

func (am *ApprovalManager) Reject(orgID string, approvalID string) error {
	return am.approvalRepository.UpdateApprovalStatus(orgID, approvalID, repositories.ApprovalStatusRejected)
}

func (am *ApprovalManager) GetApprovals(orgID string) ([]repositories.Approval, error) {
	return am.approvalRepository.GetApprovals(orgID)
}

func (am *ApprovalManager) GetApproval(orgID string, approvalID string) (*repositories.Approval, error) {
	approvals, err := am.approvalRepository.GetApprovals(orgID)
	if err != nil {
		return nil, err
	}
	for _, approval := range approvals {
		if approval.ID == approvalID {
			return &approval, nil
		}
	}
	return nil, nil
}

func toJson(value interface{}) string {
	bytes, err := json.Marshal(value)
	if err != nil {
		log.Printf("Error marshalling to JSON: %v", err)
		return ""
	}
	return string(bytes)
}
