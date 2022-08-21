package cmd

import (
    "testing"
    "time"
)

var testTime time.Time

func TestGetMinMaxTimes(t *testing.T) {
    testTime, _ = time.Parse(time.RFC3339, "2022-08-20T15:04:05+03:00")

    cases := []struct {
        in              string
        currrentTime    string
        expectedMinTime string
        expectedMaxTime string
    }{
        {
            in:              "1",
            currrentTime:    testTime.String(),
            expectedMinTime: "2022-08-14T15:04:05+03:00",
            expectedMaxTime: "2022-08-20T15:04:05+03:00",
        },
    }

    for _, testCase := range cases {
        tmin, tmax, err := getMinMaxStartTimes(testCase.in)
        if err != nil {
            t.Errorf("failed get min max times from %s %e", testCase.in, err)
        }
        if tmin.Format(time.RFC3339) != testCase.expectedMinTime {
            t.Errorf("expecte min time %s but got %s", testCase.expectedMinTime, tmin)
        }
        if tmax.Format(time.RFC3339) != testCase.expectedMaxTime {
            t.Errorf("expecte max time %s but got %s", testCase.expectedMaxTime, tmax)
        }
    }
}
