/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"log"

	"github.com/rgolangh/calgo/internal/google_calendar"
	"github.com/rgolangh/calgo/internal/presenter"
	"github.com/spf13/cobra"
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
		output := presenter.FormatCalendars(calendars, outputFormat)
		fmt.Printf("%s\n", output)
		return
	},
}

func init() {
	rootCmd.AddCommand(calendarCmd)
}
