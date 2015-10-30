package events

import (
	"github.com/vindalu/vindalu/core"
)

type IQueueMsg interface {
	Publish(evt core.Event) error
	Close()
}
