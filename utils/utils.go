package utils

import "time"

// Return the next weejend date formated like this dd/mm
func GetNextWeekendDate() string {
	today := time.Now()
	daysUntilSaturday := (6 - int(today.Weekday()) + 7) % 7
	nextSaturday := today.AddDate(0, 0, daysUntilSaturday)
	nextSaturdayString := nextSaturday.Format("02/01")
	return nextSaturdayString
}
