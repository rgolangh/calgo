// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"github.com/rgolangh/calgo/internal/google_calendar"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the tool with your google credentials",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		google_calendar.Service()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
