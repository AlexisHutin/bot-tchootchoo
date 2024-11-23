package sheet

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/AlexisHutin/bot-tchootchoo/utils"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

var (
	sheetsAPIKey  string = os.Getenv("SHEET_API_KEY")
	spreadsheetID string = os.Getenv("SPREADSHEET_ID")
)

type Service struct {
	Sheet *sheets.Service
}

func NewSheetClient(ctx context.Context) (*Service, error) {
	sheetService, err := sheets.NewService(ctx, option.WithAPIKey(sheetsAPIKey))
	if err != nil {
		fmt.Printf("Error : %s", err)
		return nil, err
	}
	
	service := &Service{
		Sheet: sheetService,
	}

	return service, nil
}

// Return list of available players this weekend
func (s *Service) GetAvailablePlayers() ([]string, error) {

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

// Return next week end row nomber
func (s *Service) getNextWeekendRow() *int {
	dateList, err := s.getDateList()
	if err != nil {
		fmt.Printf("Error : %s", err)
		return nil
	}

	nextWeekend := utils.GetNextWeekendDate()

	for key, date := range dateList {
		if date == nextWeekend {
			return &key
		}
	}

	return nil
}

// Return all week ends list
func (s *Service) getDateList() (map[int]string, error) {
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
func (s *Service) getPlayersList() (map[int]string, error) {
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
