package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/AlexisHutin/bot-tchootchoo/services/sheet"
	"github.com/AlexisHutin/bot-tchootchoo/services/slack"
	"github.com/AlexisHutin/bot-tchootchoo/types"
	"github.com/AlexisHutin/bot-tchootchoo/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configFile string
var menFlag, womenFlag, testFlag bool

func init() {
	RootCmd.AddCommand(getPlayers)

	getPlayers.Flags().StringVarP(&configFile, "config", "c", "", "Path to the configuration file (required)")
	getPlayers.MarkFlagRequired("config")

	getPlayers.Flags().BoolVarP(&menFlag, "men", "m", false, "Send message to men's coaches")
	getPlayers.Flags().BoolVarP(&womenFlag, "women", "w", false, "Send message to women's coaches")
	getPlayers.Flags().BoolVarP(&testFlag, "test", "t", false, "Send message to test group coaches")
}

type FilteredConfig struct {
	Sheet       types.SheetEntry
	Coachs      []types.SlackUser
	MessageBody types.SlackMessageList
}

var getPlayers = &cobra.Command{
	Use:   "get-players",
	Short: "send a message to the coach with available players",
	Long: `Read the Google sheet and send a slack message to the coach with the list 
			of available players this week-end`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		config, err := loadConfig(configFile)
		if err != nil {
			fmt.Printf("Error : %s", err)
			return
		}

		filteredConfig, err := filterConfig(*config, cmd)
		if err != nil {
			fmt.Printf("Error : %s", err)
			return
		}

		// GOOGLE SHEET
		sheetService, err := sheet.NewSheetClient(ctx, config, filteredConfig.Sheet)
		if err != nil {
			fmt.Printf("Error : %s", err)
			return
		}

		matchInfo, err := sheetService.GetMatchInfo()
		if err != nil {
			fmt.Printf("Error : %s", err)
			return
		}

		if len(matchInfo) == 0 {
			err := errors.New("no match cells found")
			fmt.Printf("Error : %s", err)
			return
		}

		availablePlayers, err := sheetService.GetAvailablePlayers()
		if err != nil {
			fmt.Printf("Error : %s", err)
			return
		}

		sort.Strings(availablePlayers)
		nextWeekend := utils.GetNextWeekendDate()

		// SLACK
		slackService, err := slack.NewSlackCLient(ctx, config, filteredConfig.Coachs, filteredConfig.MessageBody)
		if err != nil {
			fmt.Printf("Error : %s", err)
			return
		}

		for key, user := range filteredConfig.Coachs {
			messageData := slack.MessageData{
				UserID: user.ID,
				Body: slack.MessageBody{
					MatchDate:   nextWeekend,
					MatchInfo:   matchInfo,
					PlayersList: availablePlayers,
				},
			}

			err := slackService.SendMessage(ctx, messageData)
			if err != nil {
				fmt.Printf("Error : %s", err)
				return
			}

			fmt.Printf("Message sent to %s !\n", user.Name)
			
			// No need to sleep at last message
			if key < len(filteredConfig.Coachs) - 1 {
				fmt.Println("Sleeping for 5 seconds...")
				time.Sleep(5 * time.Second)
			}
		}

		fmt.Println("Everything's ok !!")
	},
}

// Load configurations
func loadConfig(filePath string) (*types.Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config types.Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Filter config after the flag passed (men, wommen or test)
func filterConfig(config types.Config, cmd *cobra.Command) (*FilteredConfig, error) {
	flags := []struct {
		name  string
		set   bool
		group string
	}{
		{"men", cmd.Flags().Changed("men"), "men"},
		{"women", cmd.Flags().Changed("women"), "women"},
		{"test", cmd.Flags().Changed("test"), "test"},
	}

	activeFlags := 0
	var selectedGroup string

	for _, flag := range flags {
		if flag.set {
			activeFlags++
			selectedGroup = flag.group
		}
	}

	if activeFlags == 0 {
		return nil, fmt.Errorf("no flag specified; please use --men, --women, or --test\n")
	}
	if activeFlags > 1 {
		return nil, fmt.Errorf("only one flag can be used at a time; please use only one of --men, --women, or --test\n")
	}

	var filteredConfig FilteredConfig
	switch selectedGroup {
	case "men":
		filteredConfig = FilteredConfig{
			Sheet:       config.Sheet.Men,
			Coachs:      config.Slack.Users.Coachs.Men,
			MessageBody: config.Slack.Message.Men,
		}
	case "women":
		filteredConfig = FilteredConfig{
			Sheet:       config.Sheet.Women,
			Coachs:      config.Slack.Users.Coachs.Women,
			MessageBody: config.Slack.Message.Women,
		}
	case "test":
		filteredConfig = FilteredConfig{
			Sheet:       config.Sheet.Men,
			Coachs:      config.Slack.Users.Coachs.Test,
			MessageBody: config.Slack.Message.Men,
		}
	}

	return &filteredConfig, nil
}
