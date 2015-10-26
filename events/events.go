package events

import (
	"time"

	"github.com/nats-io/gnatsd/server"
)

type EventType string

const (
	EVENT_BASE_TYPE_CREATED EventType = "created"
	EVENT_BASE_TYPE_UPDATED EventType = "updated"
	EVENT_BASE_TYPE_DELETED EventType = "deleted"
)

/*
const (
	EVENT_ASSET_TYPE_CREATED EventType = "assettype.created"
	EVENT_ASSET_CREATED      EventType = "asset.created"
	EVENT_ASSET_UPDATED      EventType = "asset.updated"
	EVENT_ASSET_DELETED      EventType = "asset.deleted"
)
*/

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

/*
func NewEvent(eventType EventType, payload interface{}) *Event {
	return &Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Payload:   payload,
	}
}
*/

type EventProcessor struct {
	// event queue to read from and publish to backend
	eventQ chan Event
	// backend queueing/messaging system
	qms IQueueMsg

	log server.Logger
}

func NewEventProcessor(ch chan Event, queueMsgSys IQueueMsg, log server.Logger) *EventProcessor {
	return &EventProcessor{eventQ: ch, qms: queueMsgSys, log: log}
}

/*
	Reads events from queue and publishes them the backend MQ system
*/
func (ep *EventProcessor) Start() {
	defer ep.qms.Close()
	for {
		evt := <-ep.eventQ
		if err := ep.qms.Publish(evt); err != nil {
			// TODO: add N retries based on config.
			ep.log.Errorf("Failed to publish event: %s\n", err)
		} else {
			ep.log.Tracef("Event published: %v\n", evt)
		}
	}
}
