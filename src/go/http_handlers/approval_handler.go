package http_handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/loopfz/gadgeto/tonic"
	"github.com/stobias123/slack_button/managers"
)

type ApprovalsController struct {
	sm *managers.SlackManager
	am *managers.ApprovalManager
}

func NewApprovalsController(sm *managers.SlackManager, am *managers.ApprovalManager) *ApprovalsController {
	return &ApprovalsController{
		sm: sm,
		am: am,
	}
}

type RequestApprovalPayload struct {
}

func (ac *ApprovalsController) RegisterHandlers(r *gin.Engine) {
	r.POST("/:org_id/approvals", tonic.Handler(ac.HandleApprovalCreate, http.StatusOK))
	r.POST("/:org_id/approvals/:approval_id/approve", tonic.Handler(ac.HandleApprovalApprove, http.StatusOK))
	r.GET("/:org_id/approvals/:approval_id", tonic.Handler(ac.GetApproval, http.StatusOK))
	r.GET("/:org_id/approvals", tonic.Handler(ac.HandleApprovalList, http.StatusOK))
}

func (ac *ApprovalsController) GetApproval(c *gin.Context) (*gin.H, error) {
	orgID := c.Param("org_id")
	approvalID := c.Param("approval_id")

	// fetch the approval and return it
	approval, err := ac.am.GetApproval(orgID, approvalID)
	if err != nil {
		return nil, err
	}

	return &gin.H{"approval": approval}, nil
}

func (ac *ApprovalsController) HandleApprovalCreate(c *gin.Context, payload *RequestApprovalPayload) (*gin.H, error) {
	orgID := c.Param("org_id")
	approvalID := uuid.New().String()

	fmt.Printf("Received webhook for org_id: %s\n", orgID)
	fmt.Println("Payload:", toJson(payload))
	ac.sm.SendApprovalButton("#test", orgID, approvalID)

	ac.am.RequestApproval(orgID, approvalID)

	return &gin.H{"id": approvalID}, nil
}

func (act *ApprovalsController) HandleApprovalList(c *gin.Context) (*gin.H, error) {
	orgID := c.Param("org_id")
	approvals, err := act.am.GetApprovals(orgID)
	if err != nil {
		return nil, err
	}

	return &gin.H{"approvals": approvals}, nil
}

func (ac *ApprovalsController) HandleApprovalApprove(c *gin.Context) (*gin.H, error) {
	orgID := c.Param("org_id")
	approvalID := c.Param("approval_id")

	fmt.Printf("Received approval for org_id: %s, approval_id: %s\n", orgID, approvalID)

	ac.am.Approve(orgID, approvalID)

	return &gin.H{"message": "Approval received"}, nil
}

func toJson(value interface{}) string {
	bytes, err := json.Marshal(value)
	if err != nil {
		log.Printf("Error marshalling to JSON: %v", err)
		return ""
	}
	return string(bytes)
}
