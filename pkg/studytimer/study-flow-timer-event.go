package studytimer

import (
	"errors"
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
			s.UpdateParticipantStates(instance.InstanceID, study.Key)
		}
	}
}

func (s *StudyTimerService) UpdateParticipantStates(instanceID string, studyKey string) {
	rules, err := s.studyDBService.GetStudyRules(instanceID, studyKey)
	if err != nil {
		log.Printf("ERROR in UpdateParticipantStates.GetStudyRules (%s, %s): %v", instanceID, studyKey, err)
		return
	}

	studyEvent := types.StudyEvent{
		Type: "TIMER",
	}

	if !s.hasRuleForEventType(rules, studyEvent) {
		log.Printf("UpdateParticipantStates (%s, %s): has no timer related rules, skipped.", instanceID, studyKey)
		return
	}

	if err := s.studyDBService.FindAndExecuteOnParticipantsStates(instanceID, studyKey, s.getAndUpdateParticipantState, rules, studyEvent); err != nil {
		log.Printf("ERROR in UpdateParticipantStates.FindAndExecuteOnParticipantsStates (%s, %s): %v", instanceID, studyKey, err)
	}
}

func (s *StudyTimerService) getAndUpdateParticipantState(
	studyDBServ *studydb.StudyDBService,
	pState types.ParticipantState,
	instanceID string,
	studyKey string,
	args ...interface{},
) (err error) {
	if len(args) != 2 {
		err = errors.New("unexpected number of args")
		log.Printf("ERROR in getAndUpdateParticipantState: %v", err)
		return
	}
	rules := args[0].([]types.Expression)
	studyEvent := args[1].(types.StudyEvent)

	for _, rule := range rules {
		pState, err = studyengine.ActionEval(rule, pState, studyEvent)
		if err != nil {
			log.Printf("ERROR in getAndUpdateParticipantState.ActionEval (%s, %s): %v", instanceID, studyKey, err)
			continue
		}
	}
	// save state back to DB
	_, err = studyDBServ.SaveParticipantState(instanceID, studyKey, pState)
	return err
}

func (s *StudyTimerService) hasRuleForEventType(rules []types.Expression, event types.StudyEvent) bool {
	for _, rule := range rules {
		if len(rule.Data) < 1 {
			continue
		}
		exp := rule.Data[0].Exp
		if exp == nil || len(exp.Data) < 1 || exp.Data[0].Str != event.Type {
			continue
		}
		return true
	}
	return false
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
