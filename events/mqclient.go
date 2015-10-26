package events

type IQueueMsg interface {
	Publish(evt Event) error
	Close()
}
