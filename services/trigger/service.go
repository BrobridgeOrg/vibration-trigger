package trigger

import (
	"bytes"
	"crypto/tls"
	"time"
	//"io/ioutil"
	"net/http"
	"strings"
	app "vibration-trigger/app/interface"

	pb "github.com/BrobridgeOrg/vibration-api-service/pb"
	"github.com/golang/protobuf/proto"
	"github.com/nats-io/stan.go"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	app        app.AppImpl
	Transport  *http.Transport
	retryCount int64
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

		service.retryCount = 1
		err = service.Query(trigger)
		if err != nil {
			log.Println(err)
			msg.Ack()
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

func (service *Service) Query(trigger pb.TimerTriggerInfo) error {

	timerInfo := trigger.Info

	// Create a request
	request, err := http.NewRequest(strings.ToUpper(timerInfo.Callback.Method), timerInfo.Callback.Uri, bytes.NewReader([]byte(timerInfo.Payload)))
	if err != nil {
		// if retry 3 times than return error
		if service.retryCount > 3 {
			return err
		} else {
			// delay service.retryCount * 3
			timer := time.NewTimer(time.Duration(service.retryCount*3) * time.Second)
			<-timer.C
			service.retryCount++
			return service.Query(trigger)
		}
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
	_, err = client.Do(request)
	if err != nil {
		// if retry 3 times than return error
		if service.retryCount > 3 {
			return err
		} else {
			// delay service.retryCount * 3
			timer := time.NewTimer(time.Duration(service.retryCount*3) * time.Second)
			<-timer.C
			service.retryCount++
			return service.Query(trigger)
		}
	}

	return nil
}
