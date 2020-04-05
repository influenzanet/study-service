package main

import (
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
