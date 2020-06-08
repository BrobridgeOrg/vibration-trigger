package app

import "github.com/nats-io/stan.go"

type EventBusImpl interface {
	GetClient() stan.Conn
}

type AppImpl interface {
	GetEventBus() EventBusImpl
}
