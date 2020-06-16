package app

import (
	log "github.com/sirupsen/logrus"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"

	app "vibration-trigger/app/interface"
	pb "vibration-trigger/pb"
	"vibration-trigger/services/trigger"
)

type GRPCServer struct {
	Trigger *trigger.Service
}

func (a *App) InitGRPCServer(host string) error {

	lis := a.connectionListener.MatchWithWriters(
		cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"),
	)

	log.WithFields(log.Fields{
		"host": host,
	}).Info("Starting gRPC server on " + host)

	// Create gRPC server
	s := grpc.NewServer()

	// Register data source adapter service
	triggerService := trigger.CreateService(app.AppImpl(a))
	a.grpcServer.Trigger = triggerService
	pb.RegisterTriggerServer(s, triggerService)
	//reflection.Register(s)

	log.WithFields(log.Fields{
		"service": "Trigger",
	}).Info("Registered service")

	// Starting server
	if err := s.Serve(lis); err != cmux.ErrListenerClosed {
		log.Fatal(err)
		return err
	}

	return nil
}
