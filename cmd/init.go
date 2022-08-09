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
    "os"

    "github.com/spf13/cobra"
    "golang.org/x/crypto/ssh/terminal"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
    Use:   "init",
    Short: "A brief description of your command",
    Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
    RunE: func(cmd *cobra.Command, args []string) error {
        fmt.Printf("" +
            "Lets setup the google calendar integration:\n" +
            "1. enable google calendar API for your account - https://console.cloud.google.com/apis/api/calendar-json.googleapis.com \n" +
            "2. create credentials:\n" +
            "   - credentials type: user-data,\n" +
            "   - scopes: 'https://www.googleapis.com/auth/calendar,\n" +
            "3. Go to 'credentials' https://console.cloud.google.com/apis/credentials/\n" +
            "   and fill in the details here:\n")

        fmt.Printf("\nCLIENT_ID:")
        clientId, err := terminal.ReadPassword(int(os.Stdin.Fd()))
        if err != nil {
            return err
        }
        fmt.Printf("\nCLIENT_SECRET:")
        clientSecret, err := terminal.ReadPassword(int(os.Stdin.Fd()))
        if err != nil {
            return err
        }
        // save the token
        fmt.Printf("\nclient id: %s\nclient secret %s\n", clientId, clientSecret)

        // init google client with it
        return nil
    },
}

func init() {
    rootCmd.AddCommand(initCmd)
}
