// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/calendar/v3"
	"testing"
	"time"
)

func TestInsert(t *testing.T) {
	events := newEvents()
	events.insert(&calendar.Event{
		Summary: "the 2nd",
		Start: &calendar.EventDateTime{
			DateTime: time.Date(2023, 9, 24, 8, 0, 0, 0, time.UTC).Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: time.Date(2023, 9, 24, 8, 50, 0, 0, time.UTC).Format(time.RFC3339),
		},
	})
	if events.Len() != 1 {
		t.Errorf("wanted size of 1 but got %d", events.Len())
	}
}

func TestAddAll(t *testing.T) {
	events := newEvents()
	eventsFromCal := []*calendar.Event{
		{
			Summary: "the last",
			Start: &calendar.EventDateTime{
				DateTime: time.Date(2023, 9, 24, 18, 0, 0, 0, time.UTC).Format(time.RFC3339),
			},
			End: &calendar.EventDateTime{
				DateTime: time.Date(2023, 9, 24, 19, 0, 0, 0, time.UTC).Format(time.RFC3339),
			},
		},
		{
			Summary: "before the last",
			Start: &calendar.EventDateTime{
				DateTime: time.Date(2023, 9, 24, 15, 0, 0, 0, time.UTC).Format(time.RFC3339),
			},
			End: &calendar.EventDateTime{
				DateTime: time.Date(2023, 9, 24, 16, 0, 0, 0, time.UTC).Format(time.RFC3339),
			},
		},
	}

	events.addAll(eventsFromCal)
	assert.Equal(t, events.Len(), len(eventsFromCal))
}

func TestInsertOrder(t *testing.T) {
	events := newEvents()

	events.insert(&calendar.Event{
		Summary: "the 2nd",
		Start: &calendar.EventDateTime{
			DateTime: time.Date(2023, 9, 24, 8, 0, 0, 0, time.UTC).Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: time.Date(2023, 9, 24, 8, 50, 0, 0, time.UTC).Format(time.RFC3339),
		},
	})

	events.insert(&calendar.Event{
		Summary: "the 1st",
		Start: &calendar.EventDateTime{
			DateTime: time.Date(2023, 9, 24, 7, 0, 0, 0, time.UTC).Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: time.Date(2023, 9, 24, 7, 50, 0, 0, time.UTC).Format(time.RFC3339),
		},
	})

	events.insert(&calendar.Event{
		Summary: "the 4th",
		Start: &calendar.EventDateTime{
			DateTime: time.Date(2023, 9, 24, 11, 0, 0, 0, time.UTC).Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: time.Date(2023, 9, 24, 11, 50, 0, 0, time.UTC).Format(time.RFC3339),
		},
	})

	events.insert(&calendar.Event{
		Summary: "the 3rd",
		Start: &calendar.EventDateTime{
			DateTime: time.Date(2023, 9, 24, 10, 0, 0, 0, time.UTC).Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: time.Date(2023, 9, 24, 10, 50, 0, 0, time.UTC).Format(time.RFC3339),
		},
	})

	events.insert(&calendar.Event{
		Summary: "the 1.5",
		Start: &calendar.EventDateTime{
			DateTime: time.Date(2023, 9, 24, 7, 56, 0, 0, time.UTC).Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: time.Date(2023, 9, 24, 7, 58, 0, 0, time.UTC).Format(time.RFC3339),
		},
	})

	eventsFromCal := []*calendar.Event{
		{
			Summary: "the last",
			Start: &calendar.EventDateTime{
				DateTime: time.Date(2023, 9, 24, 18, 0, 0, 0, time.UTC).Format(time.RFC3339),
			},
			End: &calendar.EventDateTime{
				DateTime: time.Date(2023, 9, 24, 19, 0, 0, 0, time.UTC).Format(time.RFC3339),
			},
		},
		{
			Summary: "before the last",
			Start: &calendar.EventDateTime{
				DateTime: time.Date(2023, 9, 24, 15, 0, 0, 0, time.UTC).Format(time.RFC3339),
			},
			End: &calendar.EventDateTime{
				DateTime: time.Date(2023, 9, 24, 16, 0, 0, 0, time.UTC).Format(time.RFC3339),
			},
		},
	}
	events.addAll(eventsFromCal)

	wantedOrder := []string{
		"the 1st",
		"the 1.5",
		"the 2nd",
		"the 3rd",
		"the 4th",
		"before the last",
		"the last",
	}
	c := events.Front()
	for _, wantedSummary := range wantedOrder {
		if c == nil {
			t.Errorf("no more event in the list. Either the list of events changed , or the wanted list changed.")
		}
		if assert.Equal(t, wantedSummary, c.Value.(*calendar.Event).Summary) {
			c = c.Next()
			continue
		}

	}
}
