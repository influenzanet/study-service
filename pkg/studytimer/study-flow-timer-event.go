package main

import (
	"log"

	"github.com/influenzanet/study-service/pkg/dbs/globaldb"
	"github.com/influenzanet/study-service/pkg/dbs/studydb"
	"github.com/influenzanet/study-service/pkg/studyengine"
	"github.com/influenzanet/study-service/pkg/types"
)

type StudyTimerService struct {
	globalDBService            *globaldb.GlobalDBService
	studyDBService             *studydb.StudyDBService
	TimerEventFrequency        int64 // how often the timer event should be performed (only from one instance of the service) - seconds
	TimerEventCheckIntervalMin int   // approx. how often this serice should check if to perform the timer event - seconds
	TimerEventCheckIntervalVar int   // range of the uniform random distribution - varying the check interval to avoid a steady collisions
}

func (s *StudyTimerService) StudyTimerEvent() {
	instances, err := s.globalDBService.GetAllInstances()
	if err != nil {
		log.Printf("unexpected error: %s", err.Error())
	}
	for _, instance := range instances {
		studies, err := s.studyDBService.GetStudiesByStatus(instance.InstanceID, "active", true)
		if err != nil {
			log.Printf("unexpected error: %s", err.Error())
			return
		}
		for _, study := range studies {
			if err := s.studyDBService.ShouldPerformTimerEvent(instance.InstanceID, study.Key, s.TimerEventFrequency); err != nil {
				continue
			}
			log.Printf("performing timer event for study: %s - %s", instance.InstanceID, study.Key)

			if err := s.studyDBService.FindAndExecuteOnParticipantsStates(instance.InstanceID, study.Key, s.checkAndUpdateParticipantState); err != nil {
				log.Println(err)
				continue
			}
		}
	}
}

func (s *StudyTimerService) checkAndUpdateParticipantState(studyDBServ *studydb.StudyDBService, pState types.ParticipantState, instanceID string, studyKey string) error {
	rules, err := studyDBServ.GetStudyRules(instanceID, studyKey)
	if err != nil {
		return err
	}

	studyEvent := types.StudyEvent{
		Type: "TIMER",
	}

	for _, rule := range rules {
		pState, err = studyengine.ActionEval(rule, pState, studyEvent)
		if err != nil {
			log.Println(err)
			continue
		}
	}
	// save state back to DB
	_, err = studyDBServ.SaveParticipantState(instanceID, studyKey, pState)
	return err
}
