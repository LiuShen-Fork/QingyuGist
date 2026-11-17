package gist

import (
	"time"

	"github.com/thomiceli/opengist/internal/db"
)

// HeatmapCell is one day in a contribution heatmap.
type HeatmapCell struct {
	Date     string
	Label    string
	Count    int
	Level    int
	InRange  bool
	Weekday  int
	IsFuture bool
}

// HeatmapWeek is a column of seven days, Monday first.
type HeatmapWeek struct {
	Cells []HeatmapCell
}

// HeatmapData is everything the heatmap partial needs to render.
type HeatmapData struct {
	Weeks     []HeatmapWeek
	Months    []HeatmapMonth
	Total     int
	Max       int
	FromLabel string
	ToLabel   string
	Days      int
}

// HeatmapMonth marks the first week of a month so the template can render a
// month label above the grid without JavaScript.
type HeatmapMonth struct {
	Week  int
	Left  int
	Month int
}

const heatmapDays = 371 // 53 weeks, enough to always cover a full year view

// buildHeatmap aggregates activity for a user (userID != 0) or the whole
// instance (userID == 0) into Monday-first weeks ending today.
func buildHeatmap(userID uint) *HeatmapData {
	today := time.Now()
	start := today.AddDate(0, 0, -(heatmapDays - 1))
	// Rewind to the Monday of the first week so every column has seven cells.
	for start.Weekday() != time.Monday {
		start = start.AddDate(0, 0, -1)
	}
	end := today.AddDate(0, 0, 6-int(today.Weekday())) // Sunday of the current week

	rows, err := db.GetUserActivity(userID, start.Format("2006-01-02"), end.Format("2006-01-02"))
	if err != nil {
		rows = nil
	}
	counts := make(map[string]int, len(rows))
	for _, row := range rows {
		counts[row.Day] = row.Count
	}

	data := &HeatmapData{
		FromLabel: start.Format("2006-01-02"),
		ToLabel:   today.Format("2006-01-02"),
		Days:      heatmapDays,
	}

	for _, count := range counts {
		data.Total += count
		if count > data.Max {
			data.Max = count
		}
	}

	day := start
	var week HeatmapWeek
	for !day.After(end) {
		date := day.Format("2006-01-02")
		inRange := !day.Before(today.AddDate(0, 0, -(heatmapDays-1))) && !day.After(today)
		count := counts[date]
		cell := HeatmapCell{
			Date:     date,
			Label:    day.Format("Jan 2, 2006"),
			Count:    count,
			Level:    heatmapLevel(count),
			InRange:  inRange,
			Weekday:  int(day.Weekday()),
			IsFuture: day.After(today),
		}
		if !inRange {
			cell.Count = 0
			cell.Level = 0
		}
		week.Cells = append(week.Cells, cell)

		if day.Weekday() == time.Sunday {
			data.Weeks = append(data.Weeks, week)
			week = HeatmapWeek{}
		}
		day = day.AddDate(0, 0, 1)
	}
	if len(week.Cells) > 0 {
		data.Weeks = append(data.Weeks, week)
	}

	// Month labels: mark the week that contains the first Monday of a month.
	seen := ""
	for i, w := range data.Weeks {
		if len(w.Cells) == 0 {
			continue
		}
		month := w.Cells[0].Date[:7]
		if month != seen {
			seen = month
			if i > 0 {
				t, _ := time.Parse("2006-01-02", w.Cells[0].Date)
				data.Months = append(data.Months, HeatmapMonth{Week: i, Left: i * 14, Month: int(t.Month())})
			}
		}
	}

	return data
}

// heatmapLevel maps a daily count to one of five visual intensity levels.
func heatmapLevel(count int) int {
	switch {
	case count <= 0:
		return 0
	case count == 1:
		return 1
	case count <= 3:
		return 2
	case count <= 6:
		return 3
	default:
		return 4
	}
}
