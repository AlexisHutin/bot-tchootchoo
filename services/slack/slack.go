package slack

import (
	"context"
	"fmt"

	"github.com/AlexisHutin/bot-tchootchoo/types"
	"github.com/slack-go/slack"
)

type Service struct {
	Slack  *slack.Client
}

func NewSlackCLient(ctx context.Context, globalConfig *types.Config) (*Service, error) {
	slackAPIKey := globalConfig.Slack.APIKey
	slackService := slack.New(slackAPIKey)

	service := &Service{
		Slack:  slackService,
	}

	return service, nil
}

func (s *Service) Ping(ctx context.Context, userID string) error {
	user, err := s.Slack.GetUserInfo(userID)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return err
	}

	fmt.Printf("ID: %s, Fullname: %s, Email: %s\n", user.ID, user.Profile.RealName, user.Profile.Email)
	return nil
}

func (s *Service) SendMessage(ctx context.Context, channelID string, messageBlocks []slack.Block) error {
	slack.MsgOptionAttachments()
	_, _, err := s.Slack.PostMessage(channelID, slack.MsgOptionBlocks(messageBlocks...))
	if err != nil {
		fmt.Printf("Error : %s", err)
		return err
	}

	return nil
}
