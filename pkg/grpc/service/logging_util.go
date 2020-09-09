package service

import (
	"context"
	"log"

	loggingAPI "github.com/influenzanet/logging-service/pkg/api"
)

func (s *studyServiceServer) SaveLogEvent(
	instanceID string,
	userID string,
	eventType loggingAPI.LogEventType,
	eventName string,
	msg string,
) {
	_, err := s.clients.LoggingService.SaveLogEvent(context.TODO(), &loggingAPI.NewLogEvent{
		Origin:     "study-service",
		InstanceId: instanceID,
		UserId:     userID,
		EventType:  eventType,
		EventName:  eventName,
		Msg:        msg,
	})
	if err != nil {
		log.Printf("ERROR: failed to save log: %s", err.Error())
	}
}
