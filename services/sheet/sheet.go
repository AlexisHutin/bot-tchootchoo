package sheet

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/AlexisHutin/bot-tchootchoo/types"
	"github.com/AlexisHutin/bot-tchootchoo/utils"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

var (
	teamOneCol string = "B"
	teamTwoCol string = "C"

	Home types.RGB
	Away types.RGB
	Cup  types.RGB
)

type Service struct {
	Sheet  *sheets.Service
	Config types.SheetEntry
}

func NewSheetClient(ctx context.Context, globalConfig *types.Config, config types.SheetEntry) (*Service, error) {
	sheetsAPIKey := globalConfig.Sheet.APIKey
	sheetService, err := sheets.NewService(ctx, option.WithAPIKey(sheetsAPIKey))
	if err != nil {
		fmt.Printf("Error : %s", err)
		return nil, err
	}

	service := &Service{
		Sheet:  sheetService,
		Config: config,
	}

	Home = globalConfig.Sheet.CellTypeColor.Home
	Away = globalConfig.Sheet.CellTypeColor.Away
	Cup = globalConfig.Sheet.CellTypeColor.Cup

	return service, nil
}

// === PLAYERS LIST === //
// Return list of available players this weekend
func (s *Service) GetAvailablePlayers() ([]string, error) {

	playersList, err := s.getPlayersList()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %+v", err)
	}

	row := s.getNextWeekendRow()
	sheetRange := fmt.Sprintf("%v:%v", *row, *row)

	resp, err := s.Sheet.Spreadsheets.Values.Get(s.Config.ID, sheetRange).Do()
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

// Return all week ends list
func (s *Service) getDateList() (map[int]string, error) {
	resp, err := s.Sheet.Spreadsheets.Values.Get(s.Config.ID, "A:A").Do()
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
	resp, err := s.Sheet.Spreadsheets.Values.Get(s.Config.ID, "3:3").Do()
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

// === MATCH INFO === //
func (s *Service) GetMatchInfo() (map[string]string, error) {
	nextWeekendRow := s.getNextWeekendRow()
	sheetRange := fmt.Sprintf("%v%v:%v%v", teamOneCol, *nextWeekendRow, teamTwoCol, *nextWeekendRow)
	resp, err := s.Sheet.Spreadsheets.Get(s.Config.ID).IncludeGridData(true).Ranges(sheetRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %+v", err)
	}

	// Get cells value & format (here background color)
	var cellsData []types.SheetCellData
	rawValues := resp.Sheets[0].Data[0].RowData[0].Values
	for _, rawValue := range rawValues {

		background := types.RGB{
			Blue:  rawValue.EffectiveFormat.BackgroundColor.Blue,
			Green: rawValue.EffectiveFormat.BackgroundColor.Green,
			Red:   rawValue.EffectiveFormat.BackgroundColor.Red,
		}

		cell := types.SheetCellData{
			Value:      rawValue.FormattedValue,
			Background: background,
		}

		cellsData = append(cellsData, cell)
	}

	matchInfo := make(map[string]string)
	for key, data := range cellsData {
		var infoStr string
		switch {
		case reflect.DeepEqual(data.Background, Home):
			infoStr = fmt.Sprintf("%v à domicile ", data.Value)
		case reflect.DeepEqual(data.Background, Away):
			infoStr = fmt.Sprintf("%v à l'extérieur ", data.Value)
		case reflect.DeepEqual(data.Background, Cup):
			infoStr = fmt.Sprintf("%v coupe de France ", data.Value)
		default:
			infoStr = "Pas de match"
		}

		fmt.Printf("Match found for team %v : %v\n", key+1, infoStr)
		matchInfo[fmt.Sprintf("team_%v", key+1)] = infoStr
	}

	return matchInfo, nil
}

// === OTHER === //
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
