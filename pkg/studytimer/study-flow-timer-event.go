package studytimer

import (
	"context"
	"errors"

	"github.com/coneno/logger"
	"github.com/influenzanet/study-service/pkg/dbs/studydb"
	"github.com/influenzanet/study-service/pkg/studyengine"
	"github.com/influenzanet/study-service/pkg/types"
	"github.com/influenzanet/study-service/pkg/utils"
)

func (s *StudyTimerService) StudyTimerEvent() {
	instances, err := s.globalDBService.GetAllInstances()
	if err != nil {
		logger.Error.Printf("unexpected error: %s", err.Error())
	}
	for _, instance := range instances {
		studies, err := s.studyDBService.GetStudiesByStatus(instance.InstanceID, types.STUDY_STATUS_ACTIVE, true)
		if err != nil {
			logger.Error.Printf("unexpected error: %s", err.Error())
			return
		}
		for _, study := range studies {
			if err := s.studyDBService.ShouldPerformTimerEvent(instance.InstanceID, study.Key, s.TimerEventFrequency); err != nil {
				continue
			}
			logger.Info.Printf("performing timer event for study: %s - %s", instance.InstanceID, study.Key)

			s.UpdateStudyStats(instance.InstanceID, study.Key)
			s.UpdateParticipantStates(instance.InstanceID, study)
		}
	}
}

func (s *StudyTimerService) UpdateParticipantStates(instanceID string, study types.Study) {
	rules, err := s.studyDBService.GetStudyRules(instanceID, study.Key)
	if err != nil {
		logger.Error.Printf("ERROR in UpdateParticipantStates.GetStudyRules (%s, %s): %v", instanceID, study.Key, err)
		return
	}

	studyEvent := types.StudyEvent{
		Type:       "TIMER",
		InstanceID: instanceID,
		StudyKey:   study.Key,
	}

	if !s.hasRuleForEventType(rules, studyEvent) {
		logger.Info.Printf("UpdateParticipantStates (%s, %s): has no timer related rules, skipped.", instanceID, study.Key)
		return
	}

	ctx := context.Background()
	if err := s.studyDBService.FindAndExecuteOnParticipantsStates(ctx, instanceID, study.Key, types.STUDY_STATUS_ACTIVE, s.getAndUpdateParticipantState, rules, studyEvent, study); err != nil {
		logger.Error.Printf("ERROR in UpdateParticipantStates.FindAndExecuteOnParticipantsStates (%s, %s): %v", instanceID, study.Key, err)
	}
}

func (s *StudyTimerService) getAndUpdateParticipantState(
	studyDBServ *studydb.StudyDBService,
	pState types.ParticipantState,
	instanceID string,
	studyKey string,
	args ...interface{},
) (err error) {
	if len(args) != 3 {
		err = errors.New("unexpected number of args")
		logger.Error.Printf("ERROR in getAndUpdateParticipantState: %v", err)
		return
	}
	rules := args[0].([]types.Expression)
	studyEvent := args[1].(types.StudyEvent)
	studyEvent.StudyKey = studyKey
	studyEvent.InstanceID = instanceID
	study := args[2].(types.Study)

	participantID2, err := utils.ProfileIDtoParticipantID(pState.ParticipantID, s.studyGlobalSecret, study.SecretKey, study.Configs.IdMappingMethod)
	if err != nil {
		logger.Error.Printf("unexpected error when computing confidential participant id: %v", err)
		return
	}
	studyEvent.ParticipantIDForConfidentialResponses = participantID2

	actionState := studyengine.ActionData{
		PState:          pState,
		ReportsToCreate: map[string]types.Report{},
	}

	for _, rule := range rules {
		actionState, err = studyengine.ActionEval(rule, actionState, studyEvent, studyengine.ActionConfigs{
			DBService:              s.studyDBService,
			ExternalServiceConfigs: s.studyEngineExternalServices,
		})
		if err != nil {
			logger.Error.Printf("ERROR in getAndUpdateParticipantState.ActionEval (%s, %s): %v", instanceID, studyKey, err)
			continue
		}
	}

	// save state back to DB
	_, err = studyDBServ.SaveParticipantState(instanceID, studyKey, actionState.PState)
	if err != nil {
		logger.Error.Printf("unexpected error when saving participant state: %v", err)
	}

	for _, report := range actionState.ReportsToCreate {
		report.ResponseID = "TIMER"
		err := studyDBServ.SaveReport(instanceID, studyKey, report)
		if err != nil {
			logger.Error.Printf("unexpected error while save report: %v", err)
		} else {
			logger.Debug.Printf("Report with key '%s' for participant %s saved.", report.Key, report.ParticipantID)
		}
	}
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
		logger.Error.Printf("DB ERROR for participant counting for study: %s -> %s", studyKey, err.Error())
	}
	tpCount, err := s.studyDBService.GetParticipantCountByStatus(instanceID, studyKey, types.PARTICIPANT_STUDY_STATUS_TEMPORARY)
	if err != nil {
		logger.Error.Printf("DB ERROR for participant counting for study: %s -> %s", studyKey, err.Error())
	}
	rCount, err := s.studyDBService.CountSurveyResponsesByKey(instanceID, studyKey, "", 0, 0)
	if err != nil {
		logger.Error.Printf("DB ERROR for response counting for study: %s -> %s", studyKey, err.Error())
	}

	if err := s.studyDBService.UpdateStudyStats(instanceID, studyKey, types.StudyStats{
		ParticipantCount:     pCount,
		TempParticipantCount: tpCount,
		ResponseCount:        rCount,
	}); err != nil {
		logger.Error.Printf("DB ERROR for updating stats for study: %s -> %s", studyKey, err.Error())
	}
}
