package cmd

import (
	"fmt"

	ffhbscraper "github.com/AlexisHutin/bot-tchootchoo/services/ffhb-scraper"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(ping)
}

var ping = &cobra.Command{
	Use:   "ping",
	Short: "reply pong",
	Long:  `This subcommand reply pong`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Pong !")
		matches, err := ffhbscraper.FetchMatches(ffhbscraper.Options{true,false,false})
		if err != nil {
			fmt.Println(err)
			return
		}

		groupedMatches := ffhbscraper.GroupMatches(matches)
		encodedMatches, err := ffhbscraper.EncodeMatches(groupedMatches)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Print(string(encodedMatches))
	},
}
