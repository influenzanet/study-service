package main

import (
	"log"
	"testing"
	"time"

	"github.com/influenzanet/study-service/models"
)

func TestDbAddSurveyResponse(t *testing.T) {
	testStudy := "testosaveresponse"
	testResp := models.SurveyResponse{
		Key: "test",
		Context: map[string]string{
			"test": "test",
		},
		Responses: []models.SurveyItemResponse{
			models.SurveyItemResponse{
				Key: "testosaveresponse.2",
				Response: &models.ResponseItem{
					Key:   "a",
					Value: "testv",
				},
			},
		},
	}
	t.Run("saving response", func(t *testing.T) {
		err := addSurveyResponseToDB(testInstanceID, testStudy, testResp)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
	})
}

func TestDbFindSurveyResponseForParticipant(t *testing.T) {
	testStudyKey := "teststudy_for_finding_responses"

	surveyResps := []models.SurveyResponse{
		// mix participants and order for submittedAt
		models.SurveyResponse{Key: "s1", SubmittedFor: "u1", SubmittedAt: time.Now().Add(-30 * time.Hour * 24).Unix()},
		models.SurveyResponse{Key: "s2", SubmittedFor: "u1", SubmittedAt: time.Now().Add(-32 * time.Hour * 24).Unix()},
		models.SurveyResponse{Key: "s1", SubmittedFor: "u2", SubmittedAt: time.Now().Add(-29 * time.Hour * 24).Unix()},
		models.SurveyResponse{Key: "s1", SubmittedFor: "u1", SubmittedAt: time.Now().Add(-23 * time.Hour * 24).Unix()},
		models.SurveyResponse{Key: "s1", SubmittedFor: "u1", SubmittedAt: time.Now().Add(-6 * time.Hour * 24).Unix()},
		models.SurveyResponse{Key: "s2", SubmittedFor: "u1", SubmittedAt: time.Now().Add(-5 * time.Hour * 24).Unix()},
		models.SurveyResponse{Key: "s1", SubmittedFor: "u2", SubmittedAt: time.Now().Add(-6 * time.Hour * 24).Unix()},
		models.SurveyResponse{Key: "s2", SubmittedFor: "u2", SubmittedAt: time.Now().Add(-7 * time.Hour * 24).Unix()},
		models.SurveyResponse{Key: "s1", SubmittedFor: "u1", SubmittedAt: time.Now().Add(-15 * time.Hour * 24).Unix()},
		models.SurveyResponse{Key: "s2", SubmittedFor: "u1", SubmittedAt: time.Now().Add(-14 * time.Hour * 24).Unix()},
	}
	for _, sr := range surveyResps {
		err := addSurveyResponseToDB(testInstanceID, testStudyKey, sr)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
	}

	t.Run("not existing participant", func(t *testing.T) {
		q := responseQuery{
			ParticipantID: "u3",
		}
		responses, err := findSurveyResponsesInDB(testInstanceID, testStudyKey, q)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(responses) > 0 {
			t.Errorf("unexpected number of responses found: %d", len(responses))
		}
	})

	t.Run("find last 2 surveys with key", func(t *testing.T) {
		q := responseQuery{
			ParticipantID: "u1",
			SurveyKey:     "s1",
			Limit:         2,
		}
		responses, err := findSurveyResponsesInDB(testInstanceID, testStudyKey, q)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(responses) != 2 {
			t.Errorf("unexpected number of responses found: %d", len(responses))
			return
		}
		if responses[0].Key != responses[1].Key && responses[0].Key != q.SurveyKey {
			t.Errorf("unexpected survey key: %s", responses[0].Key)
		}
	})

	t.Run("find last 2 surveys without key", func(t *testing.T) {
		q := responseQuery{
			ParticipantID: "u1",
			Limit:         2,
		}
		responses, err := findSurveyResponsesInDB(testInstanceID, testStudyKey, q)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(responses) != 2 {
			t.Errorf("unexpected number of responses found: %d", len(responses))
			return
		}
		if responses[0].Key == responses[1].Key {
			t.Errorf("unexpected survey keys: %s, %s", responses[0].Key, responses[1].Key)
		}
	})

	t.Run("find surveys after timestamp with key", func(t *testing.T) {
		q := responseQuery{
			ParticipantID: "u1",
			SurveyKey:     "s2",
			Since:         time.Now().Add(-20 * time.Hour * 24).Unix(),
		}
		responses, err := findSurveyResponsesInDB(testInstanceID, testStudyKey, q)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(responses) != 2 {
			t.Errorf("unexpected number of responses found: %d", len(responses))
			for _, r := range responses {
				log.Println(r)
			}
			return
		}
	})

	t.Run("find surveys after timestamp without key", func(t *testing.T) {
		q := responseQuery{
			ParticipantID: "u1",
			Since:         time.Now().Add(-20 * time.Hour * 24).Unix(),
		}
		responses, err := findSurveyResponsesInDB(testInstanceID, testStudyKey, q)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(responses) != 4 {
			t.Errorf("unexpected number of responses found: %d", len(responses))
			return
		}
	})
}
