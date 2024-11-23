package cmd

import (
	"context"
	"fmt"
	"sort"

	"github.com/AlexisHutin/bot-tchootchoo/services/sheet"
	"github.com/AlexisHutin/bot-tchootchoo/utils"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(getPlayers)
}

var getPlayers = &cobra.Command{
	Use:   "get-players",
	Short: "send a message to the coach with available players",
	Long: `Read the Google sheet and send a slack message to the coach with the list 
			of available players this week-end`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		// GOOGLE SHEET
		sheetService, err := sheet.NewSheetClient(ctx)
		if err != nil {
			fmt.Printf("Error : %s", err)
			return
		}

		availablePlayers, err := sheetService.GetAvailablePlayers()
		if err != nil {
			fmt.Printf("Error : %s", err)
		}
		playersNumber := len(availablePlayers)
		sort.Strings(availablePlayers)
		nextWeekend := utils.GetNextWeekendDate()

		fmt.Println(nextWeekend, playersNumber, availablePlayers)

		// SLACK
	},
}
