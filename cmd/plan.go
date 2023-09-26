// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"bytes"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/rgolangh/calgo/internal/google_calendar"
	"github.com/spf13/cobra"
	"google.golang.org/api/calendar/v3"
	"log"
	"time"
)

var (
	interactive        bool
	meetingsTime       time.Duration
	focusTime          time.Duration
	focusEventDuration time.Duration
	tasks              time.Duration
)

type Slot struct {
	StartTime time.Time
	EndTime   time.Time
}

type Focus struct {
	Duration time.Duration
	Title    string `default:"Focus Time"`
}
type Meeting struct {
	Duration    time.Duration
	Title       string
	Description string
	Attendees   []*calendar.EventAttendee
}

type Plan struct {
	eventInserter func(event *calendar.Event) (*calendar.Event, error)
	calendarId    string
	// target date of plan, either today or future
	date             time.Time
	events           Events
	overallFocusTime time.Duration
	focusDuration    time.Duration
	slots            []Slot
	focuses          []Focus
	meetings         []Meeting
}

func newPlan(calId string, service *calendar.Service) *Plan {
	events, err := service.Events.List(calendarID).
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(time.Now().Format(time.RFC3339)).
		TimeMax(endOfDay(time.Now()).Format(time.RFC3339)).
		OrderBy("startTime").
		Do()
	if err != nil {
		log.Fatalf("Unable to retrieve next ten of the user's events: %v", err)
	}
	eventInserter := func(event *calendar.Event) (*calendar.Event, error) {
		return service.Events.Insert(calendarID, event).Do()
	}

	plannedEvents := newEvents()
	plannedEvents.addAll(events.Items)
	return &Plan{
		date:             time.Now(),
		eventInserter:    eventInserter,
		calendarId:       calId,
		overallFocusTime: focusTime,
		focusDuration:    focusEventDuration,
		events:           plannedEvents,
	}
}

func (p *Plan) plan() error {
	for focusDurationToAdd := p.overallFocusTime; focusDurationToAdd >= p.focusDuration; focusDurationToAdd -= p.focusDuration {
		slot, err := p.findNextSlot()
		if err != nil {
			log.Printf("failed finding slot %v\n", err)
			continue
		}
		log.Printf("found a slot on %v\n", slot.Format(time.Kitchen))
		p.events.insert(newFocusEvent(slot, p.focusDuration))
	}
	return nil
}

func (p *Plan) commit() error {
	if interactive {
		var commit bool
		err := survey.AskOne(&survey.Confirm{Message: "Commit changes to the calendar?", Default: true}, &commit)
		if err != nil {
			return err
		}
		if !commit {
			return nil
		}
	}
	for _, newEvent := range p.getAddedEvents() {
		_, err := p.eventInserter(newEvent)
		if err != nil {
			log.Fatalf("Unable to create event. %v\n", err)
		}
	}
	return nil
}

func (p *Plan) String() string {
	buf := bytes.NewBufferString("")
	fmt.Fprintf(buf, "\n\nPlan for %v %d events\n", p.date.Format(time.RFC822), p.events.Len())
	for e := p.events.Front(); e != nil; e = e.Next() {
		v := e.Value.(*calendar.Event)
		fmt.Fprintf(buf, eventString(v))
	}
	return buf.String()
}

// planCmd represents the plan command
var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Plan your day",
	Long:  `Plan your day, add meetings, focus times, and break time.`,
	Example: `$ calgo plan --focus-time 5h --tasks 1 --break 1
# 
dd/mm/yy, today, 0 meetings
[focus time] duration(minutes) for each interval?(45): 50
[focus time] optional event name?(focus time): create calgo
✔ done

[meeting 1] event name? discuss new requirements
[meeting 1] duration?(50m):
[meeting 1] attendants (tab to autocomplete, enter twice to done):
rgo(tab) - rgolan@redhat.com
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		srv := google_calendar.Service()

		//focuses := surveyFocus()
		//meetings := surveyMeetings()
		plan := newPlan(calendarID, srv)
		err := plan.plan()
		if err != nil {
			return err
		}
		log.Println(plan)
		plan.commit()

		return nil
	},
	Args: func(cmd *cobra.Command, args []string) error {
		if focusTime < focusEventDuration {
			return fmt.Errorf("--focus-time (%s) must be greater then or equal to --focus-event-duration (%s)", focusTime, focusEventDuration)
		}
		return nil
	},
}

func surveyFocus() []Focus {
	focuses := []Focus{}
	var duration = time.Minute * 45
	var title = "Focus Time"
	// prevent 0, make sure = < focusTime
	survey.AskOne(&survey.Input{Message: "duration(minutes) for each focus slot?", Default: "45m"}, &duration)
	survey.AskOne(&survey.Input{Message: "focus time title?", Default: "Focus Time"}, &title)

	var nFocuses = int(focusTime / duration)
	for i := 0; i < nFocuses; i++ {
		focuses = append(focuses, Focus{
			Duration: duration,
			Title:    title,
		})
	}

	return focuses
}

func surveyMeetings() []Meeting {
	if meetingsTime == 0 {
		return nil
	}
	meetings := []Meeting{}
	var duration = time.Minute * 45
	var title = ""
	// prevent 0, make sure = < focusTime
	survey.AskOne(&survey.Input{Message: "duration in minutes for each meeting?", Default: "45m"}, &duration)
	survey.AskOne(&survey.Input{Message: "meeting title?"}, &title)

	var nFocuses = int(meetingsTime / duration)
	for i := 0; i < nFocuses; i++ {
		meetings = append(meetings, Meeting{
			Duration: duration,
			Title:    title,
		})
	}

	return meetings
}

func newFocusEvent(startTime time.Time, duration time.Duration) *calendar.Event {
	return &calendar.Event{
		Summary:     "Focus Time",
		Description: "Focus Time",
		EventType:   "focusTime",
		Start: &calendar.EventDateTime{
			DateTime: startTime.Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: startTime.Add(duration).Format(time.RFC3339),
		},
		//Attendees: []*calendar.EventAttendee{
		//	&calendar.EventAttendee{Email:"lpage@example.com"},
		//	&calendar.EventAttendee{Email:"sbrin@example.com"},
		//},
	}
}

func (p *Plan) getAddedEvents() []*calendar.Event {
	events := []*calendar.Event{}

	for elm := p.events.Front(); elm != nil; elm = elm.Next() {
		e := elm.Value.(*calendar.Event)
		// events with no ID are new ones, not added yet.
		if e.Id == "" {
			events = append(events, e)
		}
	}
	return events
}

// find the next slot to add an event
func (p *Plan) findNextSlot() (time.Time, error) {
	// walk the list of events, which are sorted by start time, measure
	// a free period of the wanted duration by starting a point in time, now for
	// today, or start of day if in future.
	// measure from that markpoint to next meeting start date , test if it can
	// contain the period.
	// TODO handle overlapping meetings to calculate the available periods.
	// currently naively assuming there is no overlap.
	// advanced the markpoint to the end of current meeting and start again

	var markpoint time.Time
	y, m, d := p.date.Date()
	y1, m1, d1 := time.Now().Date()
	if y == y1 && m == m1 && d == d1 {
		s := startOfDay(time.Now())
		// it maybe that the day already started and we want to plan. if now
		// is later the startofday then use it.
		if time.Now().After(s) {
			markpoint = time.Now()
		} else {
			markpoint = s
		}
	} else {
		markpoint = startOfDay(p.date)
	}
	for elm := p.events.Front(); elm != nil; elm = elm.Next() {
		nextEvent := elm.Value.(*calendar.Event)
		nextEventStartTime, err := time.Parse(time.RFC3339, nextEvent.Start.DateTime)
		if err != nil {
			return markpoint, err
		}
		if markpoint.Add(p.focusDuration).Before(nextEventStartTime) {
			return markpoint, nil
		}

		nextEventEndTime, err := time.Parse(time.RFC3339, nextEvent.End.DateTime)
		if err != nil {
			return markpoint, err
		}
		markpoint = nextEventEndTime
		if elm.Next() == nil {
			// last element
			if markpoint.Add(p.focusDuration).Before(endOfDay(p.date)) {
				return markpoint, nil
			}
		}
	}
	log.Println("couldn't find a slot")
	return markpoint, fmt.Errorf("couldn't not find a slot")
}

func init() {
	planCmd.Flags().BoolVar(&interactive, "interactive", true, "Ask before committing changes, ask optional inputs")
	planCmd.Flags().DurationVar(&focusTime, "focus-time", time.Minute*45, "desired overall focus time duration (e.g 45m, 1h20m)")
	planCmd.Flags().DurationVar(&focusEventDuration, "focus-event-duration", time.Minute*45, "desired focus time per event. An overall focus time is devided to events (e.g 45m, 1h20m)")
	planCmd.Flags().DurationVar(&meetingsTime, "meetings", 0, "desired meetings overall time duration (e.g 1h30m")
	planCmd.Flags().DurationVar(&tasks, "break", time.Hour, "desired break time duration (e.g 1h)")
	rootCmd.AddCommand(planCmd)
}

var eventFormat = func(title, startTime, endTime string) string {
	return fmt.Sprintf("%17s: %s", startTime+"-"+endTime, title)
}
