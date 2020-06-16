package app

import (
	"strconv"

	"vibration-trigger/app/eventbus"
	app "vibration-trigger/app/interface"

	"github.com/nats-io/stan.go"
	log "github.com/sirupsen/logrus"
	"github.com/soheilhy/cmux"
	"github.com/sony/sonyflake"
	"github.com/spf13/viper"
)

type App struct {
	id                 uint64
	flake              *sonyflake.Sonyflake
	eventbus           *eventbus.EventBus
	connectionListener cmux.CMux
	grpcServer         *GRPCServer
}

func CreateApp() *App {

	// Genereate a unique ID for instance
	flake := sonyflake.NewSonyflake(sonyflake.Settings{})
	id, err := flake.NextID()
	if err != nil {
		return nil
	}

	idStr := strconv.FormatUint(id, 16)

	return &App{
		id:         id,
		flake:      flake,
		grpcServer: &GRPCServer{},
		eventbus: eventbus.CreateEventBus(
			viper.GetString("service.event_server"),
			viper.GetString("service.event_cluster_id"),
			idStr,
			func(conn stan.Conn, err error) {
				log.Error("event server was disconnected")
			},
		),
	}
}

func (a *App) Init() error {

	log.WithFields(log.Fields{
		"a_id": a.id,
	}).Info("Starting application")

	// Connect to event server
	err := a.eventbus.Connect()
	if err != nil {
		return err
	}

	return nil
}

func (a *App) Uninit() {
}

func (a *App) Run() error {

	port := strconv.Itoa(viper.GetInt("service.port"))
	err := a.CreateConnectionListener(":" + port)
	if err != nil {
		return err
	}

	// gRPC
	go func() {
		err := a.InitGRPCServer(":" + port)
		if err != nil {
			log.Error(err)
		}
	}()

	err = a.Serve()
	if err != nil {
		return err
	}

	return nil
}

func (a *App) GetEventBus() app.EventBusImpl {
	return app.EventBusImpl(a.eventbus)
}
