package ffhbscraper
// Credits to Mattéo le boss for this scraper !

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/gocolly/colly/v2"
)

// -------------------- CONFIGURATION --------------------

// debugMode enables verbose logs when true
var debugMode bool

// silentMode disables normal logs when true
var silentMode bool

// Options defines runtime options for scraping
type Options struct {
	ShowNext bool // keep only the next upcoming match for each team
	Debug    bool // enable debug logging
	Silent   bool // disable normal logging
}

// -------------------- LOGGING --------------------

// logInfo prints normal messages if silentMode is off
func logInfo(v ...interface{}) {
	if !silentMode {
		log.Println(v...)
	}
}

// logDebug prints debug messages if debugMode is on
func logDebug(v ...interface{}) {
	if debugMode && !silentMode {
		log.Println(append([]interface{}{"[DEBUG]"}, v...)...)
	}
}

// logError always prints error messages
func logError(v ...interface{}) {
	log.Println("[ERROR]", fmt.Sprint(v...))
}

// -------------------- DATA STRUCTURES --------------------

// Gym represents the match venue with address and optional navigation links
type Gym struct {
	Name     string `json:"name"`
	Street   string `json:"street"`
	Zip      string `json:"zip"`
	City     string `json:"city"`
	Lat      string `json:"latitude"`
	Lon      string `json:"longitude"`
	MapsLink string `json:"mapsLink,omitempty"`
	WazeLink string `json:"wazeLink,omitempty"`
}

// DateInfo contains parsed and human-readable match date details
type DateInfo struct {
	Raw    string `json:"raw"`
	Day    string `json:"day"`
	DayNum string `json:"dayNum"`
	Month  string `json:"month"`
	Year   string `json:"year"`
	Time   string `json:"time"`
}

// Teams holds the names of home and away teams
type Teams struct {
	Home string `json:"home"`
	Away string `json:"away"`
}

// Referees represents the two referees (if available)
type Referees struct {
	Ref1 string `json:"ref1"`
	Ref2 string `json:"ref2"`
}

// Match contains the full scraped data of a handball match
type Match struct {
	URL         string    `json:"url"`
	HasReferees bool      `json:"hasReferees"`
	Gym         *Gym      `json:"gym,omitempty"`
	Date        DateInfo  `json:"date"`
	ChampDay    string    `json:"championshipDay"`
	ParsedAt    string    `json:"parsedAt"`
	IsNext      bool      `json:"isNext,omitempty"`
	Status      *string   `json:"status,omitempty"`
	Teams       *Teams    `json:"teams,omitempty"`
	Referees    *Referees `json:"referees,omitempty"`
	Sex         string    `json:"-"`
	Level       string    `json:"level"`
}

// TeamConfig defines the configuration for a given team in a championship
type TeamConfig struct {
	Name            string
	ChampionShipURL string
	TeamURL         string
	PoolURL         string
	Sex             string
	Level           string
}

// -------------------- CONSTANTS --------------------

var BASE = "https://www.ffhandball.fr/competitions/saison-2025-2026-21/"

// French date helpers
var days = []string{"dimanche", "lundi", "mardi", "mercredi", "jeudi", "vendredi", "samedi"}
var months = []string{"janvier", "février", "mars", "avril", "mai", "juin",
	"juillet", "août", "septembre", "octobre", "novembre", "décembre"}

// Teams to scrape
var teamConfigs = []TeamConfig{
	{"ASC Rennais 1", "regional/16-ans-excellence-masculine-bretagne-27844/", "equipe-1945001/", "poule-168419/", "male", "regional_excellence"},
	{"ASC Rennais 1", "regional/16-ans-excellence-feminine-bretagne-27853/", "equipe-1945199/", "poule-168469/", "female", "regional_excellence"},
}

// -------------------- UTILITIES --------------------

// CapitalizeWords capitalizes the first letter of each word, keeping separators intact
func CapitalizeWords(text string) string {
	separators := " -’'"
	var result strings.Builder
	wordStart := true

	for _, r := range text {
		if strings.ContainsRune(separators, r) {
			result.WriteRune(r)
			wordStart = true
		} else {
			if wordStart {
				result.WriteRune(unicode.ToUpper(r))
				wordStart = false
			} else {
				result.WriteRune(unicode.ToLower(r))
			}
		}
	}
	return result.String()
}

// parseTime converts a raw datetime string into time.Time
// Uses Go's reference format: "2006-01-02 15:04:05.000"
func parseTime(raw string) (time.Time, error) {
	t, err := time.Parse("2006-01-02 15:04:05.000", raw)
	if err != nil {
		logDebug("Failed to parse date:", raw, err)
		return time.Time{}, err
	}
	return t, nil
}

// buildMatchURL constructs the full match URL from its components
func buildMatchURL(teamURL, pouleURL, matchID string) string {
	return fmt.Sprintf(BASE+teamURL+pouleURL+"rencontre-%s/", matchID)
}

// parseGym parses the "salle" JSON attribute into a Gym object
func parseGym(attr string) *Gym {
	if attr == "" {
		return nil
	}
	var tmp struct {
		Equipment struct {
			Name   string `json:"libelle"`
			Street string `json:"rue"`
			Zip    string `json:"codePostal"`
			City   string `json:"ville"`
			Lat    string `json:"latitude"`
			Lon    string `json:"longitude"`
		} `json:"equipement"`
	}
	if err := json.Unmarshal([]byte(attr), &tmp); err != nil {
		logDebug("parseGym JSON error:", err)
		return nil
	}
	g := &Gym{
		Name:   CapitalizeWords(tmp.Equipment.Name),
		Street: CapitalizeWords(tmp.Equipment.Street),
		Zip:    tmp.Equipment.Zip,
		City:   CapitalizeWords(tmp.Equipment.City),
		Lat:    tmp.Equipment.Lat,
		Lon:    tmp.Equipment.Lon,
	}
	if tmp.Equipment.Lat != "" && tmp.Equipment.Lon != "" {
		g.MapsLink = fmt.Sprintf("https://www.google.com/maps/search/?api=1&query=%s,%s", tmp.Equipment.Lat, tmp.Equipment.Lon)
		g.WazeLink = fmt.Sprintf("https://waze.com/ul?ll=%s,%s&navigate=yes", tmp.Equipment.Lat, tmp.Equipment.Lon)
	}
	return g
}

// parseReferees parses the referees JSON attribute into a Referees struct
func parseReferees(attr string) *Referees {
	r := &Referees{"", ""}
	if attr == "" || strings.Contains(attr, "[null,null]") {
		return r
	}
	var refs []string
	if err := json.Unmarshal([]byte(attr), &refs); err == nil {
		if len(refs) > 0 {
			r.Ref1 = refs[0]
		}
		if len(refs) > 1 {
			r.Ref2 = refs[1]
		}
	}
	return r
}

// parseTeamsAndDate parses the match teams and date JSON attribute
func parseTeamsAndDate(attr string) (Teams, DateInfo, string) {
	var teams Teams
	var date DateInfo
	var champDay string

	if attr == "" {
		return teams, date, champDay
	}

	var data struct {
		Date            string                `json:"date"`
		ChampionshipDay string                `json:"title"`
		Home            struct{ Name string } `json:"home"`
		Away            struct{ Name string } `json:"away"`
	}

	if err := json.Unmarshal([]byte(attr), &data); err == nil {
		teams = Teams{
			Home: CapitalizeWords(data.Home.Name),
			Away: CapitalizeWords(data.Away.Name),
		}
		date.Raw = data.Date
		if t, err := parseTime(data.Date); err == nil {
			date.Day = days[int(t.Weekday())]
			date.DayNum = fmt.Sprintf("%02d", t.Day())
			date.Month = months[int(t.Month())-1]
			date.Year = fmt.Sprintf("%d", t.Year())
			date.Time = fmt.Sprintf("%02dH%02d", t.Hour(), t.Minute())
		}
		champDay = data.ChampionshipDay
	}
	return teams, date, champDay
}

// -------------------- SCRAPER --------------------

// scrapeTeam fetches matches for a given team and sends them into the channel
func scrapeTeam(team TeamConfig, matchesChan chan<- Match, wg *sync.WaitGroup) {
	defer wg.Done()
	collector := colly.NewCollector(colly.Async(true))
	collector.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 4})

	collector.OnRequest(func(r *colly.Request) { logInfo("Visiting", r.URL.String()) })
	collector.OnError(func(r *colly.Response, err error) { logError("Request error:", r.Request.URL, err) })

	// Collect list of matches
	collector.OnHTML("smartfire-component[name='competitions---rencontre-list']", func(e *colly.HTMLElement) {
		var data struct {
			Meetings []struct {
				ExtMeetingID string `json:"ext_rencontreId"`
			} `json:"rencontres"`
		}
		if err := json.Unmarshal([]byte(e.Attr("attributes")), &data); err != nil {
			logDebug("scrapeTeam JSON parse error:", err)
			return
		}
		for _, r := range data.Meetings {
			matchURL := buildMatchURL(team.ChampionShipURL, team.PoolURL, r.ExtMeetingID)
			logDebug("Queueing match URL:", matchURL)
			collector.Visit(matchURL)
		}
	})

	// Parse match details
	collector.OnHTML("html", func(e *colly.HTMLElement) {
		now := time.Now().Truncate(time.Second)
		match := Match{
			URL:      e.Request.URL.String(),
			ParsedAt: now.Format(time.RFC3339),
			Sex:      team.Sex,
			Level:    team.Level,
		}
		match.Referees = parseReferees(e.ChildAttr("smartfire-component[name='competitions---rencontre-arbitres']", "attributes"))
		match.HasReferees = match.Referees.Ref1 != "" || match.Referees.Ref2 != ""
		match.Gym = parseGym(e.ChildAttr("smartfire-component[name='competitions---rencontre-salle']", "attributes"))
		t, d, c := parseTeamsAndDate(e.ChildAttr("smartfire-component[name='score']", "attributes"))
		match.Teams = &t
		match.Date = d
		match.ChampDay = c
		matchesChan <- match
	})

	collector.Visit(BASE + team.ChampionShipURL + team.TeamURL)
	collector.Wait()
}

// -------------------- PUBLIC API --------------------

// FetchMatches runs the scraping process and returns the list of matches
func FetchMatches(opts Options) ([]Match, error) {
	debugMode = opts.Debug
	silentMode = opts.Silent

	var wg sync.WaitGroup
	matchesChan := make(chan Match, 100)

	// Launch scraping for each configured team
	for _, team := range teamConfigs {
		wg.Add(1)
		go scrapeTeam(team, matchesChan, &wg)
	}

	// Close channel once scraping is done
	go func() {
		wg.Wait()
		close(matchesChan)
	}()

	// Collect results
	var matches []Match
	for m := range matchesChan {
		matches = append(matches, m)
	}

	// Optionally filter only next matches
	if opts.ShowNext {
		matches = filterNextMatches(matches)
	}
	return matches, nil
}

// GroupMatches organizes matches by sex and then by team
func GroupMatches(matches []Match) map[string]map[string][]Match {
	grouped := groupMatchesBySexAndTeam(matches)
	for sex, matchesByKey := range grouped {
		for key, matchesSlice := range matchesByKey {
			sortMatchesByDate(matchesSlice)
			grouped[sex][key] = matchesSlice
		}
	}
	return grouped
}

// EncodeMatches encodes grouped matches into pretty-printed JSON
func EncodeMatches(grouped map[string]map[string][]Match) ([]byte, error) {
	buf := &bytes.Buffer{}
	encoder := json.NewEncoder(buf)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(grouped); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// -------------------- FILTERING --------------------

// filterNextMatches keeps only the next upcoming match per level/sex
func filterNextMatches(matches []Match) []Match {
	nextMatches := make(map[string]Match)
	for _, m := range matches {
		if m.Date.Raw == "" {
			continue
		}
		matchTime, err := parseTime(m.Date.Raw)
		if err != nil {
			continue
		}
		key := m.Sex + "|" + m.Level
		existing, exists := nextMatches[key]
		if !exists {
			m.IsNext = true
			nextMatches[key] = m
		} else {
			existingTime, err := parseTime(existing.Date.Raw)
			if err != nil || matchTime.Before(existingTime) {
				m.IsNext = true
				existing.IsNext = false
				nextMatches[key] = m
			}
		}
	}
	result := make([]Match, 0, len(nextMatches))
	for _, m := range nextMatches {
		result = append(result, m)
	}
	return result
}

// sortMatchesByDate sorts matches in ascending order by parsed date
func sortMatchesByDate(matches []Match) {
	sort.Slice(matches, func(i, j int) bool {
		t1, err1 := parseTime(matches[i].Date.Raw)
		t2, err2 := parseTime(matches[j].Date.Raw)
		if err1 != nil || err2 != nil {
			return false
		}
		return t1.Before(t2)
	})
}

// groupMatchesBySexAndTeam groups matches first by sex, then by team name
func groupMatchesBySexAndTeam(matches []Match) map[string]map[string][]Match {
	result := make(map[string]map[string][]Match)
	for _, m := range matches {
		if _, ok := result[m.Sex]; !ok {
			result[m.Sex] = make(map[string][]Match)
		}
		teamName := ""
		for _, tc := range teamConfigs {
			if tc.Sex == m.Sex && tc.Level == m.Level {
				teamName = tc.Name
				break
			}
		}
		if teamName == "" {
			teamName = "unknown"
		}
		result[m.Sex][teamName] = append(result[m.Sex][teamName], m)
	}
	return result
}
