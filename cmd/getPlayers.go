package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func init() {
	RootCmd.AddCommand(getPlayers)
}

var (
	sheetsAPIKey  string = os.Getenv("SHEET_API_KEY")
	spreadsheetID string = os.Getenv("SPREADSHEET_ID")
)

var getPlayers = &cobra.Command{
	Use:   "get-players",
	Short: "send a message to the coach with available players",
	Long: `Read the Google sheet and send a slack message to the coach with the list 
			of available players this week-end`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		sheetService, err := sheets.NewService(ctx, option.WithAPIKey(sheetsAPIKey))
		if err != nil {
			fmt.Printf("Error : %s", err)
			return
		}

		resp, err := sheetService.Spreadsheets.Values.BatchGet(spreadsheetID).Ranges("3:3", "12:12").Do()
		if err != nil {
			log.Fatalf("Unable to retrieve data from sheet: %+v", err)
		}

		if len(resp.ValueRanges) == 0 {
			fmt.Println("No data found.")
		} else {
			fmt.Printf("%v\n", len(resp.ValueRanges))
			for _, myrange := range resp.ValueRanges {
				for _, row := range myrange.Values {
					for _, cell := range row {
						fmt.Printf("%s\n", cell)
					}

				}
			}
		}
	},
}
