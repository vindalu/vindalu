package events

import (
	"github.com/nats-io/gnatsd/server"

	"github.com/vindalu/vindalu/core"
)

type EventProcessor struct {
	// event queue to read from and publish to backend
	eventQ chan core.Event
	// backend queueing/messaging system
	qms IQueueMsg

	log server.Logger
}

func NewEventProcessor(ch chan core.Event, queueMsgSys IQueueMsg, log server.Logger) *EventProcessor {
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
			// TODO: add MaxTries MaxTimeout
			ep.log.Errorf("Failed to publish: %s\n", err)
		} else {
			ep.log.Tracef("Event published: %v\n", evt)
		}
	}
}
