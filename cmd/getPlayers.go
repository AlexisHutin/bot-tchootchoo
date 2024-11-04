package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

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

		services := Services{
			Sheet: sheetService,
		}

		availablePlayers, err := services.getAvailablePlayers()
		if err != nil {
			fmt.Printf("Error : %s", err)
		}
		playersNumber := len(availablePlayers)
		sort.Strings(availablePlayers)
		nextWeekend := getNextWeekendDate()

		fmt.Println(nextWeekend, playersNumber, availablePlayers)
	},
}

type Services struct {
	Sheet *sheets.Service
}

// Return next week end row nomber
func (s *Services) getNextWeekendRow() *int {
	dateList, err := s.getDateList()
	if err != nil {
		fmt.Printf("Error : %s", err)
		return nil
	}

	nextWeekend := getNextWeekendDate()

	for key, date := range dateList {
		if date == nextWeekend {
			return &key
		}
	}

	return nil
}

// Return all week ends list
func (s *Services) getDateList() (map[int]string, error) {
	resp, err := s.Sheet.Spreadsheets.Values.Get(spreadsheetID, "A:A").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %+v", err)
	}

	var dateMap = make(map[int]string)

	if len(resp.Values) == 0 {
		fmt.Println("No data found.")
	} else {
		for key, row := range resp.Values {
			for _, cell := range row {
				cellString := cell.(string)
				dateMap[key+1] = cellString
			}
		}
	}

	delete(dateMap, 1)

	return dateMap, nil
}

// Return list of all players
func (s *Services) getPlayersList() (map[int]string, error) {
	resp, err := s.Sheet.Spreadsheets.Values.Get(spreadsheetID, "3:3").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %+v", err)
	}

	var players = make(map[int]string)
	if len(resp.Values) == 0 {
		fmt.Println("No data found.")
	} else {
		for _, row := range resp.Values {
			for key, cell := range row {
				if cell != nil && cell != "" {
					cellString := cell.(string)
					players[key] = cellString
				}
			}
		}
	}

	return players, nil
}

// Return list of available players this weekend
func (s *Services) getAvailablePlayers() ([]string, error) {

	playersList, err := s.getPlayersList()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %+v", err)
	}

	row := s.getNextWeekendRow()
	sheetRange := fmt.Sprintf("%v:%v", *row, *row)

	resp, err := s.Sheet.Spreadsheets.Values.Get(spreadsheetID, sheetRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %+v", err)
	}

	var availablePlayers []string
	if len(resp.Values) == 0 {
		fmt.Println("No data found.")
	} else {
		var value string
		for key, player := range playersList {
			value = strings.TrimSpace(resp.Values[0][key].(string))
			if value == "x" || value == "X" {
				availablePlayers = append(availablePlayers, player)
			}
		}
	}
	return availablePlayers, nil
}

// Return the next weejend date formated like this dd/mm
func getNextWeekendDate() string {
	today := time.Now()
	daysUntilSaturday := (6 - int(today.Weekday()) + 7) % 7
	nextSaturday := today.AddDate(0, 0, daysUntilSaturday)
	nextSaturdayString := nextSaturday.Format("02/01")
	return nextSaturdayString
}
