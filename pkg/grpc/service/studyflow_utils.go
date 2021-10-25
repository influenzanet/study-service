package service

import (
	"errors"
	"fmt"

	"github.com/influenzanet/study-service/pkg/dbs/studydb"
	"github.com/influenzanet/study-service/pkg/studyengine"
	"github.com/influenzanet/study-service/pkg/types"
	"github.com/influenzanet/study-service/pkg/utils"
)

func (s *studyServiceServer) profileIDToParticipantID(instanceID string, studyKey string, userID string) (string, error) {
	idMappingMethod, studySecret, err := s.studyDBservice.GetStudySecretKey(instanceID, studyKey)
	if err != nil {
		return "", err
	}
	return utils.ProfileIDtoParticipantID(userID, s.StudyGlobalSecret, studySecret, idMappingMethod)
}

func (s *studyServiceServer) checkIfParticipantExists(instanceID string, studyKey string, participantID string, withStatus string) bool {
	pState, err := s.studyDBservice.FindParticipantState(instanceID, studyKey, participantID)
	if err != nil || (withStatus != "" && pState.StudyStatus != withStatus) {
		return false
	}
	return err == nil
}

func (s *studyServiceServer) getAndPerformStudyRules(instanceID string, studyKey string, pState types.ParticipantState, event types.StudyEvent) (newState types.ParticipantState, err error) {
	newState = pState
	rules, err := s.studyDBservice.GetStudyRules(instanceID, studyKey)
	if err != nil {
		return
	}
	for _, rule := range rules {
		newState, err = studyengine.ActionEval(rule, newState, event, s.studyDBservice)
		if err != nil {
			return
		}
	}

	return newState, nil
}

func (s *studyServiceServer) resolveContextRules(instanceID string, studyKey string, participantID string, rules *types.SurveyContextDef) (sCtx types.SurveyContext, err error) {
	pState, err := s.studyDBservice.FindParticipantState(instanceID, studyKey, participantID)
	if err != nil {
		return sCtx, errors.New("no participant with this id in this study")
	}
	// participant flags:
	sCtx.ParticipantFlags = pState.Flags

	if rules == nil {
		return sCtx, nil
	}

	// mode:
	if rules.Mode != nil {
		modeRule := rules.Mode
		switch modeRule.DType {
		case "exp":
			return sCtx, errors.New("expression arg type not supported yet")
		case "str":
			sCtx.Mode = modeRule.Str
		default:
			sCtx.Mode = modeRule.Str
		}
	}

	// previous responses:
	prevRespRules := rules.PreviousResponses
	for _, rule := range prevRespRules {
		switch rule.Name {
		case "LAST_RESPONSES_BY_KEY":
			if len(rule.Data) != 2 {
				return sCtx, errors.New("LAST_RESPONSES_BY_KEY must have two arguments")
			}
			arg1 := rule.Data[0].Str
			arg2 := int64(rule.Data[1].Num)
			if arg1 == "" || arg2 == 0 {
				return sCtx, errors.New("LAST_RESPONSES_BY_KEY arguments have to be defined")
			}
			cResps, _ := s.studyDBservice.FindSurveyResponses(instanceID, studyKey, studydb.ResponseQuery{
				ParticipantID: participantID,
				SurveyKey:     arg1,
				Limit:         arg2,
			})
			sCtx.PreviousResponses = append(sCtx.PreviousResponses, cResps...)
		case "ALL_RESPONSES_SINCE":
			if len(rule.Data) != 1 {
				return sCtx, errors.New("ALL_RESPONSES_SINCE must have one argument")
			}
			arg1 := int64(rule.Data[0].Num)
			cResps, _ := s.studyDBservice.FindSurveyResponses(instanceID, studyKey, studydb.ResponseQuery{
				ParticipantID: participantID,
				Since:         arg1,
			})
			sCtx.PreviousResponses = append(sCtx.PreviousResponses, cResps...)
		case "RESPONSES_SINCE_BY_KEY":
			if len(rule.Data) != 2 {
				return sCtx, errors.New("RESPONSES_SINCE_BY_KEY must have two arguments")
			}
			since := int64(rule.Data[0].Num)
			surveyKey := rule.Data[1].Str
			if surveyKey == "" || since == 0 {
				return sCtx, errors.New("RESPONSES_SINCE_BY_KEY arguments have to be defined")
			}
			cResps, _ := s.studyDBservice.FindSurveyResponses(instanceID, studyKey, studydb.ResponseQuery{
				ParticipantID: participantID,
				SurveyKey:     surveyKey,
				Since:         since,
			})
			sCtx.PreviousResponses = append(sCtx.PreviousResponses, cResps...)
		default:
			return sCtx, errors.New("expression is not supported yet")
		}
	}
	return sCtx, nil
}

func (s *studyServiceServer) resolvePrefillRules(instanceID string, studyKey string, participantID string, rules []types.Expression) (prefills types.SurveyResponse, err error) {
	lastSurveyCache := map[string]types.SurveyResponse{}
	for _, rule := range rules {
		switch rule.Name {
		case "GET_LAST_SURVEY_ITEM":
			if len(rule.Data) != 2 {
				return prefills, errors.New("GET_LAST_SURVEY_ITEM must have two arguments")
			}
			surveyKey := rule.Data[0].Str
			itemKey := rule.Data[1].Str

			previousResp, ok := lastSurveyCache[surveyKey]
			if !ok {
				resps, err := s.studyDBservice.FindSurveyResponses(instanceID, studyKey, studydb.ResponseQuery{
					ParticipantID: participantID,
					SurveyKey:     surveyKey,
					Limit:         1,
				})

				if err != nil || len(resps) < 1 {
					continue
				}
				lastSurveyCache[surveyKey] = resps[0]
				previousResp = resps[0]
			}

			for _, item := range previousResp.Responses {
				if item.Key == itemKey {
					prefills.Responses = append(prefills.Responses, item)
					break
				}
			}
		default:
			return prefills, fmt.Errorf("expression is not supported yet: %s", rule.Name)
		}
	}
	return prefills, nil
}
