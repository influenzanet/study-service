package studydb

import (
	"testing"
	"time"

	"github.com/influenzanet/study-service/pkg/types"
)

func TestDbSaveSurveyAndContextDef(t *testing.T) {
	t.Run("Testing create survey", func(t *testing.T) {
		testSurvey := types.Survey{
			Published: time.Now().Unix(),
			SurveyDefinition: types.SurveyItem{
				Key: "ST",
				Items: []types.SurveyItem{
					{
						Key:     "Q1",
						Follows: []string{"ST"},
						Condition: &types.Expression{
							Name: "testmethod",
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
	testSurvey := types.Survey{
		Published: time.Now().Unix(),
		VersionID: "version1",
		SurveyDefinition: types.SurveyItem{
			Key: "s1",
			Items: []types.SurveyItem{
				{
					Key:     "Q1",
					Follows: []string{"ST"},
					Condition: &types.Expression{
						Name: "testmethod",
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
		err := testDBService.DeleteSurveyVersion(testInstanceID, studyKey, "wrong", "")
		if err == nil {
			t.Error("should return error")
		}
	})

	t.Run("with not existing versionID", func(t *testing.T) {
		err := testDBService.DeleteSurveyVersion(testInstanceID, studyKey, "s1", "wrong")
		if err == nil {
			t.Error("should return error")
		}
	})

	t.Run("Test removing survey definition from study", func(t *testing.T) {
		err := testDBService.DeleteSurveyVersion(testInstanceID, studyKey, testSurvey.SurveyDefinition.Key, testSurvey.VersionID)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		surveys, err := testDBService.FindSurveyDefHistory(testInstanceID, studyKey, "", true)
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
	testSurveys := []types.Survey{
		{
			Published: 10,
			SurveyDefinition: types.SurveyItem{
				Key: "s1",
				Items: []types.SurveyItem{
					{
						Key:     "Q1",
						Follows: []string{"ST"},
						Condition: &types.Expression{
							Name: "testmethod",
						},
					},
				},
			},
		},
		{
			Published: 5,
			SurveyDefinition: types.SurveyItem{
				Key: "s1",
				Items: []types.SurveyItem{
					{
						Key:     "Q1",
						Follows: []string{"ST"},
						Condition: &types.Expression{
							Name: "testmethod",
						},
					},
				},
			},
		},
	}

	studyKey := "test-study-key"
	for _, s := range testSurveys {
		_, err := testDBService.SaveSurvey(testInstanceID, studyKey, s)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
	}

	t.Run("not existing survey", func(t *testing.T) {
		_, err := testDBService.FindCurrentSurveyDef(testInstanceID, studyKey, "wrong", false)
		if err == nil {
			t.Error("should return error")
		}
	})

	t.Run("existing survey", func(t *testing.T) {
		survey, err := testDBService.FindCurrentSurveyDef(testInstanceID, studyKey, testSurveys[0].SurveyDefinition.Key, false)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if survey.SurveyDefinition.Key != testSurveys[0].SurveyDefinition.Key {
			t.Errorf("unexpected survey key: %s, %s (want, have)", testSurveys[0].SurveyDefinition.Key, survey.SurveyDefinition.Key)
		}
		if survey.Published != testSurveys[0].Published {
			t.Errorf("unexpected published date: %d, %d (want, have)", testSurveys[0].Published, survey.Published)
		}
	})
}

func TestDbFindAllSurveyDefinitions(t *testing.T) {
	testSurvey := types.Survey{
		Published: time.Now().Unix(),
		SurveyDefinition: types.SurveyItem{
			Key: "s1",
			Items: []types.SurveyItem{
				{
					Key:     "Q1",
					Follows: []string{"ST"},
					Condition: &types.Expression{
						Name: "testmethod",
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
	testSurvey.SurveyDefinition.Key = "erw"
	_, err = testDBService.SaveSurvey(testInstanceID, studyKey, testSurvey)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}

	t.Run("not existing study", func(t *testing.T) {
		surveys, err := testDBService.FindSurveyDefHistory(testInstanceID, "wrong", "wrong", true)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(surveys) > 0 {
			t.Errorf("unexpected number of surveys: %d", len(surveys))
		}
	})

	t.Run("existing study with surveys", func(t *testing.T) {
		surveys, err := testDBService.FindSurveyDefHistory(testInstanceID, studyKey, "", true)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(surveys) != 2 {
			t.Errorf("unexpected number of surveys: %d", len(surveys))
		}
	})

	t.Run("existing study with survey key", func(t *testing.T) {
		surveys, err := testDBService.FindSurveyDefHistory(testInstanceID, studyKey, testSurvey.SurveyDefinition.Key, true)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(surveys) != 1 {
			t.Errorf("unexpected number of surveys: %d", len(surveys))
		}
	})
}

func TestDbFindAllCurrentSurveyDefinitions(t *testing.T) {
	testSurveys := []types.Survey{
		{
			Published: 10,
			SurveyDefinition: types.SurveyItem{
				Key: "s1",
				Items: []types.SurveyItem{
					{
						Key:     "Q1",
						Follows: []string{"ST"},
						Condition: &types.Expression{
							Name: "testmethod",
						},
					},
				},
			},
		},
		{
			Published: 5,
			SurveyDefinition: types.SurveyItem{
				Key: "s1",
				Items: []types.SurveyItem{
					{
						Key:     "Q1",
						Follows: []string{"ST"},
						Condition: &types.Expression{
							Name: "testmethod",
						},
					},
				},
			},
		},
		{
			Published:   12,
			Unpublished: 9,
			SurveyDefinition: types.SurveyItem{
				Key: "s1",
				Items: []types.SurveyItem{
					{
						Key:     "Q1",
						Follows: []string{"ST"},
						Condition: &types.Expression{
							Name: "testmethod",
						},
					},
				},
			},
		},
		{
			Published: 6,
			SurveyDefinition: types.SurveyItem{
				Key: "s2",
				Items: []types.SurveyItem{
					{
						Key:     "Q1",
						Follows: []string{"ST"},
						Condition: &types.Expression{
							Name: "testmethod",
						},
					},
				},
			},
		},
		{
			Published:   7,
			Unpublished: 8,
			SurveyDefinition: types.SurveyItem{
				Key: "s3",
				Items: []types.SurveyItem{
					{
						Key:     "Q1",
						Follows: []string{"ST"},
						Condition: &types.Expression{
							Name: "testmethod",
						},
					},
				},
			},
		},
	}

	studyKey := "test-study-find-allcurrent-key"
	for _, s := range testSurveys {
		_, err := testDBService.SaveSurvey(testInstanceID, studyKey, s)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
	}

	surveysWithoutUnpublished, err := testDBService.GetSurveyKeysInStudy(testInstanceID, studyKey, false)
	t.Run("has error", func(t *testing.T) {
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("has 2 results", func(t *testing.T) {
		if len(surveysWithoutUnpublished) != 2 {
			t.Errorf("unexpect number of values returned: %d", len(surveysWithoutUnpublished))
		}
	})

	surveysWithUnpublished, err := testDBService.GetSurveyKeysInStudy(testInstanceID, studyKey, true)
	t.Run("has error", func(t *testing.T) {
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("has 3 results", func(t *testing.T) {
		if len(surveysWithUnpublished) != 3 {
			t.Errorf("unexpect number of values returned: %d", len(surveysWithoutUnpublished))
		}
	})

}
