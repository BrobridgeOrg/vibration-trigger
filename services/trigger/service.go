package trigger

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"strings"
	app "timer-trigger/app/interface"

	pb "github.com/BrobridgeOrg/timer-api-service/pb"
	"github.com/golang/protobuf/proto"
	"github.com/nats-io/stan.go"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	app       app.AppImpl
	Transport *http.Transport
}

func CreateService(a app.AppImpl) *Service {

	// Preparing service
	service := &Service{
		app: a,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// Send message to specific room
	ebClient := a.GetEventBus().GetClient()
	_, err := ebClient.QueueSubscribe("timer.triggered", "trigger", func(msg *stan.Msg) {

		// Unmarshal message
		var trigger pb.TimerTriggerInfo
		err := proto.Unmarshal(msg.Data, &trigger)
		if err != nil {
			msg.Ack()
			return
		}

		timerInfo := trigger.Info

		// Create a request
		request, err := http.NewRequest(strings.ToUpper(timerInfo.Callback.Method), timerInfo.Callback.Uri, bytes.NewReader([]byte(timerInfo.Payload)))
		if err != nil {
			return
		}

		// Preparing header
		request.Header.Add("Timer-ID", trigger.TimerID)
		request.Header.Add("Content-Type", "application/json")
		for key, value := range timerInfo.Callback.Headers {
			request.Header.Add(key, value)
		}

		client := http.Client{
			Transport: service.Transport,
		}
		resp, err := client.Do(request)
		if err != nil {
			return
		}

		// Require body
		defer resp.Body.Close()

		_, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return
		}

		msg.Ack()
	}, stan.SetManualAckMode())
	if err != nil {
		log.Error(err)
		return nil
	}

	return service
}
