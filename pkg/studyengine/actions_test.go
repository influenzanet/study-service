package studyengine

import (
	"strconv"
	"testing"
	"time"

	"github.com/influenzanet/study-service/pkg/types"
)

func TestActions(t *testing.T) {
	participantState := types.ParticipantState{
		ParticipantID: "participant1234",
		StudyStatus:   types.PARTICIPANT_STUDY_STATUS_ACTIVE,
		Flags: map[string]string{
			"health": "test",
		},
	}
	event := types.StudyEvent{
		Type: "SUBMIT",
		Response: types.SurveyResponse{
			Key: "test",
		},
	}

	t.Run("with wrong action name", func(t *testing.T) {
		action := types.Expression{
			Name: "WRONG",
		}
		_, err := ActionEval(action, participantState, event, nil)
		if err == nil {
			t.Error("should return an error")
		}
	})

	t.Run("IFTHEN", func(t *testing.T) {
		action2 := types.Expression{
			Name: "UPDATE_STUDY_STATUS",
			Data: []types.ExpressionArg{
				{DType: "str", Str: "testflag_cond"},
			},
		}
		action3 := types.Expression{
			Name: "UPDATE_STUDY_STATUS",
			Data: []types.ExpressionArg{
				{DType: "str", Str: "testflag_cond2"},
			},
		}
		action := types.Expression{
			Name: "IFTHEN",
			Data: []types.ExpressionArg{
				{DType: "num", Num: 0},
				{DType: "exp", Exp: &action2},
			},
		}
		newState, err := ActionEval(action, participantState, event, nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if newState.StudyStatus == action2.Data[0].Str {
			t.Errorf("error -> expected: %s, have: %s", action2.Data[0].Str, newState.StudyStatus)
		}

		action = types.Expression{
			Name: "IFTHEN",
			Data: []types.ExpressionArg{
				{DType: "num", Num: 1},
				{DType: "exp", Exp: &action2},
				{DType: "exp", Exp: &action3},
			},
		}
		newState, err = ActionEval(action, participantState, event, nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if newState.StudyStatus != action3.Data[0].Str {
			t.Errorf("error -> expected: %s, have: %s", action3.Data[0].Str, newState.StudyStatus)
		}
	})

	t.Run("IF", func(t *testing.T) {
		action2 := types.Expression{
			Name: "UPDATE_STUDY_STATUS",
			Data: []types.ExpressionArg{
				{DType: "str", Str: "testflag_cond"},
			},
		}
		action3 := types.Expression{
			Name: "UPDATE_STUDY_STATUS",
			Data: []types.ExpressionArg{
				{DType: "str", Str: "testflag_cond2"},
			},
		}
		action := types.Expression{
			Name: "IF",
			Data: []types.ExpressionArg{
				{DType: "num", Num: 0},
				{DType: "exp", Exp: &action2},
			},
		}
		newState, err := ActionEval(action, participantState, event, nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if newState.StudyStatus != types.PARTICIPANT_STUDY_STATUS_ACTIVE {
			t.Errorf("error -> expected: %s, have: %s", types.PARTICIPANT_STUDY_STATUS_ACTIVE, newState.StudyStatus)
		}

		action = types.Expression{
			Name: "IF",
			Data: []types.ExpressionArg{
				{DType: "num", Num: 1},
				{DType: "exp", Exp: &action2},
			},
		}
		newState, err = ActionEval(action, participantState, event, nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if newState.StudyStatus != action2.Data[0].Str {
			t.Errorf("error -> expected: %s, have: %s", action.Data[1].Str, newState.StudyStatus)
		}

		action = types.Expression{
			Name: "IF",
			Data: []types.ExpressionArg{
				{DType: "num", Num: 0},
				{DType: "exp", Exp: &action2},
				{DType: "exp", Exp: &action3},
			},
		}
		newState, err = ActionEval(action, participantState, event, nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if newState.StudyStatus != action3.Data[0].Str {
			t.Errorf("error -> expected: %s, have: %s", action3.Data[0].Str, newState.StudyStatus)
		}
	})

	t.Run("UPDATE_FLAG", func(t *testing.T) {
		action := types.Expression{
			Name: "UPDATE_FLAG",
			Data: []types.ExpressionArg{
				{DType: "str", Str: "key"},
				{DType: "str", Str: "value"},
			},
		}
		newState, err := ActionEval(action, participantState, event, nil)
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

	t.Run("UPDATE_FLAG with number", func(t *testing.T) {
		action := types.Expression{
			Name: "UPDATE_FLAG",
			Data: []types.ExpressionArg{
				{DType: "str", Str: "keyNum"},
				{DType: "num", Num: 14},
			},
		}
		newState, err := ActionEval(action, participantState, event, nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		v, ok := newState.Flags["keyNum"]
		if !ok {
			t.Error("could not find new flag")
			return
		}
		res, err := strconv.ParseFloat(v, 64)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if res != action.Data[1].Num {
			t.Errorf("updated status error -> expected: %s, have: %s", action.Data[1].Str, v)
		}
	})

	t.Run("REMOVE_FLAG", func(t *testing.T) {
		action := types.Expression{
			Name: "REMOVE_FLAG",
			Data: []types.ExpressionArg{
				{DType: "str", Str: "health"},
			},
		}
		newState, err := ActionEval(action, participantState, event, nil)
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
		action := types.Expression{
			Name: "ADD_NEW_SURVEY",
			Data: []types.ExpressionArg{
				{DType: "str", Str: "testSurveyKey"},
				{DType: "num", Num: float64(now - 10)},
				{DType: "num", Num: float64(now + 10)},
				{DType: "str", Str: types.ASSIGNED_SURVEY_CATEGORY_NORMAL},
			},
		}
		newState, err := ActionEval(action, participantState, event, nil)
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
		action := types.Expression{
			Name: "ADD_NEW_SURVEY",
			Data: []types.ExpressionArg{
				{DType: "str", Str: "testSurveyKey"},
				{DType: "num", Num: float64(now - 10)},
				{DType: "num", Num: float64(now + 10)},
				{DType: "str", Str: types.ASSIGNED_SURVEY_CATEGORY_NORMAL},
			},
		}
		newState, err := ActionEval(action, participantState, event, nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		newState, err = ActionEval(action, newState, event, nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(newState.AssignedSurveys) != 2 {
			t.Errorf("unexpected number of surveys: %d", len(newState.AssignedSurveys))
			return
		}

		// REMOVE_ALL_SURVEYS
		action = types.Expression{
			Name: "REMOVE_ALL_SURVEYS",
		}
		newState, err = ActionEval(action, newState, event, nil)
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
		action := types.Expression{
			Name: "ADD_NEW_SURVEY",
			Data: []types.ExpressionArg{
				{DType: "str", Str: "testSurveyKey1"},
				{DType: "num", Num: float64(now - 10)},
				{DType: "num", Num: float64(now + 10)},
				{DType: "str", Str: types.ASSIGNED_SURVEY_CATEGORY_NORMAL},
			},
		}
		newState, err := ActionEval(action, participantState, event, nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		action.Data[0].Str = "testSurveyKey2"
		newState, err = ActionEval(action, newState, event, nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		action.Data[0].Str = "testSurveyKey1"
		newState, err = ActionEval(action, newState, event, nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(newState.AssignedSurveys) != 3 {
			t.Errorf("unexpected number of surveys: %d", len(newState.AssignedSurveys))
			return
		}

		// REMOVE_SURVEY_BY_KEY
		action = types.Expression{
			Name: "REMOVE_SURVEY_BY_KEY",
			Data: []types.ExpressionArg{
				{DType: "str", Str: "testSurveyKey1"},
				{DType: "str", Str: "last"},
			},
		}
		newState, err = ActionEval(action, newState, event, nil)
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
		action := types.Expression{
			Name: "ADD_NEW_SURVEY",
			Data: []types.ExpressionArg{
				{DType: "str", Str: "testSurveyKey1"},
				{DType: "num", Num: float64(now - 10)},
				{DType: "num", Num: float64(now + 10)},
				{DType: "str", Str: types.ASSIGNED_SURVEY_CATEGORY_NORMAL},
			},
		}
		newState, err := ActionEval(action, participantState, event, nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		action.Data[0].Str = "testSurveyKey2"
		newState, err = ActionEval(action, newState, event, nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		action.Data[0].Str = "testSurveyKey1"
		newState, err = ActionEval(action, newState, event, nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(newState.AssignedSurveys) != 3 {
			t.Errorf("unexpected number of surveys: %d", len(newState.AssignedSurveys))
			return
		}

		// REMOVE_SURVEY_BY_KEY
		action = types.Expression{
			Name: "REMOVE_SURVEY_BY_KEY",
			Data: []types.ExpressionArg{
				{DType: "str", Str: "testSurveyKey1"},
				{DType: "str", Str: "first"},
			},
		}
		newState, err = ActionEval(action, newState, event, nil)
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
		action := types.Expression{
			Name: "ADD_NEW_SURVEY",
			Data: []types.ExpressionArg{
				{DType: "str", Str: "testSurveyKey1"},
				{DType: "num", Num: float64(now - 10)},
				{DType: "num", Num: float64(now + 10)},
				{DType: "str", Str: types.ASSIGNED_SURVEY_CATEGORY_NORMAL},
			},
		}
		newState, err := ActionEval(action, participantState, event, nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		newState, err = ActionEval(action, newState, event, nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		action.Data[0].Str = "testSurveyKey2"
		newState, err = ActionEval(action, newState, event, nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(newState.AssignedSurveys) != 3 {
			t.Errorf("unexpected number of surveys: %d", len(newState.AssignedSurveys))
			return
		}

		// REMOVE_SURVEYS_BY_KEY
		action = types.Expression{
			Name: "REMOVE_SURVEYS_BY_KEY",
			Data: []types.ExpressionArg{
				{DType: "str", Str: "testSurveyKey1"},
			},
		}
		newState, err = ActionEval(action, newState, event, nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}

		if len(newState.AssignedSurveys) != 1 {
			t.Errorf("unexpected number of surveys: %d", len(newState.AssignedSurveys))
		}
	})

	// Report actions
	participantState = types.ParticipantState{
		ParticipantID: "participant1234",
		StudyStatus:   types.PARTICIPANT_STUDY_STATUS_ACTIVE,
		Flags: map[string]string{
			"health": "test",
		},
		Reports: []types.SurveyItemResponse{
			{Key: "test.1"},
			{Key: "test.2.1"},
			{Key: "test.1"},
		},
	}
	event = types.StudyEvent{
		Type: "SUBMIT",
		Response: types.SurveyResponse{
			Key: "test",
			Responses: []types.SurveyItemResponse{
				{Key: "test.1"},
				{Key: "test.2.1"},
				{Key: "test.2.3"},
			},
		},
	}
	t.Run("ADD_REPORT not existing key", func(t *testing.T) {
		action := types.Expression{
			Name: "ADD_REPORT",
			Data: []types.ExpressionArg{
				{DType: "str", Str: "test.2.2"},
			},
		}
		newState, err := ActionEval(action, participantState, event, nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(newState.Reports) != 3 {
			t.Errorf("unexpected number of reports: %d", len(newState.Reports))
		}
	})

	t.Run("ADD_REPORT", func(t *testing.T) {
		action := types.Expression{
			Name: "ADD_REPORT",
			Data: []types.ExpressionArg{
				{DType: "str", Str: "test.2.1"},
			},
		}
		newState, err := ActionEval(action, participantState, event, nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(newState.Reports) != 4 {
			t.Errorf("unexpected number of reports: %d", len(newState.Reports))
		}
	})

	t.Run("REMOVE_ALL_REPORTS", func(t *testing.T) {
		action := types.Expression{
			Name: "REMOVE_ALL_REPORTS",
		}
		newState, err := ActionEval(action, participantState, event, nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(newState.Reports) > 0 {
			t.Errorf("unexpected number of reports: %d", len(newState.Reports))
		}
	})

	t.Run("REMOVE_REPORT_BY_KEY first", func(t *testing.T) {
		action := types.Expression{
			Name: "REMOVE_REPORT_BY_KEY",
			Data: []types.ExpressionArg{
				{DType: "str", Str: "test.1"},
				{DType: "str", Str: "first"},
			},
		}
		newState, err := ActionEval(action, participantState, event, nil)
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
		action := types.Expression{
			Name: "REMOVE_REPORT_BY_KEY",
			Data: []types.ExpressionArg{
				{DType: "str", Str: "test.1"},
				{DType: "str", Str: "last"},
			},
		}
		newState, err := ActionEval(action, participantState, event, nil)
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
		action := types.Expression{
			Name: "REMOVE_REPORTS_BY_KEY",
			Data: []types.ExpressionArg{
				{DType: "str", Str: "test.1"},
			},
		}
		newState, err := ActionEval(action, participantState, event, nil)
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

	t.Run("ADD_MESSAGE", func(t *testing.T) {
		action := types.Expression{
			Name: "ADD_MESSAGE",
			Data: []types.ExpressionArg{
				{DType: "str", Str: "testMessage"},
				{DType: "num", Num: float64(time.Now().Unix() - 10)},
			},
		}
		newState, err := ActionEval(action, participantState, event, nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(newState.Messages) != 1 {
			t.Errorf("unexpected number of messages: %d", len(newState.Messages))
			return
		}
		if newState.Messages[0].Type != "testMessage" {
			t.Errorf("unexpected message: %s", newState.Messages[0].Type)
		}

		action2 := types.Expression{
			Name: "ADD_MESSAGE",
			Data: []types.ExpressionArg{
				{DType: "str", Str: "testMessage2"},
				{DType: "exp", Exp: &types.Expression{
					Name: "timestampWithOffset",
					Data: []types.ExpressionArg{
						{DType: "num", Num: 0},
					},
				}},
			},
		}
		newState, err = ActionEval(action2, newState, event, nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(newState.Messages) != 2 {
			t.Errorf("unexpected number of messages: %d", len(newState.Messages))
			return
		}
		if newState.Messages[1].Type != "testMessage2" {
			t.Errorf("unexpected message: %s", newState.Messages[0].Type)
		}
		participantState = newState
	})

	t.Run("REMOVE_ALL_MESSAGES", func(t *testing.T) {
		action := types.Expression{
			Name: "REMOVE_ALL_MESSAGES",
			Data: []types.ExpressionArg{},
		}

		if len(participantState.Messages) != 2 {
			t.Errorf("unexpected number of messages: %d", len(participantState.Messages))
		}

		newState, err := ActionEval(action, participantState, event, nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(newState.Messages) != 0 {
			t.Errorf("unexpected number of messages: %d", len(newState.Messages))
			return
		}
	})

	t.Run("REMOVE_MESSAGES_BY_TYPE", func(t *testing.T) {
		action := types.Expression{
			Name: "REMOVE_MESSAGES_BY_TYPE",
			Data: []types.ExpressionArg{
				{DType: "str", Str: "testMessage1"},
			},
		}
		tP := types.ParticipantState{
			Messages: []types.ParticipantMessage{
				{Type: "testMessage1"},
				{Type: "testMessage1"},
				{Type: "testMessage2"},
				{Type: "testMessage3"},
				{Type: "testMessage4"},
			},
		}

		newState, err := ActionEval(action, tP, event, nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(newState.Messages) != 3 {
			t.Errorf("unexpected number of messages: %d", len(newState.Messages))
			return
		}
	})
}
