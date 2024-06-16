package managers

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	logrus "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

type SlackConfig struct {
	AppToken string
	BotToken string
	Channel  string
}

type SlackManager struct {
	client            *socketmode.Client
	socketmodeHandler *socketmode.SocketmodeHandler
	approvalManager   *ApprovalManager
	config            SlackConfig
}

func NewSlackManager(am *ApprovalManager) *SlackManager {
	config := SlackConfig{
		AppToken: os.Getenv("SLACK_APP_TOKEN"),
		BotToken: os.Getenv("SLACK_BOT_TOKEN"),
		Channel:  os.Getenv("SLACK_CHANNEL"),
	}
	logrus.Infof("Slack config: %v", config)

	api := slack.New(
		config.BotToken,
		slack.OptionDebug(false),
		slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
		slack.OptionAppLevelToken(config.AppToken),
	)

	client := socketmode.New(
		api,
		socketmode.OptionDebug(false),
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	socketmodeHandler := socketmode.NewSocketmodeHandler(client)

	return &SlackManager{
		client:            client,
		socketmodeHandler: socketmodeHandler,
		approvalManager:   am,
		config:            config,
	}
}

func (sm *SlackManager) RegisterHandlers() {
	sm.socketmodeHandler.Handle(socketmode.EventTypeConnectionError, sm.middlewareConnectionError)
	sm.socketmodeHandler.HandleInteraction(slack.InteractionTypeBlockActions, sm.HandleApprovalBlockClick)
}

func (sm *SlackManager) middlewareConnectionError(evt *socketmode.Event, client *socketmode.Client) {
	fmt.Println("Connection failed. Retrying later...")
}

func (sm *SlackManager) Run(ctx context.Context) {
	sm.RegisterHandlers()
	sm.socketmodeHandler.RunEventLoopContext(ctx)
}

func (sm *SlackManager) SendApprovalButton(message string, orgID string, requestID string) {
	approvalBlock := sm.getApprovalBlock(message, orgID, requestID)
	_, _, _, error := sm.client.SendMessage(sm.config.Channel, slack.MsgOptionBlocks(approvalBlock...))
	if error != nil {
		log.Printf("failed posting message: %v", error)
	}
}

func (sm *SlackManager) HandleApprovalBlockClick(evt *socketmode.Event, client *socketmode.Client) {
	var payload interface{}
	callback, ok := evt.Data.(slack.InteractionCallback)
	if !ok {
		fmt.Printf("Ignored %+v\n", evt)
		return
	}
	client.Ack(*evt.Request, payload)
	for _, action := range callback.ActionCallback.BlockActions {
		if strings.Contains(action.Value, "approve") {
			orgID := strings.Split(action.Value, "_")[1]
			approvalID := strings.Split(action.Value, "_")[2]
			logrus.Infof("Approving deploy for org_id: %s, approval_id: %s", orgID, approvalID)
			sm.approvalManager.Approve(orgID, approvalID)
		}
	}
	_, _, _, err := client.Client.UpdateMessage(callback.Channel.ID, callback.Message.Msg.Timestamp, slack.MsgOptionText("Deploy approved", false))
	if err != nil {
		client.Debugf("failed posting message: %v", err)
	}
}

func (sm *SlackManager) getApprovalBlock(message string, orgID string, requestID string) []slack.Block {

	approvalValue := fmt.Sprintf("approve_%s_%s", orgID, requestID)
	denyValue := fmt.Sprintf("deny_%s_%s", orgID, requestID)

	// Header Section
	logrus.Infof("Creating approval block with message: %s", message)
	headerText := slack.NewTextBlockObject("mrkdwn", message, false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	// Approve and Deny Buttons
	approveBtnTxt := slack.NewTextBlockObject("plain_text", "Approve", false, false)
	approveBtn := slack.NewButtonBlockElement("", approvalValue, approveBtnTxt)

	denyBtnTxt := slack.NewTextBlockObject("plain_text", "Deny", false, false)
	denyBtn := slack.NewButtonBlockElement("", denyValue, denyBtnTxt)

	actionBlock := slack.NewActionBlock("", approveBtn, denyBtn)

	// Build Message with blocks created above

	msg := slack.NewBlockMessage(
		headerSection,
		actionBlock,
	)
	return msg.Blocks.BlockSet
}
