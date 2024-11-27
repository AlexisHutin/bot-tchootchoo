package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/AlexisHutin/bot-tchootchoo/types"
	"github.com/slack-go/slack"
	"gopkg.in/yaml.v3"
)

var (
	slackToken string = os.Getenv("SLACK_API_TOKEN")
	configFile string = "config.yml"
)

type Service struct {
	Slack  *slack.Client
	Config Config
}

type Config struct {
	Users         []types.SlackUser
	MessageBody   types.SlackMessageList
	MessageCommon *types.SlackMessageCommon
}

func NewSlackCLient(ctx context.Context, users []types.SlackUser, messageBody types.SlackMessageList) (*Service, error) {
	slackService := slack.New(slackToken)
	messageCommon, err := loadSlackCommonMessageConfig(configFile)
	if err != nil {
		fmt.Printf("Error : %s", err)
		return nil, err
	}

	config := Config{
		Users:         users,
		MessageBody:   messageBody,
		MessageCommon: messageCommon,
	}
	service := &Service{
		Slack:  slackService,
		Config: config,
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
	messageBlocks, err := s.messageBuilder(data.Body)
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

func (s *Service) messageBuilder(body MessageBody) ([]slack.Block, error) {
	// Header
	headerMsg := fmt.Sprintf(s.Config.MessageCommon.Header.Text, body.MatchDate)
	headerText := slack.NewTextBlockObject(s.Config.MessageCommon.Header.Type, headerMsg, false, false)
	headerSection := slack.NewHeaderBlock(headerText)

	// Body
	playersListStr := strings.Join(body.PlayersList, ", ")
	playersListLen := len(body.PlayersList)
	listMsg := fmt.Sprintf(s.Config.MessageBody.List.Text, playersListLen, playersListStr)
	listField := slack.NewTextBlockObject(s.Config.MessageBody.List.Type, listMsg, false, false)
	listSection := slack.NewSectionBlock(listField, nil, nil)

	// End
	endField := slack.NewTextBlockObject(s.Config.MessageCommon.End.Type, s.Config.MessageCommon.End.Text, false, false)
	endSection := slack.NewSectionBlock(endField, nil, nil)

	// Context (assistance link)
	contextField := slack.NewTextBlockObject(s.Config.MessageCommon.Help.Type, s.Config.MessageCommon.Help.Text, false, false)
	contextSection := slack.NewSectionBlock(contextField, nil, nil)

	// Put message together
	messageBlocks := make([]slack.Block, 0)
	messageBlocks = append(messageBlocks, headerSection)
	messageBlocks = append(messageBlocks, listSection)
	messageBlocks = append(messageBlocks, endSection)
	messageBlocks = append(messageBlocks, contextSection)

	// Log
	messageJson := slack.NewBlockMessage(
		headerSection,
		listSection,
		endSection,
		contextSection,
	)

	message, err := json.MarshalIndent(messageJson, "", "	")
	if err != nil {
		fmt.Printf("Error : %s", err)
		return nil, err
	}

	fmt.Println(string(message))

	return messageBlocks, nil
}

func loadSlackCommonMessageConfig(filePath string) (*types.SlackMessageCommon, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening config file: %v", err)
	}
	defer file.Close()

	var config types.Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("error decoding config file: %v", err)
	}

	return &config.Slack.Message.Common, nil
}
