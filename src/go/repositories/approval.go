package repositories

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

type ApprovalStatus string

const (
	ApprovalStatusPending  ApprovalStatus = "pending"
	ApprovalStatusApproved ApprovalStatus = "approved"
	ApprovalStatusRejected ApprovalStatus = "rejected"
)

type Approval struct {
	ID     string         `json:"id"`
	OrgID  string         `json:"org_id"`
	Status ApprovalStatus `json:"status"`
}

type ApprovalRepository interface {
	CreateApproval(*Approval) error
	UpdateApprovalStatus(orgID string, approvalID string, status ApprovalStatus) error
	GetApprovals(orgID string) ([]Approval, error)
}

type ApprovalRepositoryInMemory struct {
	approvalMap map[string]map[string]Approval
}

func NewApprovalRepositoryInMemory() *ApprovalRepositoryInMemory {
	return &ApprovalRepositoryInMemory{
		approvalMap: make(map[string]map[string]Approval),
	}
}

func (ar *ApprovalRepositoryInMemory) CreateApproval(a *Approval) error {
	if _, ok := ar.approvalMap[a.OrgID]; !ok {
		ar.approvalMap[a.OrgID] = make(map[string]Approval)
	}
	ar.approvalMap[a.OrgID][a.ID] = *a
	return nil
}

func (ar *ApprovalRepositoryInMemory) UpdateApprovalStatus(orgID string, approvalID string, status ApprovalStatus) error {
	if _, ok := ar.approvalMap[orgID]; !ok {
		return fmt.Errorf("organization ID %s not found", orgID)
	}
	if approval, ok := ar.approvalMap[orgID][approvalID]; !ok {
		return fmt.Errorf("approval ID %s not found for organization ID %s", approvalID, orgID)
	} else {
		approval.Status = status
		log.Infof("Updated approval status for org_id: %s, approval_id: %s, status: %s", orgID, approvalID, status)
		ar.approvalMap[orgID][approvalID] = approval
	}
	return nil
}

func (ar *ApprovalRepositoryInMemory) GetApprovals(orgID string) ([]Approval, error) {
	if orgApprovals, ok := ar.approvalMap[orgID]; ok {
		approvals := make([]Approval, 0, len(orgApprovals))
		for _, approval := range orgApprovals {
			approvals = append(approvals, approval)
		}
		return approvals, nil
	}
	return nil, fmt.Errorf("no approvals found for organization ID %s", orgID)
}

func (ar *ApprovalRepositoryInMemory) GetApproval(orgID string, approvalID string) (*Approval, error) {
	if orgApprovals, ok := ar.approvalMap[orgID]; ok {
		if approval, ok := orgApprovals[approvalID]; ok {
			return &approval, nil
		}
	}
	return nil, fmt.Errorf("approval ID %s not found for organization ID %s", approvalID, orgID)
}
