package main

import (
	"errors"
	"fmt"

	"github.com/influenzanet/study-service/models"
	"github.com/influenzanet/study-service/studyengine"
	"github.com/influenzanet/study-service/utils"
)

func userIDToParticipantID(instanceID string, studyKey string, userID string) (string, error) {
	studySecret, err := getStudySecretKey(instanceID, studyKey)
	if err != nil {
		return "", err
	}
	return utils.UserIDtoParticipantID(userID, conf.Study.GlobalSecret, studySecret)
}

func checkIfParticipantExists(instanceID string, studyKey string, participantID string) bool {
	_, err := findParticipantStateDB(instanceID, studyKey, participantID)
	return err == nil
}

func getAndPerformStudyRules(instanceID string, studyKey string, pState models.ParticipantState, event models.StudyEvent) (newState models.ParticipantState, err error) {
	newState = pState
	rules, err := getStudyRules(instanceID, studyKey)
	if err != nil {
		return
	}
	for _, rule := range rules {
		newState, err = studyengine.ActionEval(rule, newState, event)
		if err != nil {
			return
		}
	}

	return newState, nil
}

func resolveContextRules(instanceID string, studyKey string, participantID string, rules models.SurveyContextDef) (sCtx models.SurveyContext, err error) {
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
			cResps, _ := findSurveyResponsesInDB(instanceID, studyKey, responseQuery{
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
			cResps, _ := findSurveyResponsesInDB(instanceID, studyKey, responseQuery{
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
			cResps, _ := findSurveyResponsesInDB(instanceID, studyKey, responseQuery{
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

func resolvePrefillRules(instanceID string, studyKey string, participantID string, rules []models.Expression) (prefills models.SurveyResponse, err error) {
	for _, rule := range rules {
		switch rule.Name {
		case "GET_LAST_SURVEY_ITEM":
			if len(rule.Data) != 2 {
				return prefills, errors.New("GET_LAST_SURVEY_ITEM must have two arguments")
			}
			surveyKey := rule.Data[0].Str
			itemKey := rule.Data[1].Str
			resps, err := findSurveyResponsesInDB(instanceID, studyKey, responseQuery{
				ParticipantID: participantID,
				SurveyKey:     surveyKey,
				Limit:         1,
			})

			if err != nil || len(resps) < 1 {
				continue
			}
			for _, item := range resps[0].Responses {
				if item.Key == itemKey {
					prefills.Responses = append(prefills.Responses, item)
					continue
				}
			}
		default:
			return prefills, fmt.Errorf("expression is not supported yet: %s", rule.Name)
		}
	}
	return prefills, nil
}
