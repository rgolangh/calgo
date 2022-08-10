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
    "context"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "time"

    "google.golang.org/api/option"

    "github.com/spf13/cobra"
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    "google.golang.org/api/calendar/v3"
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
        //fmt.Printf("" +
        //    "Lets setup the google calendar integration:\n" +
        //    "1. enable google calendar API for your account - https://console.cloud.google.com/apis/api/calendar-json.googleapis.com \n" +
        //    "2. create credentials:\n" +
        //    "   - credentials type: user-data,\n" +
        //    "   - scopes: 'https://www.googleapis.com/auth/calendar,\n" +
        //    "3. Go to 'credentials' https://console.cloud.google.com/apis/credentials/\n" +
        //    "   and fill in the details here:\n")
        //
        //fmt.Printf("\nCLIENT_ID:")
        //clientId, err := terminal.ReadPassword(int(os.Stdin.Fd()))
        //if err != nil {
        //    return err
        //}
        //fmt.Printf("\nCLIENT_SECRET:")
        //clientSecret, err := terminal.ReadPassword(int(os.Stdin.Fd()))
        //if err != nil {
        //    return err
        //}
        //// save the token
        //fmt.Printf("\nclient id: %s\nclient secret %s\n", clientId, clientSecret)
        code := make(chan string)
        go func() {
            http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
                code <- request.URL.Query().Get("code")
            })
            http.ListenAndServe("localhost:8111", nil)
        }()

        ctx := context.Background()
        b, err := ioutil.ReadFile("credentials.json")
        if err != nil {
            log.Fatalf("Unable to read client secret file: %v", err)
        }

        // If modifying these scopes, delete your previously saved token.json.
        config, err := google.ConfigFromJSON(b, calendar.CalendarScope)
        if err != nil {
            log.Fatalf("Unable to parse client secret file to config: %v", err)
        }
        client := getClient(config, code)

        srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
        if err != nil {
            log.Fatalf("Unable to retrieve Calendar client: %v", err)
        }

        t := time.Now().Format(time.RFC3339)
        events, err := srv.Events.List("primary").ShowDeleted(false).
            SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
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
                fmt.Printf("%v (%v)\n", item.Summary, date)
            }
        }
        // init google client with it
        return nil
    },
}

func init() {
    rootCmd.AddCommand(initCmd)
}

func getTokenFromWeb(config *oauth2.Config, code <-chan string) *oauth2.Token {
    authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
    fmt.Printf("Go to the following link in your browser then type the "+
        "authorization code: \n%v\n", authURL)

    var authCode string
    select {
    case authCode = <-code:
        fmt.Printf("got the auth-code %s\n", authCode)
    }

    tok, err := config.Exchange(context.TODO(), authCode)
    if err != nil {
        log.Fatalf("Unable to retrieve token from web: %v", err)
    }
    return tok
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config, code <-chan string) *http.Client {
    // The file token.json stores the user's access and refresh tokens, and is
    // created automatically when the authorization flow completes for the first
    // time.
    tokFile := "token.json"
    tok, err := tokenFromFile(tokFile)
    if err != nil {
        tok = getTokenFromWeb(config, code)
        saveToken(tokFile, tok)
    }
    return config.Client(context.Background(), tok)
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
    f, err := os.Open(file)
    if err != nil {
        return nil, err
    }
    defer f.Close()
    tok := &oauth2.Token{}
    err = json.NewDecoder(f).Decode(tok)
    return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
    fmt.Printf("Saving credential file to: %s\n", path)
    f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
    if err != nil {
        log.Fatalf("Unable to cache oauth token: %v", err)
    }
    defer f.Close()
    json.NewEncoder(f).Encode(token)
}
