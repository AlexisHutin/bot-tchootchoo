package cmd

import (
	// "context"
	"fmt"
	"time"

	ffhbscraper "github.com/AlexisHutin/bot-tchootchoo/services/ffhb-scraper"
	// slackpkg "github.com/AlexisHutin/bot-tchootchoo/services/slack"
	"github.com/AlexisHutin/bot-tchootchoo/utils"
	"github.com/spf13/cobra"
)

var nextWeekend = &cobra.Command{
	Use:   "next-weekend",
	Short: "send a message with next weekend's match info",
	Long:  `Scrap the ffhb website and send a slack message with next week-end's match info`,
	Run: func(cmd *cobra.Command, args []string) {
		// ctx := context.Background()
		// config, err := loadConfig(configFile)
		// testChanID := "C082SC0DX36"

		
		matches, err := ffhbscraper.FetchMatches(ffhbscraper.Options{false, true, false})
		if err != nil {
			fmt.Println(err)
			return
		}

		nextWeekend, err := time.Parse("02/01", utils.GetNextWeekendDate())
		if err != nil {
			panic(err)
		}

		nextWeekend = nextWeekend.AddDate(time.Now().Year(), 0, 0)

		var nextWeekendMatches []ffhbscraper.Match
		for _, match := range matches {
			if match.Date.Raw == "" {
				continue
			}

			matchDate, err := time.Parse("2006-01-02 15:04:05.000", match.Date.Raw)
			if err != nil {
				panic(err)
			}

			if nextWeekend.Equal(matchDate.Truncate(24 * time.Hour)) {
				nextWeekendMatches = append(nextWeekendMatches, match)
			}
		}

		groupedMatches := ffhbscraper.GroupMatches(nextWeekendMatches)
		encodedMatches, err := ffhbscraper.EncodeMatches(groupedMatches)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Print(string(encodedMatches))
		fmt.Print(utils.GetNextWeekendDate())

		// SLACK
		// slackService, err := slackpkg.NewSlackCLient(ctx, config)
		// if err != nil {
		// 	fmt.Printf("Error : %s", err)
		// 	return
		// }
	},
}
