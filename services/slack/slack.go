package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/slack-go/slack"
)

var (
	slackToken string = os.Getenv("SLACK_API_TOKEN")

	msgHeaderType string = "plain_text"
	msgHeaderText string = "Week end du %s"

	msgListSectionType string = "mrkdwn"
	msgListSectionText string = "Voilà la liste des *%d* joueurs dispo : %s"

	msgEndSectionType string = "plain_text"
	msgEndSectionText string = "Tchoo Tchoo !"
)

type Service struct {
	Slack *slack.Client
}

func NewSlackCLient(ctx context.Context) (*Service, error) {
	slackService := slack.New(slackToken)

	service := &Service{
		Slack: slackService,
	}

	return service, nil
}

func (s *Service) Ping(ctx context.Context, userID string) error {
	user, err := s.Slack.GetUserInfo(userID)
	if err != nil {
		fmt.Printf("%s\n", err)
		return err
	}

	fmt.Printf("ID: %s, Fullname: %s, Email: %s\n", user.ID, user.Profile.RealName, user.Profile.Email)
	return nil
}

type MessageData struct {
	UserID string
	Body   MessageBody
}

type MessageBody struct {
	MatchDate   string
	PlayersList []string
}

func (s *Service) SendMessage(ctx context.Context, data MessageData) error {
	messageBlocks, err := messageBuilder(data.Body)
	if err != nil {
		fmt.Printf("Error : %s", err)
		return err
	}

	slack.MsgOptionAttachments()
	_, _, err = s.Slack.PostMessage(data.UserID, slack.MsgOptionBlocks(messageBlocks...))
	if err != nil {
		fmt.Printf("Error : %s", err)
		return err
	}

	return nil
}

func messageBuilder(body MessageBody) ([]slack.Block, error) {
	// Header
	headerMsg := fmt.Sprintf(msgHeaderText, body.MatchDate)
	headerText := slack.NewTextBlockObject(msgHeaderType, headerMsg, false, false)
	headerSection := slack.NewHeaderBlock(headerText)

	// Body
	playersListStr := strings.Join(body.PlayersList, ", ")
	playersListLen := len(body.PlayersList)
	listMsg := fmt.Sprintf(msgListSectionText, playersListLen, playersListStr)
	listField := slack.NewTextBlockObject(msgListSectionType, listMsg, false, false)
	listSection := slack.NewSectionBlock(listField, nil, nil)

	// End
	endField := slack.NewTextBlockObject(msgEndSectionType, msgEndSectionText, false, false)
	endSection := slack.NewSectionBlock(endField, nil, nil)

	// Put message together
	messageBlocks := make([]slack.Block, 0)
	messageBlocks = append(messageBlocks, headerSection)
	messageBlocks = append(messageBlocks, listSection)
	messageBlocks = append(messageBlocks, endSection)

	// Log
	messageJson := slack.NewBlockMessage(
		headerSection,
		listSection,
		endSection,
	)
	
	message, err := json.MarshalIndent(messageJson, "", "	")
	if err != nil {
		fmt.Printf("Error : %s", err)
		return nil, err
	}

	fmt.Println(string(message))

	return messageBlocks, nil
}
