package main

import (
	"testing"
	"time"

	"github.com/influenzanet/study-service/models"
)

func TestDbCreateSurvey(t *testing.T) {
	t.Run("Testing create survey", func(t *testing.T) {
		testSurvey := models.Survey{
			Current: models.SurveyVersion{
				Published: time.Now().Unix(),
				SurveyDefinition: models.SurveyItem{
					Key: "ST",
					Items: []models.SurveyItem{
						models.SurveyItem{
							Key:     "Q1",
							Follows: []string{"ST"},
							Condition: models.Expression{
								Name: "testmethod",
							},
						},
					},
				},
			},
		}

		studyKey := "test-study-key"
		_, err := addSurveyToDB(testInstanceID, studyKey, testSurvey)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
	})
}
