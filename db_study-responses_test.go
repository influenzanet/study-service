package main

import (
	"testing"

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
