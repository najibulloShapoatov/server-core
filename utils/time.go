package utils

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// TimeAgo takes a time input and returns a string representation of how much time elapsed since the given time
func TimeAgo(datetime time.Time) string {
	if datetime.IsZero() {
		return ""
	}

	timeFormats := [][]interface{}{
		{60, "{0} seconds", 1},
		{120, "1 minute ago", "1 minute from now"},
		{3600, "{0} minutes", 60},
		{7200, "1 hour ago", "1 hour from now"},
		{86400, "{0} hours", 3600},
		{172800, "Yesterday", "Tomorrow"},
		{604800, "{0} days", 86400},
		{1209600, "Last week", "Next week"},
		{2419200, "{0} weeks", 604800},
		{4838400, "Last month", "Next month"},
		{29030400, "{0} months", 2419200},
		{58060800, "Last year", "Next year"},
		{2903040000, "{0} years", 29030400},
		{5806080000, "Last century", "Next century"},
		{58060800000, "{0} centuries", 2903040000},
	}

	var seconds = int(math.Floor(time.Since(datetime).Seconds()))
	var token = "{0} ago" // {0} ago
	var choice = 1

	if seconds == 0 {
		return "Just now" // Just now
	}
	if seconds < 0 {
		seconds = -seconds
		token = "{0} from now" // {0} from now
		choice = 2
	}
	for _, format := range timeFormats {
		if seconds < format[0].(int) {
			if _, ok := format[2].(string); ok {
				return format[choice].(string)
			}
			return strings.Replace(token, "{0}",
				strings.Replace(
					format[1].(string), "{0}",
					fmt.Sprintf("%d", int(math.Floor(float64(seconds)/float64(format[2].(int))))), -1),
				-1,
			)
		}
	}
	return ""
}

// FirstDayOfISOWeek returns the date of the Monday in the given week/year
func FirstDayOfISOWeek(year int, week int, timezone *time.Location) time.Time {
	date := time.Date(year, 0, 0, 0, 0, 0, 0, timezone)
	isoYear, isoWeek := date.ISOWeek()

	// iterate back to Monday
	for date.Weekday() != time.Monday {
		date = date.AddDate(0, 0, -1)
		isoYear, isoWeek = date.ISOWeek()
	}

	// iterate forward to the first day of the first week
	for isoYear < year {
		date = date.AddDate(0, 0, 7)
		isoYear, isoWeek = date.ISOWeek()
	}

	// iterate forward to the first day of the given week
	for isoWeek < week {
		date = date.AddDate(0, 0, 7)
		_, isoWeek = date.ISOWeek()
	}

	return date
}

// WorkingDays returns the number of working days between the two dates. It only calculates 5 working days a week
// and takes nothing else into consideration
func WorkingDays(startDate time.Time, endDate time.Time) int {
	if startDate.Equal(endDate) {
		return 1
	}

	if startDate.IsZero() || endDate.IsZero() {
		return 0
	}

	days := 0
	for {
		if startDate.Equal(endDate) || startDate.After(endDate) {
			return days
		}
		if startDate.Weekday() != time.Saturday && startDate.Weekday() != time.Sunday {
			days++
		}
		startDate = startDate.Add(time.Hour * 24)
	}
}
