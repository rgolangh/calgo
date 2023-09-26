// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rgolangh/calgo/internal/google_calendar"
	"google.golang.org/api/calendar/v3"

	"github.com/spf13/cobra"
)

const dateExpressionRegex = "^(s|m|t|w|th|f|sa)|(-?\\d+(-\\d+)?)$"

var (
	calendarID string
)

const maxEvents = 10
const sortField = "startTime"
const showDeleted = false

// listCmd is for listing events on a calendar
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List events",
	Long: fmt.Sprintf(`List events from a calendar with a default number of %d
and sorted by their %s
	`, maxEvents, sortField),
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return nil
		}
		// first arg is either a day expression, or a range if it has a hyphen
		compile, err := regexp.Compile(dateExpressionRegex)
		if err != nil {
			return err
		}
		if !compile.MatchString(args[0]) {
			return fmt.Errorf("first argument is not a day or range expression")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		tmin, tmax, err := getTimeBoundaries(args)
		if err != nil {
			return err
		}
		srv := google_calendar.Service()

		events, err := srv.Events.List(calendarID).
			ShowDeleted(showDeleted).
			SingleEvents(true).
			TimeMin(tmin.Format(time.RFC3339)).
			TimeMax(tmax.Format(time.RFC3339)).
			MaxResults(maxEvents).
			OrderBy(sortField).
			Do()
		if err != nil {
			log.Fatalf("Unable to retrieve next ten of the user's events: %v", err)
		}

		fmt.Printf("Upcoming events(%d):\n", len(events.Items))
		if len(events.Items) == 0 {
			fmt.Println("No upcoming events found.")
		} else {
			for _, item := range events.Items {
				fmt.Printf(eventString(item))
			}
		}
		return nil
	},
}

func getTimeBoundaries(args []string) (time.Time, time.Time, error) {
	var tmin, tmax time.Time
	var err error = nil
	if len(args) == 0 {
		tmin = time.Now()
		tmax = time.Now()
	} else {
		tmin, tmax, err = parseDatetimeExpression(time.Now(), args[0])
		if err != nil {
			return tmin, tmax, err
		}
	}
	tmin = startOfDay(tmin)
	tmax = endOfDay(tmax)
	return tmin, tmax, nil
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 8, 0, 0, 0, t.Location())
}

func endOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 20, 0, 0, 0, t.Location())
}

// parseDatetimeExpression get a day, or range expression in the form of
// n or n-n where n is the day and return 2 datetime representing
// the min start time and max start time of meeting to search for
func parseDatetimeExpression(pointInTime time.Time, s string) (time.Time, time.Time, error) {
	tmin := pointInTime
	tmax := pointInTime

	compile, err := regexp.Compile(dateExpressionRegex)
	if err != nil {
		return tmin, tmax, err
	}
	if !compile.MatchString(s) {
		return tmin, tmax, fmt.Errorf("basic validation - first argument is not a day or range expression")
	}

	compile, err = regexp.Compile("^((\\d+)|(s|m|t|w|th|f|sa))(-)((\\d+)|(s|m|t|w|th|f|sa))?$")
	if err != nil {
		return tmin, tmax, err
	}
	submatch := compile.FindSubmatch([]byte(s))
	isRange := len(submatch) > 1
	if isRange {
		split := strings.Split(s, "-")
		tmin, err = timeFromExpression(pointInTime, split[0])
		if err != nil {
			return tmin, tmax, err
		}
		tmax, err = timeFromExpression(pointInTime, split[1])
		if err != nil {
			return tmin, tmax, err
		}
	} else {
		tmax, err = timeFromExpression(pointInTime, s)
		tmin = tmax
		if err != nil {
			return tmin, tmax, err
		}
	}
	return tmin, tmax, nil
}

// is this date expression worth a specialized module of its own?
func timeFromExpression(pointInTime time.Time, s string) (time.Time, error) {
	if strings.HasPrefix(s, "+") || strings.HasPrefix(s, "-") {
		atoi, err := strconv.Atoi(s)
		if err != nil {
			return pointInTime, err
		}
		pointInTime = pointInTime.AddDate(0, 0, atoi)
		return pointInTime, nil
	}

	var targetWeekday time.Weekday
	switch s {
	case "1", "s":
		targetWeekday = time.Sunday
	case "2", "m":
		targetWeekday = time.Monday
	case "3", "t":
		targetWeekday = time.Tuesday
	case "4", "w":
		targetWeekday = time.Wednesday
	case "5", "th":
		targetWeekday = time.Thursday
	case "6", "f":
		targetWeekday = time.Friday
	case "7", "sa":
		targetWeekday = time.Saturday
	default:
		return pointInTime, fmt.Errorf("unsupported expression %s", s)
	}
	// The calculated delta is always to the next coming target weekday
	// always in a range of one week from now. For navigating or paging through the
	// calendar there should be other expression forms or subcommands.
	var delta int
	// calculate delta to the targetWeekDay. Weekdays values are 0-6
	if targetWeekday >= pointInTime.Weekday() {
		// e.g Thursday > Monday
		delta = int(targetWeekday - pointInTime.Weekday())
	} else {
		// e.g Sunday < Thursday
		delta = 7 - int(pointInTime.Weekday()-targetWeekday)
	}
	return pointInTime.AddDate(0, 0, delta), nil
}

func eventString(e *calendar.Event) string {
	parse, err := time.Parse(time.RFC3339, e.Start.DateTime)
	if err != nil {
		fmt.Println(err)
	}
	end, err := time.Parse(time.RFC3339, e.End.DateTime)
	if err != nil {
		fmt.Println(err)
	}
	var n = ""
	if len(e.Id) == 0 {
		n = "[+]"
	}
	return fmt.Sprintf("%-17s - %-3s %v\n", fmt.Sprintf("%s - %s", parse.Format(time.Kitchen), end.Format(time.Kitchen)), n, e.Summary)
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVar(&calendarID, "calendar-id", "primary", "id of the calendar")
}
