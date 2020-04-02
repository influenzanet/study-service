package studyengine

import (
	"testing"

	"github.com/influenzanet/study-service/models"
)

func TestActions(t *testing.T) {
	participantState := models.ParticipantState{
		ParticipantID: "participant1234",
		Flags: models.ParticipantStateFlags{
			Status: "test",
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
			Name: "UPDATE_FLAG",
			Data: []models.ExpressionArg{
				models.ExpressionArg{DType: "str", Str: "status"},
				models.ExpressionArg{DType: "str", Str: "testflag_cond"},
			},
		}
		action3 := models.Expression{
			Name: "UPDATE_FLAG",
			Data: []models.ExpressionArg{
				models.ExpressionArg{DType: "str", Str: "status"},
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
		if newState.Flags.Status == action2.Data[1].Str {
			t.Errorf("error -> expected: %s, have: %s", action.Data[1].Str, newState.Flags.Status)
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
		if newState.Flags.Status != action3.Data[1].Str {
			t.Errorf("error -> expected: %s, have: %s", action3.Data[1].Str, newState.Flags.Status)
		}
	})

	t.Run("UPDATE_FLAG", func(t *testing.T) {
		action := models.Expression{
			Name: "UPDATE_FLAG",
			Data: []models.ExpressionArg{
				models.ExpressionArg{DType: "str", Str: "status"},
				models.ExpressionArg{DType: "str", Str: "test2"},
			},
		}
		newState, err := ActionEval(action, participantState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if newState.Flags.Status != action.Data[1].Str {
			t.Errorf("updated status error -> expected: %s, have: %s", action.Data[1].Str, newState.Flags.Status)
		}
	})

	// Survey actions:
	t.Run("ADD_NEW_SURVEY", func(t *testing.T) {
		t.Error("test unimplemented")
	})
	t.Run("REMOVE_ALL_SURVEYS", func(t *testing.T) {
		t.Error("test unimplemented")
	})
	t.Run("REMOVE_SURVEY_BY_KEY", func(t *testing.T) {
		t.Error("test unimplemented")
	})
	t.Run("REMOVE_SURVEYS_BY_KEY", func(t *testing.T) {
		t.Error("test unimplemented")
	})

	// Report actions
	t.Run("ADD_REPORT", func(t *testing.T) {
		t.Error("test unimplemented")
	})
	t.Run("REMOVE_ALL_REPORTS", func(t *testing.T) {
		t.Error("test unimplemented")
	})
	t.Run("REMOVE_REPORT_BY_KEY", func(t *testing.T) {
		t.Error("test unimplemented")
	})
	t.Run("REMOVE_REPORTS_BY_KEY", func(t *testing.T) {
		t.Error("test unimplemented")
	})
}
