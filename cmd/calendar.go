// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"fmt"
	"log"

	pretty "github.com/jedib0t/go-pretty/v6/text"
	"github.com/rgolangh/calgo/internal/google_calendar"
	"github.com/spf13/cobra"
	"google.golang.org/api/calendar/v3"
)

// calendarCmd represents the calendar command
var calendarCmd = &cobra.Command{
	Use:   "calendar",
	Short: "List all user's calendars",
	Run: func(cmd *cobra.Command, args []string) {
		srv := google_calendar.Service()

		calendars, err := srv.CalendarList.List().Do()
		if err != nil {
			log.Fatalf("unable to retrieve calendars")
		}
		if len(calendars.Items) == 0 {
			fmt.Printf("No calendars")
			return
		}
		printCalendars(calendars)
		return
	},
}

func init() {
	rootCmd.AddCommand(calendarCmd)
}

func printCalendars(calendars *calendar.CalendarList) {
	fmt.Printf("%s%s%s\n", pretty.AlignLeft.Apply("id", 75), pretty.AlignLeft.Apply("summary", 75), pretty.AlignLeft.Apply("primary", 24))
	for _, c := range calendars.Items {
		if c.Primary {
			fmt.Printf("%s%s%v\n", pretty.AlignLeft.Apply(c.Id, 75), pretty.AlignLeft.Apply(c.Summary, 75), pretty.AlignLeft.Apply("true", 24))
			continue
		}
		fmt.Printf("%s%s\n", pretty.AlignLeft.Apply(c.Id, 75), pretty.AlignLeft.Apply(c.Summary, 75))
	}
}
