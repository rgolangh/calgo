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
		compile, err := regexp.Compile("^\\d+(-\\d+)?$")
		if err != nil {
			return err
		}
		if !compile.MatchString(args[0]) {
			return fmt.Errorf("first argument is not a day or range expression")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		tmin, tmax, err := getMinMaxStartTimes(args[0])
		srv := google_calendar.Service()

		events, err := srv.Events.List(calendarID).ShowDeleted(false).
			SingleEvents(true).TimeMin(tmin.Format(time.RFC3339)).TimeMax(tmax.Format(time.RFC3339)).MaxResults(10).OrderBy("startTime").Do()
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

// getMinMaxTimes get a day, or range expression in the form of
// n or n-n where n is the day and return 2 datetime representing
// the min start time and max start time of meeting to search for
func getMinMaxStartTimes(s string) (time.Time, time.Time, error) {
	var tmin, tmax time.Time

	compile, err := regexp.Compile("^\\d+(-\\d+)?$")
	if err != nil {
		return tmin, tmax, err
	}
	if !compile.MatchString(s) {
		return tmin, tmax, fmt.Errorf("first argument is not a day or range expression")
	}

	isRange := strings.Contains(s, "-")
	if !isRange {
		tmax, err = timeFromExpression(s)
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
		return t.AddDate(0, 0, atoi), nil
	}

	// currently supporting extracting dates for this week only
	switch s {
	case "1", "s":
		t.AddDate(0, 0, int(time.Sunday-t.Weekday()))
	case "2", "m":
		t.AddDate(0, 0, int(time.Monday-t.Weekday()))
	case "3", "t":
		t.AddDate(0, 0, int(time.Tuesday-t.Weekday()))
	case "4", "w":
		t.AddDate(0, 0, int(time.Wednesday-t.Weekday()))
	case "5", "th":
		t.AddDate(0, 0, int(time.Thursday-t.Weekday()))
	case "6", "f":
		t.AddDate(0, 0, int(time.Friday-t.Weekday()))
	case "7", "sa":
		t.AddDate(0, 0, int(time.Saturday-t.Weekday()))
	default:
		return t, fmt.Errorf("non supported expression %s", s)
	}

	return t, nil
}

func printEvent(e *calendar.Event) {
	fmt.Printf("%v - %v\n", e.Start.DateTime, e.Summary)
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVar(&calendarID, "calendar_id", "primary", "id of the calendar")
}
