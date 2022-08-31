/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
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

// initCmd represents the init command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List events",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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
			ShowDeleted(false).
			SingleEvents(true).
			TimeMin(tmin.Format(time.RFC3339)).
			TimeMax(tmax.Format(time.RFC3339)).
			MaxResults(10).
			OrderBy("startTime").
			Do()
		if err != nil {
			log.Fatalf("Unable to retrieve next ten of the user's events: %v", err)
		}

		fmt.Println("Upcoming events:")
		if len(events.Items) == 0 {
			fmt.Println("No upcoming events found.")
		} else {
			for _, item := range events.Items {
				date := item.Start.DateTime
				if date == "" {
					date = item.Start.Date
				}
				printEvent(item)
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
		tmin, tmax, err = parseDatetimeExpression(args[0])
		if err != nil {
			return tmin, tmax, err
		}
	}
	tmin = startOfDay(tmin)
	tmax = endOfDay(tmax)
	return tmin, tmax, nil
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func endOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// parseDatetimeExpression get a day, or range expression in the form of
// n or n-n where n is the day and return 2 datetime representing
// the min start time and max start time of meeting to search for
func parseDatetimeExpression(s string) (time.Time, time.Time, error) {
	tmin := time.Now()
	tmax := time.Now()

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
		tmin, err = timeFromExpression(split[0])
		if err != nil {
			return tmin, tmax, err
		}
		tmax, err = timeFromExpression(split[1])
		if err != nil {
			return tmin, tmax, err
		}
	} else {
		tmax, err = timeFromExpression(s)
		tmin = tmax
		if err != nil {
			return tmin, tmax, err
		}
	}
	return tmin, tmax, nil
}

func timeFromExpression(s string) (time.Time, error) {
	t := time.Now()
	if strings.HasPrefix(s, "+") || strings.HasPrefix(s, "-") {
		atoi, err := strconv.Atoi(s)
		if err != nil {
			return t, err
		}
		t = t.AddDate(0, 0, atoi)
		return t, nil
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
		return t, fmt.Errorf("unsupported expression %s", s)
	}
	// The calculated delta is always to the next coming target weekday
	// always in a range of one week from now. For navigating or paging through the
	// calendar there should be other expression forms or subcommands.
	var delta int
	// calculate delta to the targetWeekDay. Weekdays values are 0-6
	if targetWeekday >= t.Weekday() {
		// e.g Thursday > Monday
		delta = int(targetWeekday - t.Weekday())
	} else {
		// e.g Sunday < Thursday
		delta = 7 - int(t.Weekday()-targetWeekday)
	}
	return t.AddDate(0, 0, delta), nil
}

func printEvent(e *calendar.Event) {
	fmt.Printf("%v - %v\n", e.Start.DateTime, e.Summary)
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVar(&calendarID, "calendar_id", "primary", "id of the calendar")
}
