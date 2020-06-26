package eventbus

import (
	"time"

	nats "github.com/nats-io/nats.go"
	stan "github.com/nats-io/stan.go"
	log "github.com/sirupsen/logrus"
)

type EventBus struct {
	host              string
	clusterID         string
	clientName        string
	natsConn          *nats.Conn
	client            stan.Conn
	reconnectHandler  func(natsConn *nats.Conn)
	disconnectHandler func(natsConn *nats.Conn)
}

func CreateEventBus(host string, clusterID string, clientName string, reconnectHandler func(natsConn *nats.Conn), disconnectHandler func(natsConn *nats.Conn)) *EventBus {
	return &EventBus{
		host:              host,
		clusterID:         clusterID,
		clientName:        clientName,
		reconnectHandler:  reconnectHandler,
		disconnectHandler: disconnectHandler,
	}
}

func (eb *EventBus) Connect() error {

	log.WithFields(log.Fields{
		"host":       eb.host,
		"clientName": eb.clientName,
		"clusterID":  eb.clusterID,
	}).Info("Connecting to event server")

	if eb.natsConn == nil {
		// generater natsconnect
		nc, err := nats.Connect(eb.host,
			nats.PingInterval(10*time.Second),
			nats.MaxPingsOutstanding(3),
			nats.MaxReconnects(-1),
			nats.ReconnectHandler(eb.reconnectHandler),
			nats.DisconnectHandler(eb.disconnectHandler),
		)
		if err != nil {
			return err
		}

		eb.natsConn = nc
	}

	// Connect to queue server
	sc, err := stan.Connect(
		eb.clusterID,
		eb.clientName,
		stan.NatsConn(eb.natsConn),
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
