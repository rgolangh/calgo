// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"container/list"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/calendar/v3"
	"testing"
	"time"
)

type testCase struct {
	focusTime      time.Duration
	focusDuration  time.Duration
	meetingTime    time.Duration
	existingEvents []*calendar.Event
	wantedEvents   []*calendar.Event
	expectedErr    error
}

var testCases = []testCase{
	{
		focusTime:     45 * time.Minute,
		focusDuration: 45 * time.Minute,
		meetingTime:   0 * time.Minute,
		existingEvents: []*calendar.Event{
			{
				Id: "1", // Id is an indication this event is persisted on the calendar
				Start: &calendar.EventDateTime{
					DateTime: time.Date(2023, 9, 24, 8, 0, 0, 0, time.UTC).Format(time.RFC3339),
				},
				End: &calendar.EventDateTime{
					DateTime: time.Date(2023, 9, 24, 8, 50, 0, 0, time.UTC).Format(time.RFC3339),
				},
			},
		},
		wantedEvents: []*calendar.Event{
			{
				Summary:     "Focus Time",
				Description: "Focus Time",
				EventType:   "focusTime",
				Start: &calendar.EventDateTime{
					DateTime: time.Date(2023, 9, 24, 8, 50, 0, 0, time.UTC).Format(time.RFC3339),
				},
				End: &calendar.EventDateTime{
					DateTime: time.Date(2023, 9, 24, 9, 35, 0, 0, time.UTC).Format(time.RFC3339),
				},
			},
		},
		expectedErr: nil,
	},
	//{
	//	focusTime:      0,
	//	focusDuration:  0,
	//	meetingTime:    0,
	//	existingEvents: nil,
	//	wantedEvents:   nil,
	//	expectedErr:    fmt.Errorf("failed to plan, focusDuration is 0"),
	//},
}

func TestPlan(t *testing.T) {
	for _, tc := range testCases {
		planedEvents := Events{list.New()}
		planedEvents.addAll(tc.existingEvents)
		eventInserter := func(event *calendar.Event) (*calendar.Event, error) {
			return event, nil
		}

		p := &Plan{
			date:             time.Now(),
			overallFocusTime: tc.focusTime,
			focusDuration:    tc.focusDuration,
			eventInserter:    eventInserter,
			events:           planedEvents,
		}

		err := p.plan()
		assert.ErrorIs(t, tc.expectedErr, err)
		assert.Equal(t, tc.wantedEvents, p.getAddedEvents())
	}
}
