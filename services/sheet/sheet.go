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

// Fixed columns
const (
	teamOneCol = "B"
	teamTwoCol = "C"
)

var (
	Home types.RGB
	Away types.RGB
	Cup  types.RGB
)

type Service struct {
	Sheet  *sheets.Service
	Config types.SheetEntry
}

// Create a new Google Sheets client with provided configuration
func NewSheetClient(ctx context.Context, globalConfig *types.Config, config types.SheetEntry) (*Service, error) {
	sheetService, err := sheets.NewService(ctx, option.WithAPIKey(globalConfig.Sheet.APIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create sheet service: %w", err)
	}

	Home = globalConfig.Sheet.CellTypeColor.Home
	Away = globalConfig.Sheet.CellTypeColor.Away
	Cup = globalConfig.Sheet.CellTypeColor.Cup

	return &Service{Sheet: sheetService, Config: config}, nil
}

// Fetch the complete list of players from the sheet
func (s *Service) getPlayersList() ([]string, error) {
	resp, err := s.Sheet.Spreadsheets.Values.Get(s.Config.ID, "3:3").Do()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch players list: %w", err)
	}
	if len(resp.Values) == 0 {
		return nil, nil
	}

	var players []string
	for _, row := range resp.Values {
		for idx, cell := range row {
			val, ok := cell.(string)
			if !ok || val == "" {
				continue
			}
			// Skip first two columns
			if idx >= 2 {
				players = append(players, val)
			}
		}
	}
	return players, nil
}

// Fetch the list of players available for the next weekend
func (s *Service) GetAvailablePlayers() ([]string, error) {
	players, err := s.getPlayersList()
	if err != nil {
		return nil, err
	}
	if len(players) == 0 {
		return nil, nil
	}

	row := s.getNextWeekendRow()
	if row == nil {
		return nil, fmt.Errorf("no weekend row found")
	}

	sheetRange := fmt.Sprintf("%d:%d", *row, *row)
	resp, err := s.Sheet.Spreadsheets.Values.Get(s.Config.ID, sheetRange).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch availability: %w", err)
	}
	if len(resp.Values) == 0 {
		return nil, nil
	}

	var available []string
	for idx, player := range players {
		if idx < len(resp.Values[0]) {
			value, ok := resp.Values[0][idx].(string)
			if ok && (strings.EqualFold(value, "x")) {
				available = append(available, player)
			}
		}
	}
	return available, nil
}

// Fetch the list of all dates from the sheet
func (s *Service) getDateList() ([]string, error) {
	resp, err := s.Sheet.Spreadsheets.Values.Get(s.Config.ID, "A:A").Do()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch dates: %w", err)
	}
	if len(resp.Values) == 0 {
		return nil, nil
	}

	var dates []string
	for i, row := range resp.Values {
		if i == 0 {
			continue // skip header
		}
		for _, cell := range row {
			if val, ok := cell.(string); ok && val != "" {
				dates = append(dates, val)
			}
		}
	}
	return dates, nil
}

// Find the row number corresponding to the next weekend date
func (s *Service) getNextWeekendRow() *int {
	dates, err := s.getDateList()
	if err != nil {
		log.Printf("failed to fetch dates: %v", err)
		return nil
	}

	next := utils.GetNextWeekendDate()
	for i, date := range dates {
		if date == next {
			row := i + 2 // +2 because of header and 0-based index
			return &row
		}
	}
	return nil
}

// Fetch match information for the next weekend (home/away/cup)
func (s *Service) GetMatchInfo() (map[string]string, error) {
	row := s.getNextWeekendRow()
	if row == nil {
		return nil, fmt.Errorf("no next weekend row found")
	}

	sheetRange := fmt.Sprintf("%s%d:%s%d", teamOneCol, *row, teamTwoCol, *row)
	resp, err := s.Sheet.Spreadsheets.Get(s.Config.ID).IncludeGridData(true).Ranges(sheetRange).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch match info: %w", err)
	}

	rawValues := resp.Sheets[0].Data[0].RowData[0].Values
	matchInfo := make(map[string]string, len(rawValues))

	for idx, raw := range rawValues {
		bg := types.RGB{
			Red:   raw.EffectiveFormat.BackgroundColor.Red,
			Green: raw.EffectiveFormat.BackgroundColor.Green,
			Blue:  raw.EffectiveFormat.BackgroundColor.Blue,
		}

		value := raw.FormattedValue
		var info string

		switch {
		case reflect.DeepEqual(bg, Home):
			info = fmt.Sprintf("%s at home", value)
		case reflect.DeepEqual(bg, Away):
			info = fmt.Sprintf("%s away", value)
		case reflect.DeepEqual(bg, Cup):
			info = fmt.Sprintf("%s French Cup", value)
		default:
			info = "No match"
		}

		matchInfo[fmt.Sprintf("team_%d", idx+1)] = info
	}
	return matchInfo, nil
}
