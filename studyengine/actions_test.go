package studyengine

import (
	"testing"
	"time"

	"github.com/influenzanet/study-service/models"
)

func TestActions(t *testing.T) {
	participantState := models.ParticipantState{
		ParticipantID: "participant1234",
		StudyStatus:   "active",
		Flags: map[string]string{
			"health": "test",
		},
	}
	event := models.StudyEvent{
		Type: "SUBMIT",
		Response: models.SurveyResponse{
			Key: "test",
		},
	}

	t.Run("with wrong action name", func(t *testing.T) {
		action := models.Expression{
			Name: "WRONG",
		}
		_, err := ActionEval(action, participantState, event)
		if err == nil {
			t.Error("should return an error")
		}
	})

	t.Run("IFTHEN", func(t *testing.T) {
		action2 := models.Expression{
			Name: "UPDATE_STUDY_STATUS",
			Data: []models.ExpressionArg{
				models.ExpressionArg{DType: "str", Str: "testflag_cond"},
			},
		}
		action3 := models.Expression{
			Name: "UPDATE_STUDY_STATUS",
			Data: []models.ExpressionArg{
				models.ExpressionArg{DType: "str", Str: "testflag_cond2"},
			},
		}
		action := models.Expression{
			Name: "IFTHEN",
			Data: []models.ExpressionArg{
				models.ExpressionArg{DType: "num", Num: 0},
				models.ExpressionArg{DType: "exp", Exp: action2},
			},
		}
		newState, err := ActionEval(action, participantState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if newState.StudyStatus == action2.Data[0].Str {
			t.Errorf("error -> expected: %s, have: %s", action.Data[1].Str, newState.StudyStatus)
		}

		action = models.Expression{
			Name: "IFTHEN",
			Data: []models.ExpressionArg{
				models.ExpressionArg{DType: "num", Num: 1},
				models.ExpressionArg{DType: "exp", Exp: action2},
				models.ExpressionArg{DType: "exp", Exp: action3},
			},
		}
		newState, err = ActionEval(action, participantState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if newState.StudyStatus != action3.Data[0].Str {
			t.Errorf("error -> expected: %s, have: %s", action3.Data[1].Str, newState.StudyStatus)
		}
	})

	t.Run("UPDATE_FLAG", func(t *testing.T) {
		action := models.Expression{
			Name: "UPDATE_FLAG",
			Data: []models.ExpressionArg{
				models.ExpressionArg{DType: "str", Str: "key"},
				models.ExpressionArg{DType: "str", Str: "value"},
			},
		}
		newState, err := ActionEval(action, participantState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		v, ok := newState.Flags["key"]
		if !ok {
			t.Error("could not find new flag")
			return
		}
		if v != action.Data[1].Str {
			t.Errorf("updated status error -> expected: %s, have: %s", action.Data[1].Str, v)
		}
	})

	t.Run("REMOVE_FLAG", func(t *testing.T) {
		action := models.Expression{
			Name: "REMOVE_FLAG",
			Data: []models.ExpressionArg{
				models.ExpressionArg{DType: "str", Str: "health"},
			},
		}
		newState, err := ActionEval(action, participantState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		_, ok := newState.Flags["health"]
		if ok {
			t.Error("should not find value")
			return
		}
	})

	// Survey actions:
	t.Run("ADD_NEW_SURVEY", func(t *testing.T) {
		now := time.Now().Unix()
		action := models.Expression{
			Name: "ADD_NEW_SURVEY",
			Data: []models.ExpressionArg{
				models.ExpressionArg{DType: "str", Str: "testSurveyKey"},
				models.ExpressionArg{DType: "num", Num: float64(now - 10)},
				models.ExpressionArg{DType: "num", Num: float64(now + 10)},
			},
		}
		newState, err := ActionEval(action, participantState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(newState.AssignedSurveys) != 1 {
			t.Errorf("updated number of surveys: %d", len(newState.AssignedSurveys))
			return
		}
		if newState.AssignedSurveys[0].ValidFrom != now-10 {
			t.Errorf("unexpected validFrom: have %d, exprected: %d", newState.AssignedSurveys[0].ValidFrom, now-10)
		}
		if newState.AssignedSurveys[0].ValidUntil != now+10 {
			t.Errorf("unexpected validFrom: have %d, exprected: %d", newState.AssignedSurveys[0].ValidUntil, now+10)
		}
	})

	t.Run("REMOVE_ALL_SURVEYS", func(t *testing.T) {
		// Add surveys first
		now := time.Now().Unix()
		action := models.Expression{
			Name: "ADD_NEW_SURVEY",
			Data: []models.ExpressionArg{
				models.ExpressionArg{DType: "str", Str: "testSurveyKey"},
				models.ExpressionArg{DType: "num", Num: float64(now - 10)},
				models.ExpressionArg{DType: "num", Num: float64(now + 10)},
			},
		}
		newState, err := ActionEval(action, participantState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		newState, err = ActionEval(action, newState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(newState.AssignedSurveys) != 2 {
			t.Errorf("unexpected number of surveys: %d", len(newState.AssignedSurveys))
			return
		}

		// REMOVE_ALL_SURVEYS
		action = models.Expression{
			Name: "REMOVE_ALL_SURVEYS",
		}
		newState, err = ActionEval(action, newState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}

		if len(newState.AssignedSurveys) > 0 {
			t.Error("should not have surveys any more")
		}
	})

	t.Run("REMOVE_SURVEY_BY_KEY last", func(t *testing.T) {
		// Add surveys first
		now := time.Now().Unix()
		action := models.Expression{
			Name: "ADD_NEW_SURVEY",
			Data: []models.ExpressionArg{
				models.ExpressionArg{DType: "str", Str: "testSurveyKey1"},
				models.ExpressionArg{DType: "num", Num: float64(now - 10)},
				models.ExpressionArg{DType: "num", Num: float64(now + 10)},
			},
		}
		newState, err := ActionEval(action, participantState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		action.Data[0].Str = "testSurveyKey2"
		newState, err = ActionEval(action, newState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		action.Data[0].Str = "testSurveyKey1"
		newState, err = ActionEval(action, newState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(newState.AssignedSurveys) != 3 {
			t.Errorf("unexpected number of surveys: %d", len(newState.AssignedSurveys))
			return
		}

		// REMOVE_SURVEY_BY_KEY
		action = models.Expression{
			Name: "REMOVE_SURVEY_BY_KEY",
			Data: []models.ExpressionArg{
				models.ExpressionArg{DType: "str", Str: "testSurveyKey1"},
				models.ExpressionArg{DType: "str", Str: "last"},
			},
		}
		newState, err = ActionEval(action, newState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}

		if len(newState.AssignedSurveys) != 2 {
			t.Errorf("unexpected number of surveys: %d", len(newState.AssignedSurveys))
			return
		}
		if newState.AssignedSurveys[0].SurveyKey != "testSurveyKey1" {
			t.Errorf("unexpected survey key at pos 0: %s", newState.AssignedSurveys[0].SurveyKey)
		}
	})

	t.Run("REMOVE_SURVEY_BY_KEY first", func(t *testing.T) {
		// Add surveys first
		now := time.Now().Unix()
		action := models.Expression{
			Name: "ADD_NEW_SURVEY",
			Data: []models.ExpressionArg{
				models.ExpressionArg{DType: "str", Str: "testSurveyKey1"},
				models.ExpressionArg{DType: "num", Num: float64(now - 10)},
				models.ExpressionArg{DType: "num", Num: float64(now + 10)},
			},
		}
		newState, err := ActionEval(action, participantState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		action.Data[0].Str = "testSurveyKey2"
		newState, err = ActionEval(action, newState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		action.Data[0].Str = "testSurveyKey1"
		newState, err = ActionEval(action, newState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(newState.AssignedSurveys) != 3 {
			t.Errorf("unexpected number of surveys: %d", len(newState.AssignedSurveys))
			return
		}

		// REMOVE_SURVEY_BY_KEY
		action = models.Expression{
			Name: "REMOVE_SURVEY_BY_KEY",
			Data: []models.ExpressionArg{
				models.ExpressionArg{DType: "str", Str: "testSurveyKey1"},
				models.ExpressionArg{DType: "str", Str: "first"},
			},
		}
		newState, err = ActionEval(action, newState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}

		if len(newState.AssignedSurveys) != 2 {
			t.Errorf("unexpected number of surveys: %d", len(newState.AssignedSurveys))
		}
		if newState.AssignedSurveys[0].SurveyKey != "testSurveyKey2" {
			t.Errorf("unexpected survey key at pos 0: %s", newState.AssignedSurveys[0].SurveyKey)
		}
	})

	t.Run("REMOVE_SURVEYS_BY_KEY", func(t *testing.T) {
		// Add surveys first
		now := time.Now().Unix()
		action := models.Expression{
			Name: "ADD_NEW_SURVEY",
			Data: []models.ExpressionArg{
				models.ExpressionArg{DType: "str", Str: "testSurveyKey1"},
				models.ExpressionArg{DType: "num", Num: float64(now - 10)},
				models.ExpressionArg{DType: "num", Num: float64(now + 10)},
			},
		}
		newState, err := ActionEval(action, participantState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		newState, err = ActionEval(action, newState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		action.Data[0].Str = "testSurveyKey2"
		newState, err = ActionEval(action, newState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(newState.AssignedSurveys) != 3 {
			t.Errorf("unexpected number of surveys: %d", len(newState.AssignedSurveys))
			return
		}

		// REMOVE_SURVEYS_BY_KEY
		action = models.Expression{
			Name: "REMOVE_SURVEYS_BY_KEY",
			Data: []models.ExpressionArg{
				models.ExpressionArg{DType: "str", Str: "testSurveyKey1"},
			},
		}
		newState, err = ActionEval(action, newState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}

		if len(newState.AssignedSurveys) != 1 {
			t.Errorf("unexpected number of surveys: %d", len(newState.AssignedSurveys))
		}
	})

	// Report actions
	participantState = models.ParticipantState{
		ParticipantID: "participant1234",
		StudyStatus:   "active",
		Flags: map[string]string{
			"health": "test",
		},
		Reports: []models.SurveyItemResponse{
			models.SurveyItemResponse{Key: "test.1"},
			models.SurveyItemResponse{Key: "test.2.1"},
			models.SurveyItemResponse{Key: "test.1"},
		},
	}
	event = models.StudyEvent{
		Type: "SUBMIT",
		Response: models.SurveyResponse{
			Key: "test",
			Responses: []models.SurveyItemResponse{
				models.SurveyItemResponse{Key: "test.1"},
				models.SurveyItemResponse{Key: "test.2.1"},
				models.SurveyItemResponse{Key: "test.2.3"},
			},
		},
	}
	t.Run("ADD_REPORT not existing key", func(t *testing.T) {
		action := models.Expression{
			Name: "ADD_REPORT",
			Data: []models.ExpressionArg{
				models.ExpressionArg{DType: "str", Str: "test.2.2"},
			},
		}
		newState, err := ActionEval(action, participantState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(newState.Reports) != 3 {
			t.Errorf("unexpected number of reports: %d", len(newState.Reports))
		}
	})

	t.Run("ADD_REPORT", func(t *testing.T) {
		action := models.Expression{
			Name: "ADD_REPORT",
			Data: []models.ExpressionArg{
				models.ExpressionArg{DType: "str", Str: "test.2.1"},
			},
		}
		newState, err := ActionEval(action, participantState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(newState.Reports) != 4 {
			t.Errorf("unexpected number of reports: %d", len(newState.Reports))
		}
	})

	t.Run("REMOVE_ALL_REPORTS", func(t *testing.T) {
		action := models.Expression{
			Name: "REMOVE_ALL_REPORTS",
		}
		newState, err := ActionEval(action, participantState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(newState.Reports) > 0 {
			t.Errorf("unexpected number of reports: %d", len(newState.Reports))
		}
	})

	t.Run("REMOVE_REPORT_BY_KEY first", func(t *testing.T) {
		action := models.Expression{
			Name: "REMOVE_REPORT_BY_KEY",
			Data: []models.ExpressionArg{
				models.ExpressionArg{DType: "str", Str: "test.1"},
				models.ExpressionArg{DType: "str", Str: "first"},
			},
		}
		newState, err := ActionEval(action, participantState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(newState.Reports) != 2 {
			t.Errorf("unexpected number of reports: %d", len(newState.Reports))
			return
		}
		if newState.Reports[0].Key == "test.1" {
			t.Errorf("unexpected first item key: %s", newState.Reports[0].Key)
		}
	})

	t.Run("REMOVE_REPORT_BY_KEY last", func(t *testing.T) {
		action := models.Expression{
			Name: "REMOVE_REPORT_BY_KEY",
			Data: []models.ExpressionArg{
				models.ExpressionArg{DType: "str", Str: "test.1"},
				models.ExpressionArg{DType: "str", Str: "last"},
			},
		}
		newState, err := ActionEval(action, participantState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(newState.Reports) != 2 {
			t.Errorf("unexpected number of reports: %d", len(newState.Reports))
			return
		}
		if newState.Reports[0].Key != "test.1" {
			t.Errorf("unexpected first item key: %s", newState.Reports[0].Key)
		}
	})

	t.Run("REMOVE_REPORTS_BY_KEY", func(t *testing.T) {
		action := models.Expression{
			Name: "REMOVE_REPORTS_BY_KEY",
			Data: []models.ExpressionArg{
				models.ExpressionArg{DType: "str", Str: "test.1"},
			},
		}
		newState, err := ActionEval(action, participantState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(newState.Reports) != 1 {
			t.Errorf("unexpected number of reports: %d", len(newState.Reports))
			return
		}
		if newState.Reports[0].Key != "test.2.1" {
			t.Errorf("unexpected first item key: %s", newState.Reports[0].Key)
		}
	})
}
