package main

import (
	"github.com/gin-gonic/gin"
	"github.com/stobias123/slack_button/http_handlers"
	"github.com/stobias123/slack_button/managers"
)

type ApprovalPayload struct {
	Approved   bool   `json:"approved"`
	ApprovalID string `json:"approval_id"`
}

type Server struct {
	Router *gin.Engine
}

func NewServer(sm *managers.SlackManager, am *managers.ApprovalManager) *Server {
	router := gin.Default()

	config := &http_handlers.ApprovalsControllerConfig{
		Url: "https://approvals.fly.dev",
	}

	approvalsController := http_handlers.NewApprovalsController(config, sm, am)
	approvalsController.RegisterHandlers(router)

	return &Server{
		Router: router,
	}
}
