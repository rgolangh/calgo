// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"container/list"
	"google.golang.org/api/calendar/v3"
	"time"
)

type Events struct {
	*list.List
}

// insert and event , insertion order is by datetime
func (p *Events) insert(event *calendar.Event) {
	if p.Len() == 0 {
		p.PushFront(event)
		return
	}
	for currentElement := p.Front(); currentElement != nil; currentElement = currentElement.Next() {
		candidateEndTime, _ := time.Parse(time.RFC3339, event.End.DateTime)
		currentEvent := currentElement.Value.(*calendar.Event)
		currentStartTime, _ := time.Parse(time.RFC3339, currentEvent.Start.DateTime)
		if candidateEndTime.Before(currentStartTime) {
			p.InsertBefore(event, currentElement)
			return
		}
	}
	p.PushBack(event)
}

func (p *Events) addAll(events []*calendar.Event) {
	for _, e := range events {
		p.insert(e)
	}
}

func newEvents() Events {
	return Events{list.New()}
}
