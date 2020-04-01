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
		action := models.Action{
			Name: "WRONG",
		}
		_, err := ActionEval(action, participantState, event)
		if err == nil {
			t.Error("should return an error")
		}
	})

	t.Run("IFTHEN", func(t *testing.T) {
		t.Error("test unimplemented")
	})

	t.Run("UPDATE_FLAG_STATUS", func(t *testing.T) {
		action := models.Action{
			Name: "UPDATE_FLAG_STATUS",
			Args: []models.ExpressionArg{
				models.ExpressionArg{DType: "str", Str: "test2"},
			},
		}
		newState, err := ActionEval(action, participantState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if newState.Flags.Status != action.Args[0].Str {
			t.Errorf("updated status error -> expected: %s, have: %s", action.Args[0].Str, newState.Flags.Status)
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
