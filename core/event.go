package core

import (
	"time"
)

type EventType string

const (
	EVENT_BASE_TYPE_CREATED EventType = "created"
	EVENT_BASE_TYPE_UPDATED EventType = "updated"
	EVENT_BASE_TYPE_DELETED EventType = "deleted"
)

type Event struct {
	Type      string
	Timestamp time.Time
	Payload   interface{}
}

func NewEvent(evtType EventType, subtype string, payload interface{}) (evt *Event) {
	evt = &Event{
		Timestamp: time.Now(),
		Payload:   payload,
		Type:      string(evtType),
	}

	if len(subtype) > 0 {
		evt.Type += "." + subtype
	}
	return
}
