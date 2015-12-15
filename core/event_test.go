package core

import (
	"testing"
)

func Test_NewEvent(t *testing.T) {
	evt := NewEvent(EVENT_BASE_TYPE_CREATED, "server", nil)
	if evt.Type != "created.server" {
		t.Fatal("Wrong event type")
	}
}
