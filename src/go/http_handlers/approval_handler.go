package http_handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/loopfz/gadgeto/tonic"
	"github.com/stobias123/slack_button/managers"
)

type ApprovalsController struct {
	mqttClient mqtt.Client
	sm         *managers.SlackManager
	am         *managers.ApprovalManager
}

func NewApprovalsController(mqttClient mqtt.Client, sm *managers.SlackManager, am *managers.ApprovalManager) *ApprovalsController {
	return &ApprovalsController{
		mqttClient: mqttClient,
		sm:         sm,
		am:         am,
	}
}

type RequestApprovalPayload struct {
}

func (ac *ApprovalsController) RegisterHandlers(r *gin.Engine) {
	r.POST("/:org_id/approvals", tonic.Handler(ac.HandleApprovalCreate, http.StatusOK))
	r.GET("/:org_id/approvals", tonic.Handler(ac.HandleApprovalList, http.StatusOK))
}

func (ac *ApprovalsController) HandleApprovalCreate(c *gin.Context, payload *RequestApprovalPayload) (*gin.H, error) {
	orgID := c.Param("org_id")
	approvalID := uuid.New().String()

	fmt.Printf("Received webhook for org_id: %s\n", orgID)
	fmt.Println("Payload:", toJson(payload))
	ac.sm.SendApprovalButton("#test", orgID, approvalID)

	ac.am.RequestApproval(orgID, approvalID)

	return &gin.H{"message": "Webhook received"}, nil
}

func (act *ApprovalsController) HandleApprovalList(c *gin.Context) (*gin.H, error) {
	orgID := c.Param("org_id")
	approvals, err := act.am.GetApprovals(orgID)
	if err != nil {
		return nil, err
	}

	return &gin.H{"approvals": approvals}, nil
}

func toJson(value interface{}) string {
	bytes, err := json.Marshal(value)
	if err != nil {
		log.Printf("Error marshalling to JSON: %v", err)
		return ""
	}
	return string(bytes)
}
