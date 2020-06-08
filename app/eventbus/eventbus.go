package eventbus

import (
	nats "github.com/nats-io/nats.go"
	stan "github.com/nats-io/stan.go"
	log "github.com/sirupsen/logrus"
)

type EventBus struct {
	host            string
	clusterID       string
	clientName      string
	client          stan.Conn
	connLostHandler stan.ConnectionLostHandler
}

func CreateEventBus(host string, clusterID string, clientName string, handler stan.ConnectionLostHandler) *EventBus {
	return &EventBus{
		host:            host,
		clusterID:       clusterID,
		clientName:      clientName,
		connLostHandler: handler,
	}
}

func (eb *EventBus) Connect() error {

	log.WithFields(log.Fields{
		"host":       eb.host,
		"clientName": eb.clientName,
		"clusterID":  eb.clusterID,
	}).Info("Connecting to event server")

	// generater natsconnect
	nc, err := nats.Connect(eb.host,
		nats.MaxReconnects(-1),
		nats.ReconnectHandler(func(natsConn *nats.Conn) {
			log.Warn("reconnect successed")
		}),
	)
	if err != nil {
		return err
	}

	// Connect to queue server
	sc, err := stan.Connect(
		eb.clusterID,
		eb.clientName,
		//stan.NatsURL(eb.host),
		stan.NatsConn(nc),
		stan.Pings(10, 3),
		stan.SetConnectionLostHandler(eb.connLostHandler),
	)
	if err != nil {
		return err
	}

	eb.client = sc

	return nil
}

func (eb *EventBus) Close() {
	eb.client.Close()
}

func (eb *EventBus) GetClient() stan.Conn {
	return eb.client
}
