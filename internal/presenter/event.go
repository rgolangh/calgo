package presenter

import (
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/api/calendar/v3"
)

func FormatEvents(events *calendar.Events, format string) string {
	switch format {
	case prettyFormat:
		return formatEventsPretty(events)
	case jsonFormat:
		return formatEventsJson(events)
	default:
		return "unknown format"
	}
}

func formatEventsJson(events *calendar.Events) string {
	type event struct {
		Date    string `json:"date"`
		Summary string `json:"summary"`
	}
	evs := make([]event, 0, len(events.Items))
	for _, e := range events.Items {
		evs = append(evs, event{
			Date:    e.Start.DateTime,
			Summary: e.Summary,
		})
	}
	data, _ := json.Marshal(evs)
	return string(data)
}

func formatEventsPretty(events *calendar.Events) string {
	var sb strings.Builder
	for _, event := range events.Items {
		fmt.Fprintf(&sb, "%s - %s\n", event.Start.DateTime, event.Summary)
	}
	return sb.String()
}
