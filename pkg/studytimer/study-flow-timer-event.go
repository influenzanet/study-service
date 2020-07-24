package studytimer

import (
	"log"

	"github.com/influenzanet/study-service/pkg/dbs/studydb"
	"github.com/influenzanet/study-service/pkg/studyengine"
	"github.com/influenzanet/study-service/pkg/types"
)

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

			s.UpdateStudyStats(instance.InstanceID, study.Key)

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

func (s *StudyTimerService) UpdateStudyStats(instanceID string, studyKey string) {
	pCount, err := s.studyDBService.GetParticipantCountByStatus(instanceID, studyKey, types.PARTICIPANT_STUDY_STATUS_ACTIVE)
	if err != nil {
		log.Printf("DB ERROR for participant counting for study: %s -> %s", studyKey, err.Error())
	}
	rCount, err := s.studyDBService.CountSurveyResponsesByKey(instanceID, studyKey, "", 0, 0)
	if err != nil {
		log.Printf("DB ERROR for response counting for study: %s -> %s", studyKey, err.Error())
	}

	if err := s.studyDBService.UpdateStudyStats(instanceID, studyKey, types.StudyStats{
		ParticipantCount: pCount,
		ResponseCount:    rCount,
	}); err != nil {
		log.Printf("DB ERROR for updating stats for study: %s -> %s", studyKey, err.Error())
	}
}
