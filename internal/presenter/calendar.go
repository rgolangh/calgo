package presenter

import (
	"encoding/json"
	"fmt"
	"strings"

	prettyText "github.com/jedib0t/go-pretty/v6/text"
	"google.golang.org/api/calendar/v3"
)

const (
	prettyFormat string = "pretty"
	jsonFormat   string = "json"
)

func FormatCalendars(calendars *calendar.CalendarList, format string) string {
	switch format {
	case prettyFormat:
		return formatCalendarsPretty(calendars)
	case jsonFormat:
		return formatCalendarsJson(calendars)
	default:
		return "unknown format"
	}
}

func formatCalendarsJson(calendars *calendar.CalendarList) string {
	type cal struct {
		ID      string `json:"id"`
		Summary string `json:"summary"`
		Primary bool   `json:"primary"`
	}
	cals := make([]cal, 0, len(calendars.Items))
	for _, c := range calendars.Items {
		cals = append(cals, cal{
			ID:      c.Id,
			Summary: c.Summary,
			Primary: c.Primary,
		})
	}
	data, _ := json.Marshal(cals)
	return string(data)
}

func formatCalendarsPretty(calendars *calendar.CalendarList) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s%s%s\n", prettyText.AlignLeft.Apply("id", 75), prettyText.AlignLeft.Apply("summary", 75), prettyText.AlignLeft.Apply("primary", 24))
	for _, c := range calendars.Items {
		if c.Primary {
			fmt.Fprintf(&sb, "%s%s%v\n", prettyText.AlignLeft.Apply(c.Id, 75), prettyText.AlignLeft.Apply(c.Summary, 75), prettyText.AlignLeft.Apply("true", 24))
			continue
		}
		fmt.Fprintf(&sb, "%s%s\n", prettyText.AlignLeft.Apply(c.Id, 75), prettyText.AlignLeft.Apply(c.Summary, 75))
	}
	return sb.String()
}
