package events

import (
	"github.com/nats-io/gnatsd/server"
	"github.com/nats-io/nats"
)

type NatsClient struct {
	conn        *nats.Conn
	encodedConn *nats.EncodedConn

	log server.Logger
}

func NewNatsClient(servers []string, log server.Logger) (nclient *NatsClient, err error) {
	opts := nats.DefaultOptions
	opts.Servers = servers

	nclient = &NatsClient{log: log}

	if nclient.conn, err = opts.Connect(); err != nil {
		return
	}

	log.Debugf("nats client connected to: %v!\n", nclient.conn.ConnectedUrl())

	nclient.conn.Opts.ReconnectedCB = func(nc *nats.Conn) {
		log.Debugf("nats client reconnected to: %v!\n", nc.ConnectedUrl())
	}

	nclient.conn.Opts.DisconnectedCB = func(_ *nats.Conn) {
		log.Debugf("nats client disconnected!\n")
	}

	nclient.encodedConn, err = nats.NewEncodedConn(nclient.conn, nats.JSON_ENCODER)

	return
}

func (nc *NatsClient) GetConn() *nats.EncodedConn {
	return nc.encodedConn
}

func (nc *NatsClient) Publish(evt Event) error {
	nc.log.Tracef("Publishing: %v\n", evt)

	return nc.encodedConn.Publish(string(evt.Type), &evt)
}

func (nc *NatsClient) Subscribe(topic string, ch chan *Event) (err error) {

	if err = nc.encodedConn.BindSendChan(topic, ch); err != nil {
		return
	}

	if _, err = nc.encodedConn.BindRecvChan(topic, ch); err != nil {
		return
	}
	return
}

func (nc *NatsClient) Close() {
	nc.encodedConn.Close()
	nc.conn.Close()
}
