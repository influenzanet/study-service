package main

import (
	"testing"
	"time"

	"github.com/influenzanet/study-service/models"
)

func TestDbAddSurveyAndContextDef(t *testing.T) {
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

func TestDbFindSurveyDefinition(t *testing.T) {
	testSurvey := models.Survey{
		Current: models.SurveyVersion{
			Published: time.Now().Unix(),
			SurveyDefinition: models.SurveyItem{
				Key: "s1",
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

	t.Run("not existing survey", func(t *testing.T) {
		_, err := findSurveyDefDB(testInstanceID, studyKey, "wrong")
		if err == nil {
			t.Error("should return error")
		}
	})

	t.Run("existing survey", func(t *testing.T) {
		survey, err := findSurveyDefDB(testInstanceID, studyKey, testSurvey.Current.SurveyDefinition.Key)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if survey.Current.SurveyDefinition.Key != testSurvey.Current.SurveyDefinition.Key {
			t.Errorf("unexpected survey key: %s, %s (want, have)", testSurvey.Current.SurveyDefinition.Key, survey.Current.SurveyDefinition.Key)
		}
	})
}
