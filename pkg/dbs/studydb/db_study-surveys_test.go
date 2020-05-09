package studydb

import (
	"testing"
	"time"

	"github.com/influenzanet/study-service/pkg/models"
)

func TestDbSaveSurveyAndContextDef(t *testing.T) {
	t.Run("Testing create survey", func(t *testing.T) {
		testSurvey := models.Survey{
			Current: models.SurveyVersion{
				Published: time.Now().Unix(),
				SurveyDefinition: models.SurveyItem{
					Key: "ST",
					Items: []models.SurveyItem{
						{
							Key:     "Q1",
							Follows: []string{"ST"},
							Condition: &models.Expression{
								Name: "testmethod",
							},
						},
					},
				},
			},
		}

		studyKey := "test-study-key"
		_, err := testDBService.SaveSurvey(testInstanceID, studyKey, testSurvey)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
	})
}

func TestDbRemoveSurveyFromStudy(t *testing.T) {
	testSurvey := models.Survey{
		Current: models.SurveyVersion{
			Published: time.Now().Unix(),
			SurveyDefinition: models.SurveyItem{
				Key: "s1",
				Items: []models.SurveyItem{
					{
						Key:     "Q1",
						Follows: []string{"ST"},
						Condition: &models.Expression{
							Name: "testmethod",
						},
					},
				},
			},
		},
	}

	studyKey := "test-for-db-removing-survey"
	_, err := testDBService.SaveSurvey(testInstanceID, studyKey, testSurvey)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}

	t.Run("with not existing key", func(t *testing.T) {
		err := testDBService.RemoveSurveyFromStudy(testInstanceID, studyKey, "wrong")
		if err == nil {
			t.Error("should return error")
		}
	})

	t.Run("Test removing survey definition from study", func(t *testing.T) {
		err := testDBService.RemoveSurveyFromStudy(testInstanceID, studyKey, testSurvey.Current.SurveyDefinition.Key)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		surveys, err := testDBService.FindAllSurveyDefsForStudy(testInstanceID, studyKey, true)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(surveys) > 0 {
			t.Errorf("unexpected number of surveys: %d", len(surveys))
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
					{
						Key:     "Q1",
						Follows: []string{"ST"},
						Condition: &models.Expression{
							Name: "testmethod",
						},
					},
				},
			},
		},
	}

	studyKey := "test-study-key"
	_, err := testDBService.SaveSurvey(testInstanceID, studyKey, testSurvey)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}

	t.Run("not existing survey", func(t *testing.T) {
		_, err := testDBService.FindSurveyDef(testInstanceID, studyKey, "wrong")
		if err == nil {
			t.Error("should return error")
		}
	})

	t.Run("existing survey", func(t *testing.T) {
		survey, err := testDBService.FindSurveyDef(testInstanceID, studyKey, testSurvey.Current.SurveyDefinition.Key)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if survey.Current.SurveyDefinition.Key != testSurvey.Current.SurveyDefinition.Key {
			t.Errorf("unexpected survey key: %s, %s (want, have)", testSurvey.Current.SurveyDefinition.Key, survey.Current.SurveyDefinition.Key)
		}
	})
}

func TestDbFindAllSurveyDefinitions(t *testing.T) {
	testSurvey := models.Survey{
		Current: models.SurveyVersion{
			Published: time.Now().Unix(),
			SurveyDefinition: models.SurveyItem{
				Key: "s1",
				Items: []models.SurveyItem{
					{
						Key:     "Q1",
						Follows: []string{"ST"},
						Condition: &models.Expression{
							Name: "testmethod",
						},
					},
				},
			},
		},
	}

	studyKey := "test-study-key-for-find-all-surveys"
	_, err := testDBService.SaveSurvey(testInstanceID, studyKey, testSurvey)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	testSurvey.Current.SurveyDefinition.Key = "erw"
	_, err = testDBService.SaveSurvey(testInstanceID, studyKey, testSurvey)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}

	t.Run("not existing study", func(t *testing.T) {
		surveys, err := testDBService.FindAllSurveyDefsForStudy(testInstanceID, "wrong", true)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(surveys) > 0 {
			t.Errorf("unexpected number of surveys: %d", len(surveys))
		}
	})

	t.Run("existing study with surveys", func(t *testing.T) {
		surveys, err := testDBService.FindAllSurveyDefsForStudy(testInstanceID, studyKey, true)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(surveys) != 2 {
			t.Errorf("unexpected number of surveys: %d", len(surveys))
		}
	})
}
