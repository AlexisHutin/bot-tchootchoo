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

var (
	sheetsAPIKey  string = os.Getenv("SHEET_API_KEY")
	spreadsheetID string = os.Getenv("SPREADSHEET_ID")
)

func init() {
	RootCmd.AddCommand(ping)
}

var ping = &cobra.Command{
	Use:   "ping",
	Short: "reply pong",
	Long:  `This subcommand reply pong`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		sheetService, err := sheets.NewService(ctx, option.WithAPIKey(sheetsAPIKey))
		if err != nil {
			fmt.Printf("Error : %s", err)
			return
		}

		resp, err := sheetService.Spreadsheets.Values.BatchGet(spreadsheetID).Ranges("3:3", "12:12").Do()

		if err != nil {
			log.Fatalf("Unable to retrieve data from sheet: %v", err)
		}

		if len(resp.ValueRanges) == 0 {
			fmt.Println("No data found.")
		} else {
			fmt.Printf("%v\n", len(resp.ValueRanges))
			for _, myrange := range resp.ValueRanges {
				for _, row := range myrange.Values {
					fmt.Printf("%s\n", row)
				}
			}
		}

		fmt.Println("Pong !")
	},
}
