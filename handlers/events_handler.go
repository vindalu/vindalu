package handlers

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/vindalu/vindalu/events"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

/*
	Ignore client messages to prevent buffers from filling up.
*/
func (ir *Inventory) websocketReadLoop(c *websocket.Conn, closeConn chan bool) {
	//
	// Call reader once (blocking call) and close connection if any data is recieved.  Writing
	// is not allowed.
	//
	_, _, err := c.NextReader()
	if err != nil {
		ir.log.Errorf("Closing websocket connection: %s\n", err)
	} else {
		ir.log.Errorf("Client attempted to write event! Closing connection!\n")
	}

	c.WriteJSON(map[string]string{"error": "Cannot write to stream!"})
	c.Close()

	ir.log.Noticef("Signalling websocket connection close\n")
	// Signal close
	closeConn <- true
}

/* This handler needs to be registered with a wrapper to inject the pub-sub port */
func (ir *Inventory) WebsocketHandler(w http.ResponseWriter, r *http.Request, pubSubPort int) {
	urlVars := mux.Vars(r)
	if len(urlVars["topic"]) < 1 {
		ir.writeAndLogResponse(w, r, 400, map[string]string{"Content-Type": "text/plain"}, []byte(`Invalid topic`))
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		ir.writeAndLogResponse(w, r, 500, map[string]string{"Content-Type": "text/plain"}, []byte(err.Error()))
		return
	}
	defer conn.Close()

	// Create new subscriber client
	client, err := events.NewNatsClient([]string{fmt.Sprintf("nats://localhost:%d", pubSubPort)}, ir.log)
	if err != nil {
		ir.writeAndLogResponse(w, r, 500, map[string]string{"Content-Type": "text/plain"}, []byte(err.Error()))
		return
	}
	defer client.Close()

	var (
		reChan    = make(chan *events.Event)
		closeChan = make(chan bool)
	)

	// Burn client messages
	go ir.websocketReadLoop(conn, closeChan)
	// Subscribe to events
	if err = client.Subscribe(urlVars["topic"], reChan); err != nil {
		ir.writeAndLogResponse(w, r, 500, map[string]string{"Content-Type": "text/plain"}, []byte(err.Error()))
		return
	}

	// Read events
	for {
		select {
		case <-closeChan:
			ir.log.Noticef("Closing websocket client connection!\n")
			return
		case msg := <-reChan:
			ir.log.Tracef("Sending to websocket client: %v\n", msg)
			if err = conn.WriteJSON(msg); err != nil {
				ir.log.Errorf("Websocket write: %s\n", err)
				conn.Close()
			}
			break
		}
	}
}
