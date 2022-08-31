package cmd

import (
	"testing"
	"time"
)

func TestDateExpressionParsing(t *testing.T) {
	var tuesdayThe30th, _ = time.Parse(time.RFC3339, "2022-08-30T15:04:05Z03:00") // Tuesday of some week

	cases := []struct {
		in                 string
		currrentTime       time.Time
		expectedMinWeekday time.Weekday
		expectedMinDateDay int
		expectedMaxWeekday time.Weekday
		expectedMaxDateDay int
	}{
		{
			in:                 "1",
			currrentTime:       tuesdayThe30th,
			expectedMinWeekday: time.Sunday,
			expectedMinDateDay: 4,
			expectedMaxWeekday: time.Sunday,
			expectedMaxDateDay: 4,
		},
		{
			in:                 "s",
			currrentTime:       tuesdayThe30th,
			expectedMinWeekday: time.Sunday,
			expectedMinDateDay: 4,
			expectedMaxWeekday: time.Sunday,
			expectedMaxDateDay: 4,
		},
		{
			in:                 "5",
			currrentTime:       tuesdayThe30th,
			expectedMinWeekday: time.Thursday,
			expectedMinDateDay: 1,
			expectedMaxWeekday: time.Thursday,
			expectedMaxDateDay: 1,
		},
		{
			// don't expect negative values to be common. but maybe
			// it will be handy for viewing past events?
			in:                 "-1",
			currrentTime:       tuesdayThe30th,
			expectedMinWeekday: time.Monday,
			expectedMinDateDay: 29,
			expectedMaxWeekday: time.Monday,
			expectedMaxDateDay: 29,
		},
		{
			in:                 "1-2",
			currrentTime:       tuesdayThe30th,
			expectedMinWeekday: time.Sunday,
			expectedMinDateDay: 4,
			expectedMaxWeekday: time.Monday,
			expectedMaxDateDay: 5,
		},
		{
			in:                 "s-m",
			currrentTime:       tuesdayThe30th,
			expectedMinWeekday: time.Sunday,
			expectedMinDateDay: 4,
			expectedMaxWeekday: time.Monday,
			expectedMaxDateDay: 5,
		},
		{
			in:                 "t-sa",
			currrentTime:       tuesdayThe30th,
			expectedMinWeekday: time.Tuesday,
			expectedMinDateDay: 30,
			expectedMaxWeekday: time.Saturday,
			expectedMaxDateDay: 3,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.in, func(t *testing.T) {
			tmin, tmax, err := parseDatetimeExpression(testCase.in)
			if err != nil {
				t.Errorf("failed get min max times from %q error: %v", testCase.in, err)
			}
			if tmin.Weekday() != testCase.expectedMinWeekday {
				t.Errorf("expected the weekday to be %s but got %s", testCase.expectedMinWeekday, tmin.Weekday().String())
			}
			if tmin.Day() != testCase.expectedMinDateDay {
				t.Errorf("expected the day to be %d but got %d", testCase.expectedMinDateDay, tmin.Day())
			}
			if tmax.Weekday() != testCase.expectedMaxWeekday {
				t.Errorf("expected the max weekday to be %s but got %s", testCase.expectedMaxWeekday, tmax.Weekday().String())
			}
			if tmax.Day() != testCase.expectedMaxDateDay {
				t.Errorf("expected the day to be %d but got %d", testCase.expectedMaxDateDay, tmax.Day())
			}
		})
	}
}
