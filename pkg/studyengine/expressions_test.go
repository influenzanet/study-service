package studyengine

import (
	"testing"
	"time"

	"github.com/influenzanet/study-service/pkg/dbs/studydb"
	"github.com/influenzanet/study-service/pkg/types"
)

// Reference/Lookup methods
func TestEvalCheckEventType(t *testing.T) {
	exp := types.Expression{Name: "checkEventType", Data: []types.ExpressionArg{
		{DType: "str", Str: "ENTER"},
	}}

	t.Run("for matching", func(t *testing.T) {
		EvalContext := EvalContext{
			Event: types.StudyEvent{Type: "ENTER"},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if !ret.(bool) {
			t.Errorf("unexpected type or value: %s", ret)
		}
	})

	t.Run("for not matching", func(t *testing.T) {
		EvalContext := EvalContext{
			Event: types.StudyEvent{Type: "enter"},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected type or value: %s", ret)
		}
	})
}

func TestEvalCheckSurveyResponseKey(t *testing.T) {
	exp := types.Expression{Name: "checkSurveyResponseKey", Data: []types.ExpressionArg{
		{DType: "str", Str: "weekly"},
	}}

	t.Run("for no survey responses at all", func(t *testing.T) {
		EvalContext := EvalContext{
			Event: types.StudyEvent{Type: "SUBMIT"},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected type or value: %s", ret)
		}
	})

	t.Run("not matching key", func(t *testing.T) {
		EvalContext := EvalContext{
			Event: types.StudyEvent{
				Type: "SUBMIT",
				Response: types.SurveyResponse{
					Key:       "intake",
					Responses: []types.SurveyItemResponse{},
				},
			},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected type or value: %s", ret)
		}
	})

	t.Run("for matching key", func(t *testing.T) {
		EvalContext := EvalContext{
			Event: types.StudyEvent{
				Type: "SUBMIT",
				Response: types.SurveyResponse{
					Key:       "weekly",
					Responses: []types.SurveyItemResponse{},
				},
			},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if !ret.(bool) {
			t.Errorf("unexpected type or value: %s", ret)
		}
	})
}

func TestEvalHasStudyStatus(t *testing.T) {
	t.Run("with not matching state", func(t *testing.T) {
		exp := types.Expression{Name: "hasStudyStatus", Data: []types.ExpressionArg{
			{DType: "str", Str: types.PARTICIPANT_STUDY_STATUS_EXITED},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{StudyStatus: types.PARTICIPANT_STUDY_STATUS_ACTIVE},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})

	t.Run("with matching state", func(t *testing.T) {
		exp := types.Expression{Name: "hasStudyStatus", Data: []types.ExpressionArg{
			{DType: "str", Str: types.PARTICIPANT_STUDY_STATUS_ACTIVE},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{StudyStatus: types.PARTICIPANT_STUDY_STATUS_ACTIVE},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if !ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})
}

type MockStudyDBService struct {
	Responses []types.SurveyResponse
}

func (db MockStudyDBService) FindSurveyResponses(instanceID string, studyKey string, query studydb.ResponseQuery) (responses []types.SurveyResponse, err error) {

	for _, r := range db.Responses {
		if query.ParticipantID != r.ParticipantID {
			continue
		}
		if len(query.SurveyKey) > 0 && query.SurveyKey != r.Key {
			continue
		}
		if query.Since > 0 && r.SubmittedAt < query.Since {
			continue
		}
		if query.Until > 0 && r.SubmittedAt > query.Until {
			continue
		}
		responses = append(responses, r)
	}

	return responses, nil
}

func TestEvalCheckConditionForOldResponses(t *testing.T) {
	testResponses := []types.SurveyResponse{
		{
			Key: "S1", ParticipantID: "P1", SubmittedAt: 10, Responses: []types.SurveyItemResponse{
				{Key: "S1.Q1", Response: &types.ResponseItem{
					Key: "rg", Items: []types.ResponseItem{
						{Key: "scg", Items: []types.ResponseItem{{Key: "1"}}},
					},
				}}},
		},
		{
			Key: "S1", ParticipantID: "P1", SubmittedAt: 13, Responses: []types.SurveyItemResponse{
				{Key: "S1.Q1", Response: &types.ResponseItem{
					Key: "rg", Items: []types.ResponseItem{
						{Key: "scg", Items: []types.ResponseItem{{Key: "1"}}},
					},
				}}},
		},
		{
			Key: "S1", ParticipantID: "P2", SubmittedAt: 13, Responses: []types.SurveyItemResponse{
				{Key: "S1.Q1", Response: &types.ResponseItem{
					Key: "rg", Items: []types.ResponseItem{
						{Key: "scg", Items: []types.ResponseItem{{Key: "1"}}},
					},
				}}},
		},
		{
			Key: "S2", ParticipantID: "P1", SubmittedAt: 15, Responses: []types.SurveyItemResponse{
				{Key: "S2.Q1", Response: &types.ResponseItem{
					Key: "rg", Items: []types.ResponseItem{
						{Key: "scg", Items: []types.ResponseItem{{Key: "1"}}},
					},
				}}},
		},
		{
			Key: "S1", ParticipantID: "P1", SubmittedAt: 17, Responses: []types.SurveyItemResponse{
				{Key: "S1.Q1", Response: &types.ResponseItem{
					Key: "rg", Items: []types.ResponseItem{
						{Key: "scg", Items: []types.ResponseItem{{Key: "1"}}},
					},
				}}},
		},
		{
			Key: "S1", ParticipantID: "P1", SubmittedAt: 22, Responses: []types.SurveyItemResponse{
				{Key: "S1.Q1", Response: &types.ResponseItem{
					Key: "rg", Items: []types.ResponseItem{
						{Key: "scg", Items: []types.ResponseItem{{Key: "2"}}},
					},
				}}},
		},
	}

	t.Run("missing DB config", func(t *testing.T) {
		exp := types.Expression{Name: "checkConditionForOldResponses"}

		EvalContext := EvalContext{
			DbService: nil,
		}
		_, err := ExpressionEval(exp, EvalContext)
		if err == nil {
			t.Error("should return error")
			return
		}
	})

	t.Run("missing instanceID", func(t *testing.T) {
		exp := types.Expression{Name: "checkConditionForOldResponses"}

		EvalContext := EvalContext{
			DbService: MockStudyDBService{},
			Event: types.StudyEvent{
				StudyKey: "testStudy",
			},
		}
		_, err := ExpressionEval(exp, EvalContext)
		if err == nil {
			t.Error("should return error")
			return
		}
	})

	t.Run("missing studyKey", func(t *testing.T) {
		exp := types.Expression{Name: "checkConditionForOldResponses"}

		EvalContext := EvalContext{
			DbService: MockStudyDBService{
				Responses: testResponses,
			},
			Event: types.StudyEvent{
				StudyKey:   "testStudy",
				InstanceID: "testInstance",
			},
		}
		_, err := ExpressionEval(exp, EvalContext)
		if err == nil {
			t.Error("should return error")
			return
		}
	})

	t.Run("missing condition", func(t *testing.T) {
		exp := types.Expression{Name: "checkConditionForOldResponses"}

		EvalContext := EvalContext{
			DbService: MockStudyDBService{
				Responses: testResponses,
			},
			Event: types.StudyEvent{
				StudyKey:   "testStudy",
				InstanceID: "testInstance",
			},
			ParticipantState: types.ParticipantState{
				ParticipantID: "P1",
			},
		}
		_, err := ExpressionEval(exp, EvalContext)
		if err == nil {
			t.Error("should return error")
			return
		}
	})

	t.Run("checkType all", func(t *testing.T) {
		exp := types.Expression{Name: "checkConditionForOldResponses", Data: []types.ExpressionArg{
			{Exp: &types.Expression{
				Name: "responseHasKeysAny",
				Data: []types.ExpressionArg{
					{Str: "S1.Q1", DType: "str"},
					{Str: "rg.scg", DType: "str"},
					{Str: "1", DType: "str"},
				},
			}, DType: "exp"},
		}}

		EvalContext := EvalContext{
			DbService: MockStudyDBService{
				Responses: testResponses,
			},
			Event: types.StudyEvent{
				StudyKey:   "testStudy",
				InstanceID: "testInstance",
			},
			ParticipantState: types.ParticipantState{
				ParticipantID: "P1",
			},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected value retrieved: %d", ret)
		}
	})

	t.Run("checkType any", func(t *testing.T) {
		exp := types.Expression{Name: "checkConditionForOldResponses", Data: []types.ExpressionArg{
			{Exp: &types.Expression{
				Name: "responseHasKeysAny",
				Data: []types.ExpressionArg{
					{Str: "S1.Q1", DType: "str"},
					{Str: "rg.scg", DType: "str"},
					{Str: "1", DType: "str"},
				},
			}, DType: "exp"},
			{Str: "any", DType: "str"},
		}}

		EvalContext := EvalContext{
			DbService: MockStudyDBService{
				Responses: testResponses,
			},
			Event: types.StudyEvent{
				StudyKey:   "testStudy",
				InstanceID: "testInstance",
			},
			ParticipantState: types.ParticipantState{
				ParticipantID: "P1",
			},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if !ret.(bool) {
			t.Errorf("unexpected value retrieved: %d", ret)
		}
	})

	t.Run("checkType count - with enough", func(t *testing.T) {
		exp := types.Expression{Name: "checkConditionForOldResponses", Data: []types.ExpressionArg{
			{Exp: &types.Expression{
				Name: "responseHasKeysAny",
				Data: []types.ExpressionArg{
					{Str: "S1.Q1", DType: "str"},
					{Str: "rg.scg", DType: "str"},
					{Str: "1", DType: "str"},
				},
			}, DType: "exp"},
			{Num: 3, DType: "num"},
		}}

		EvalContext := EvalContext{
			DbService: MockStudyDBService{
				Responses: testResponses,
			},
			Event: types.StudyEvent{
				StudyKey:   "testStudy",
				InstanceID: "testInstance",
			},
			ParticipantState: types.ParticipantState{
				ParticipantID: "P1",
			},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if !ret.(bool) {
			t.Errorf("unexpected value retrieved: %d", ret)
		}
	})

	t.Run("checkType count - with not enough", func(t *testing.T) {
		exp := types.Expression{Name: "checkConditionForOldResponses", Data: []types.ExpressionArg{
			{Exp: &types.Expression{
				Name: "responseHasKeysAny",
				Data: []types.ExpressionArg{
					{Str: "S1.Q1", DType: "str"},
					{Str: "rg.scg", DType: "str"},
					{Str: "1", DType: "str"},
				},
			}, DType: "exp"},
			{Num: 4, DType: "num"},
		}}

		EvalContext := EvalContext{
			DbService: MockStudyDBService{
				Responses: testResponses,
			},
			Event: types.StudyEvent{
				StudyKey:   "testStudy",
				InstanceID: "testInstance",
			},
			ParticipantState: types.ParticipantState{
				ParticipantID: "P1",
			},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected value retrieved: %d", ret)
		}
	})

	t.Run("filter for survey type", func(t *testing.T) {
		exp := types.Expression{Name: "checkConditionForOldResponses", Data: []types.ExpressionArg{
			{Exp: &types.Expression{
				Name: "responseHasKeysAny",
				Data: []types.ExpressionArg{
					{Str: "S1.Q1", DType: "str"},
					{Str: "rg.scg", DType: "str"},
					{Str: "1", DType: "str"},
				},
			}, DType: "exp"},
			{Num: 4, DType: "num"},
			{Str: "S2", DType: "str"},
		}}

		EvalContext := EvalContext{
			DbService: MockStudyDBService{
				Responses: testResponses,
			},
			Event: types.StudyEvent{
				StudyKey:   "testStudy",
				InstanceID: "testInstance",
			},
			ParticipantState: types.ParticipantState{
				ParticipantID: "P1",
			},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected value retrieved: %d", ret)
		}
	})

	t.Run("filter for interval type", func(t *testing.T) {
		exp := types.Expression{Name: "checkConditionForOldResponses", Data: []types.ExpressionArg{
			{Exp: &types.Expression{
				Name: "responseHasKeysAny",
				Data: []types.ExpressionArg{
					{Str: "S1.Q1", DType: "str"},
					{Str: "rg.scg", DType: "str"},
					{Str: "1", DType: "str"},
				},
			}, DType: "exp"},
			{Num: 2, DType: "num"},
			{Str: "", DType: "str"},
			{Num: 16, DType: "num"},
			{Num: 18, DType: "num"},
		}}

		EvalContext := EvalContext{
			DbService: MockStudyDBService{
				Responses: testResponses,
			},
			Event: types.StudyEvent{
				StudyKey:   "testStudy",
				InstanceID: "testInstance",
			},
			ParticipantState: types.ParticipantState{
				ParticipantID: "P1",
			},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected value retrieved: %d", ret)
		}
	})
}

func TestEvalGetStudyEntryTime(t *testing.T) {
	t.Run("try retrieve entered at time", func(t *testing.T) {
		exp := types.Expression{Name: "getStudyEntryTime"}
		tStart := time.Now().Unix()
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{
				StudyStatus: types.PARTICIPANT_STUDY_STATUS_ACTIVE,
				EnteredAt:   tStart,
			},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if ret.(float64) != float64(tStart) {
			t.Errorf("unexpected value retrieved: %d", ret)
		}
	})
}

func TestEvalHasSurveyKeyAssigned(t *testing.T) {
	t.Run("has survey assigned", func(t *testing.T) {
		exp := types.Expression{Name: "hasSurveyKeyAssigned", Data: []types.ExpressionArg{
			{DType: "str", Str: "test1"},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{
				StudyStatus: types.PARTICIPANT_STUDY_STATUS_ACTIVE,
				AssignedSurveys: []types.AssignedSurvey{
					{SurveyKey: "test1"},
					{SurveyKey: "test2"},
				},
			},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if !ret.(bool) {
			t.Errorf("unexpected value retrieved: %d", ret)
		}
	})

	t.Run("doesn't have the survey assigned", func(t *testing.T) {
		exp := types.Expression{Name: "hasSurveyKeyAssigned", Data: []types.ExpressionArg{
			{DType: "str", Str: "test1"},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{
				StudyStatus: types.PARTICIPANT_STUDY_STATUS_ACTIVE,
				AssignedSurveys: []types.AssignedSurvey{
					{SurveyKey: "test2"},
				},
			},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected value retrieved: %d", ret)
		}
	})

	t.Run("missing argument", func(t *testing.T) {
		exp := types.Expression{Name: "hasSurveyKeyAssigned", Data: []types.ExpressionArg{}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{
				StudyStatus: types.PARTICIPANT_STUDY_STATUS_ACTIVE,
				AssignedSurveys: []types.AssignedSurvey{
					{SurveyKey: "test2"},
				},
			},
		}
		_, err := ExpressionEval(exp, EvalContext)
		if err == nil {
			t.Error("should throw an error about missing arg")
			return
		}
	})

	t.Run("wrong argument", func(t *testing.T) {
		exp := types.Expression{Name: "hasSurveyKeyAssigned", Data: []types.ExpressionArg{
			{DType: "exp", Exp: &types.Expression{}},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{
				StudyStatus: types.PARTICIPANT_STUDY_STATUS_ACTIVE,
				AssignedSurveys: []types.AssignedSurvey{
					{SurveyKey: "test2"},
				},
			},
		}
		_, err := ExpressionEval(exp, EvalContext)
		if err == nil {
			t.Error("should throw an error about arg type")
			return
		}
	})
}

func TestEvalGetSurveyKeyAssignedFrom(t *testing.T) {
	t.Run("has survey assigned", func(t *testing.T) {
		exp := types.Expression{Name: "getSurveyKeyAssignedFrom", Data: []types.ExpressionArg{
			{DType: "str", Str: "test1"},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{
				StudyStatus: types.PARTICIPANT_STUDY_STATUS_ACTIVE,
				AssignedSurveys: []types.AssignedSurvey{
					{SurveyKey: "test1", ValidFrom: 10, ValidUntil: 100},
					{SurveyKey: "test2", ValidFrom: 10, ValidUntil: 100},
				},
			},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if ret.(float64) != 10 {
			t.Errorf("unexpected value retrieved: %d", ret)
		}
	})

	t.Run("doesn't have the survey assigned", func(t *testing.T) {
		exp := types.Expression{Name: "getSurveyKeyAssignedFrom", Data: []types.ExpressionArg{
			{DType: "str", Str: "test1"},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{
				StudyStatus: types.PARTICIPANT_STUDY_STATUS_ACTIVE,
				AssignedSurveys: []types.AssignedSurvey{
					{SurveyKey: "test2", ValidFrom: 10, ValidUntil: 100},
				},
			},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if ret.(float64) != -1 {
			t.Errorf("unexpected value retrieved: %d", ret)
		}
	})

	t.Run("missing argument", func(t *testing.T) {
		exp := types.Expression{Name: "getSurveyKeyAssignedFrom", Data: []types.ExpressionArg{}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{
				StudyStatus: types.PARTICIPANT_STUDY_STATUS_ACTIVE,
				AssignedSurveys: []types.AssignedSurvey{
					{SurveyKey: "test1", ValidFrom: 10, ValidUntil: 100},
					{SurveyKey: "test2", ValidFrom: 10, ValidUntil: 100},
				},
			},
		}
		_, err := ExpressionEval(exp, EvalContext)
		if err == nil {
			t.Error("should throw an error about missing arg")
			return
		}
	})

	t.Run("wrong argument", func(t *testing.T) {
		exp := types.Expression{Name: "getSurveyKeyAssignedFrom", Data: []types.ExpressionArg{
			{DType: "exp", Exp: &types.Expression{}},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{
				StudyStatus: types.PARTICIPANT_STUDY_STATUS_ACTIVE,
				AssignedSurveys: []types.AssignedSurvey{
					{SurveyKey: "test1", ValidFrom: 10, ValidUntil: 100},
					{SurveyKey: "test2", ValidFrom: 10, ValidUntil: 100},
				},
			},
		}
		_, err := ExpressionEval(exp, EvalContext)
		if err == nil {
			t.Error("should throw an error about arg type")
			return
		}
	})
}

func TestEvalGetSurveyKeyAssignedUntil(t *testing.T) {
	t.Run("has survey assigned", func(t *testing.T) {
		exp := types.Expression{Name: "getSurveyKeyAssignedUntil", Data: []types.ExpressionArg{
			{DType: "str", Str: "test1"},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{
				StudyStatus: types.PARTICIPANT_STUDY_STATUS_ACTIVE,
				AssignedSurveys: []types.AssignedSurvey{
					{SurveyKey: "test1", ValidFrom: 10, ValidUntil: 100},
					{SurveyKey: "test2", ValidFrom: 10, ValidUntil: 100},
				},
			},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if ret.(float64) != 100 {
			t.Errorf("unexpected value retrieved: %d", ret)
		}
	})

	t.Run("doesn't have the survey assigned", func(t *testing.T) {
		exp := types.Expression{Name: "getSurveyKeyAssignedUntil", Data: []types.ExpressionArg{
			{DType: "str", Str: "test1"},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{
				StudyStatus: types.PARTICIPANT_STUDY_STATUS_ACTIVE,
				AssignedSurveys: []types.AssignedSurvey{
					{SurveyKey: "test2", ValidFrom: 10, ValidUntil: 100},
				},
			},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if ret.(float64) != -1 {
			t.Errorf("unexpected value retrieved: %d", ret)
		}
	})

	t.Run("missing argument", func(t *testing.T) {
		exp := types.Expression{Name: "getSurveyKeyAssignedUntil", Data: []types.ExpressionArg{}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{
				StudyStatus: types.PARTICIPANT_STUDY_STATUS_ACTIVE,
				AssignedSurveys: []types.AssignedSurvey{
					{SurveyKey: "test1", ValidFrom: 10, ValidUntil: 100},
					{SurveyKey: "test2", ValidFrom: 10, ValidUntil: 100},
				},
			},
		}
		_, err := ExpressionEval(exp, EvalContext)
		if err == nil {
			t.Error("should throw an error about missing arg")
			return
		}
	})

	t.Run("wrong argument", func(t *testing.T) {
		exp := types.Expression{Name: "getSurveyKeyAssignedUntil", Data: []types.ExpressionArg{
			{DType: "exp", Exp: &types.Expression{}},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{
				StudyStatus: types.PARTICIPANT_STUDY_STATUS_ACTIVE,
				AssignedSurveys: []types.AssignedSurvey{
					{SurveyKey: "test1", ValidFrom: 10, ValidUntil: 100},
					{SurveyKey: "test2", ValidFrom: 10, ValidUntil: 100},
				},
			},
		}
		_, err := ExpressionEval(exp, EvalContext)
		if err == nil {
			t.Error("should throw an error about arg type")
			return
		}
	})
}

func TestEvalHasParticipantFlag(t *testing.T) {
	t.Run("participant hasn't got any participant flags (empty / nil)", func(t *testing.T) {
		exp := types.Expression{Name: "hasParticipantFlag", Data: []types.ExpressionArg{
			{DType: "str", Str: "key1"},
			{DType: "str", Str: "value1"},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{
				StudyStatus: types.PARTICIPANT_STUDY_STATUS_ACTIVE,
			},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if ret.(bool) {
			t.Error("should be false")
		}
	})

	t.Run("participant has other participant flags, but this key is missing", func(t *testing.T) {
		exp := types.Expression{Name: "hasParticipantFlag", Data: []types.ExpressionArg{
			{DType: "str", Str: "key1"},
			{DType: "str", Str: "value1"},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{
				StudyStatus: types.PARTICIPANT_STUDY_STATUS_ACTIVE,
				Flags: map[string]string{
					"key2": "value1",
				},
			},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if ret.(bool) {
			t.Error("should be false")
		}
	})

	t.Run("participant has correct participant flag's key, but value is different", func(t *testing.T) {
		exp := types.Expression{Name: "hasParticipantFlag", Data: []types.ExpressionArg{
			{DType: "str", Str: "key1"},
			{DType: "str", Str: "value1"},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{
				StudyStatus: types.PARTICIPANT_STUDY_STATUS_ACTIVE,
				Flags: map[string]string{
					"key1": "value2",
				},
			},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if ret.(bool) {
			t.Error("should be false")
		}
	})

	t.Run("participant has correct participant flag's key and value is same", func(t *testing.T) {
		exp := types.Expression{Name: "hasParticipantFlag", Data: []types.ExpressionArg{
			{DType: "str", Str: "key1"},
			{DType: "str", Str: "value1"},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{
				StudyStatus: types.PARTICIPANT_STUDY_STATUS_ACTIVE,
				Flags: map[string]string{
					"key1": "value1",
				},
			},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if !ret.(bool) {
			t.Error("should be true")
		}
	})

	t.Run("missing arguments", func(t *testing.T) {
		exp := types.Expression{Name: "hasParticipantFlag", Data: []types.ExpressionArg{}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{
				StudyStatus: types.PARTICIPANT_STUDY_STATUS_ACTIVE,
				Flags: map[string]string{
					"key1": "value1",
				},
			},
		}
		_, err := ExpressionEval(exp, EvalContext)
		if err == nil {
			t.Error("should throw error")
			return
		}
	})

	t.Run("using num at 1st argument (expressions allowed, should return string)", func(t *testing.T) {
		exp := types.Expression{Name: "hasParticipantFlag", Data: []types.ExpressionArg{
			{DType: "num", Num: 22},
			{DType: "str", Str: "value1"},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{
				StudyStatus: types.PARTICIPANT_STUDY_STATUS_ACTIVE,
				Flags: map[string]string{
					"key1": "value1",
				},
			},
		}
		_, err := ExpressionEval(exp, EvalContext)
		if err == nil {
			t.Error("should throw error")
			return
		}
	})

	t.Run("missing arguments", func(t *testing.T) {
		exp := types.Expression{Name: "hasParticipantFlag", Data: []types.ExpressionArg{
			{DType: "str", Str: "key1"},
			{DType: "num", Num: 22},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{
				StudyStatus: types.PARTICIPANT_STUDY_STATUS_ACTIVE,
				Flags: map[string]string{
					"key1": "value1",
				},
			},
		}
		_, err := ExpressionEval(exp, EvalContext)
		if err == nil {
			t.Error("should throw error")
			return
		}
	})
}

func TestEvalHasParticipantFlagKey(t *testing.T) {
	t.Run("participant hasn't got any participant flags (empty / nil)", func(t *testing.T) {
		exp := types.Expression{Name: "hasParticipantFlagKey", Data: []types.ExpressionArg{
			{DType: "str", Str: "key1"},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{
				StudyStatus: types.PARTICIPANT_STUDY_STATUS_ACTIVE,
			},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if ret.(bool) {
			t.Error("should be false")
		}
	})

	t.Run("participant has other key", func(t *testing.T) {
		exp := types.Expression{Name: "hasParticipantFlagKey", Data: []types.ExpressionArg{
			{DType: "str", Str: "key1"},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{
				StudyStatus: types.PARTICIPANT_STUDY_STATUS_ACTIVE,
				Flags: map[string]string{
					"key2": "1",
				},
			},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if ret.(bool) {
			t.Error("should be false")
		}
	})

	t.Run("participant has correct key", func(t *testing.T) {
		exp := types.Expression{Name: "hasParticipantFlagKey", Data: []types.ExpressionArg{
			{DType: "str", Str: "key1"},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{
				StudyStatus: types.PARTICIPANT_STUDY_STATUS_ACTIVE,
				Flags: map[string]string{
					"key2": "1",
					"key1": "1",
				},
			},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if !ret.(bool) {
			t.Error("should be true")
		}
	})
}

func TestEvalGetResponseValueAsNum(t *testing.T) {
	testEvalContext := EvalContext{
		Event: types.StudyEvent{
			Type: "SUBMIT",
			Response: types.SurveyResponse{
				Key: "weekly",
				Responses: []types.SurveyItemResponse{
					{
						Key: "weekly.Q1", Response: &types.ResponseItem{
							Key: "rg", Items: []types.ResponseItem{
								{Key: "1", Value: "something"},
								{Key: "2"},
							}},
					},
					{
						Key: "weekly.Q2", Response: &types.ResponseItem{
							Key: "rg", Items: []types.ResponseItem{
								{Key: "1", Value: "123.23", Dtype: "date"},
							}},
					},
				},
			},
		},
	}

	//
	t.Run("no survey item response found", func(t *testing.T) {
		exp := types.Expression{Name: "getResponseValueAsNum", Data: []types.ExpressionArg{
			{DType: "str", Str: "weekly.Q3"},
			{DType: "str", Str: "rg.1"},
		}}
		_, err := ExpressionEval(exp, testEvalContext)
		if err == nil {
			t.Error("should return an error")
			return
		}
	})

	t.Run("repsonse item in question missing", func(t *testing.T) {
		exp := types.Expression{Name: "getResponseValueAsNum", Data: []types.ExpressionArg{
			{DType: "str", Str: "weekly.Q1"},
			{DType: "str", Str: "rg.wrong"},
		}}
		_, err := ExpressionEval(exp, testEvalContext)
		if err == nil {
			t.Error("should return an error")
			return
		}
	})

	t.Run("has empty value", func(t *testing.T) {
		exp := types.Expression{Name: "getResponseValueAsNum", Data: []types.ExpressionArg{
			{DType: "str", Str: "weekly.Q1"},
			{DType: "str", Str: "rg.2"},
		}}
		_, err := ExpressionEval(exp, testEvalContext)
		if err == nil {
			t.Error("should return an error")
			return
		}
	})

	t.Run("repsonse item's value is not a number", func(t *testing.T) {
		exp := types.Expression{Name: "getResponseValueAsNum", Data: []types.ExpressionArg{
			{DType: "str", Str: "weekly.Q1"},
			{DType: "str", Str: "rg.1"},
		}}
		_, err := ExpressionEval(exp, testEvalContext)
		if err == nil {
			t.Error("should return an error")
			return
		}
	})

	t.Run("is number", func(t *testing.T) {
		exp := types.Expression{Name: "getResponseValueAsNum", Data: []types.ExpressionArg{
			{DType: "str", Str: "weekly.Q2"},
			{DType: "str", Str: "rg.1"},
		}}
		v, err := ExpressionEval(exp, testEvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if v.(float64) != 123.23 {
			t.Errorf("unexpected value: %b", v)
		}
	})
}

func TestEvalCountResponseItems(t *testing.T) {
	testEvalContext := EvalContext{
		Event: types.StudyEvent{
			Type: "SUBMIT",
			Response: types.SurveyResponse{
				Key: "weekly",
				Responses: []types.SurveyItemResponse{
					{
						Key: "weekly.Q1", Response: &types.ResponseItem{
							Key: "rg", Items: []types.ResponseItem{
								{Key: "mcg", Items: []types.ResponseItem{
									{Key: "1"},
									{Key: "2"},
									{Key: "3"},
								}},
							}},
					},
					{
						Key: "weekly.Q2", Response: &types.ResponseItem{
							Key: "rg", Items: []types.ResponseItem{
								{Key: "mcg", Items: []types.ResponseItem{}},
							}},
					},
				},
			},
		},
	}

	//
	t.Run("no survey item response found", func(t *testing.T) {
		exp := types.Expression{Name: "countResponseItems", Data: []types.ExpressionArg{
			{DType: "str", Str: "weekly.Q3"},
			{DType: "str", Str: "rg.mcg"},
		}}
		_, err := ExpressionEval(exp, testEvalContext)
		if err == nil {
			t.Error("should return an error")
			return
		}
	})

	t.Run("repsonse item in question missing", func(t *testing.T) {
		exp := types.Expression{Name: "countResponseItems", Data: []types.ExpressionArg{
			{DType: "str", Str: "weekly.Q1"},
			{DType: "str", Str: "rg.wrong"},
		}}
		_, err := ExpressionEval(exp, testEvalContext)
		if err == nil {
			t.Error("should return an error")
			return
		}
	})

	t.Run("has empty value", func(t *testing.T) {
		exp := types.Expression{Name: "countResponseItems", Data: []types.ExpressionArg{
			{DType: "str", Str: "weekly.Q2"},
			{DType: "str", Str: "rg.mcg"},
		}}
		v, err := ExpressionEval(exp, testEvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if v.(float64) != 0.0 {
			t.Errorf("unexpected value: %b", v)
		}
	})

	t.Run("has 3 values", func(t *testing.T) {
		exp := types.Expression{Name: "countResponseItems", Data: []types.ExpressionArg{
			{DType: "str", Str: "weekly.Q1"},
			{DType: "str", Str: "rg.mcg"},
		}}
		v, err := ExpressionEval(exp, testEvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if v.(float64) != 3.0 {
			t.Errorf("unexpected value: %b", v)
		}
	})
}

func TestEvalGetResponseValueAsStr(t *testing.T) {
	testEvalContext := EvalContext{
		Event: types.StudyEvent{
			Type: "SUBMIT",
			Response: types.SurveyResponse{
				Key: "weekly",
				Responses: []types.SurveyItemResponse{
					{
						Key: "weekly.Q1", Response: &types.ResponseItem{
							Key: "rg", Items: []types.ResponseItem{
								{Key: "1", Value: "something"},
								{Key: "2"},
							}},
					},
				},
			},
		},
	}

	//
	t.Run("no survey item response found", func(t *testing.T) {
		exp := types.Expression{Name: "getResponseValueAsStr", Data: []types.ExpressionArg{
			{DType: "str", Str: "weekly.Q3"},
			{DType: "str", Str: "rg.1"},
		}}
		_, err := ExpressionEval(exp, testEvalContext)
		if err == nil {
			t.Error("should return an error")
			return
		}
	})

	t.Run("repsonse item in question missing", func(t *testing.T) {
		exp := types.Expression{Name: "getResponseValueAsStr", Data: []types.ExpressionArg{
			{DType: "str", Str: "weekly.Q1"},
			{DType: "str", Str: "rg.wrong"},
		}}
		_, err := ExpressionEval(exp, testEvalContext)
		if err == nil {
			t.Error("should return an error")
			return
		}
	})

	t.Run("has empty value", func(t *testing.T) {
		exp := types.Expression{Name: "getResponseValueAsStr", Data: []types.ExpressionArg{
			{DType: "str", Str: "weekly.Q1"},
			{DType: "str", Str: "rg.2"},
		}}
		v, err := ExpressionEval(exp, testEvalContext)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if v != "" {
			t.Errorf("unexpected value: %s instead of %s", v, "blank")
		}
	})

	t.Run("has value", func(t *testing.T) {
		exp := types.Expression{Name: "getResponseValueAsStr", Data: []types.ExpressionArg{
			{DType: "str", Str: "weekly.Q1"},
			{DType: "str", Str: "rg.1"},
		}}
		v, err := ExpressionEval(exp, testEvalContext)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if v != "something" {
			t.Errorf("unexpected value: %s instead of %s", v, "something")
		}
	})
}

func TestMustGetStrValue(t *testing.T) {
	testEvalContext := EvalContext{}

	t.Run("not string value", func(t *testing.T) {
		_, err := testEvalContext.mustGetStrValue(types.ExpressionArg{
			Num:   0,
			DType: "num",
		})
		if err == nil {
			t.Error("should produce error")
		}
	})

	t.Run("string value", func(t *testing.T) {
		v, err := testEvalContext.mustGetStrValue(types.ExpressionArg{
			Str:   "hello",
			DType: "str",
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if v != "hello" {
			t.Errorf("unexpected value: %s", v)
		}
	})
}

func TestEvalResponseHasOnlyKeysOtherThan(t *testing.T) {
	testEvalContext := EvalContext{
		Event: types.StudyEvent{
			Type: "SUBMIT",
			Response: types.SurveyResponse{
				Key:       "wwekly",
				Responses: []types.SurveyItemResponse{},
			},
		},
	}

	t.Run("no survey item response found", func(t *testing.T) {
		exp := types.Expression{Name: "responseHasOnlyKeysOtherThan", Data: []types.ExpressionArg{
			{DType: "str", Str: "weekly.G1.Q1"},
			{DType: "str", Str: "rg.mcg"},
			{DType: "str", Str: "1"},
			{DType: "str", Str: "2"},
		}}
		testEvalContext.Event.Response.Responses = []types.SurveyItemResponse{
			{Key: "weekly.G1.Q2", Response: &types.ResponseItem{Key: "rg", Items: []types.ResponseItem{{Key: "mcg", Items: []types.ResponseItem{
				{Key: "0"},
			}}}}},
		}
		ret, err := ExpressionEval(exp, testEvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}

	})

	t.Run("with response item found, but no response parent group", func(t *testing.T) {
		exp := types.Expression{Name: "responseHasOnlyKeysOtherThan", Data: []types.ExpressionArg{
			{DType: "str", Str: "weekly.G1.Q1"},
			{DType: "str", Str: "rg.mcg"},
			{DType: "str", Str: "1"},
			{DType: "str", Str: "2"},
		}}
		testEvalContext.Event.Response.Responses = []types.SurveyItemResponse{
			{Key: "weekly.G1.Q1", Response: &types.ResponseItem{Key: "rg", Items: []types.ResponseItem{{Key: "scg", Items: []types.ResponseItem{
				{Key: "0"},
			}}}}},
		}
		ret, err := ExpressionEval(exp, testEvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}

	})

	t.Run("response group does include at least one", func(t *testing.T) {
		exp := types.Expression{Name: "responseHasOnlyKeysOtherThan", Data: []types.ExpressionArg{
			{DType: "str", Str: "weekly.G1.Q1"},
			{DType: "str", Str: "rg.mcg"},
			{DType: "str", Str: "1"},
			{DType: "str", Str: "2"},
		}}
		testEvalContext.Event.Response.Responses = []types.SurveyItemResponse{
			{Key: "weekly.G1.Q1", Response: &types.ResponseItem{Key: "rg", Items: []types.ResponseItem{{Key: "mcg", Items: []types.ResponseItem{
				{Key: "0"},
				{Key: "1"},
				{Key: "3"},
			}}}}},
		}
		ret, err := ExpressionEval(exp, testEvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}

	})

	t.Run("response group is empty", func(t *testing.T) {
		exp := types.Expression{Name: "responseHasOnlyKeysOtherThan", Data: []types.ExpressionArg{
			{DType: "str", Str: "weekly.G1.Q1"},
			{DType: "str", Str: "rg.mcg"},
			{DType: "str", Str: "1"},
			{DType: "str", Str: "2"},
		}}
		testEvalContext.Event.Response.Responses = []types.SurveyItemResponse{
			{Key: "weekly.G1.Q1", Response: &types.ResponseItem{Key: "rg", Items: []types.ResponseItem{{Key: "mcg", Items: []types.ResponseItem{}}}}},
		}
		ret, err := ExpressionEval(exp, testEvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}

	})

	t.Run("response group includes all and other responses", func(t *testing.T) {
		exp := types.Expression{Name: "responseHasOnlyKeysOtherThan", Data: []types.ExpressionArg{
			{DType: "str", Str: "weekly.G1.Q1"},
			{DType: "str", Str: "rg.mcg"},
			{DType: "str", Str: "1"},
			{DType: "str", Str: "2"},
		}}
		testEvalContext.Event.Response.Responses = []types.SurveyItemResponse{
			{Key: "weekly.G1.Q1", Response: &types.ResponseItem{Key: "rg", Items: []types.ResponseItem{{Key: "mcg", Items: []types.ResponseItem{
				{Key: "0"},
				{Key: "1"},
				{Key: "2"},
			}}}}},
		}
		ret, err := ExpressionEval(exp, testEvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}

	})

	t.Run("response group includes none of the options", func(t *testing.T) {
		exp := types.Expression{Name: "responseHasOnlyKeysOtherThan", Data: []types.ExpressionArg{
			{DType: "str", Str: "weekly.G1.Q1"},
			{DType: "str", Str: "rg.mcg"},
			{DType: "str", Str: "1"},
			{DType: "str", Str: "2"},
		}}
		testEvalContext.Event.Response.Responses = []types.SurveyItemResponse{
			{Key: "weekly.G1.Q1", Response: &types.ResponseItem{Key: "rg", Items: []types.ResponseItem{{Key: "mcg", Items: []types.ResponseItem{
				{Key: "0"},
				{Key: "3"},
			}}}}},
		}
		ret, err := ExpressionEval(exp, testEvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if !ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})
}

func TestEvalResponseHasKeysAny(t *testing.T) {
	testEvalContext := EvalContext{
		Event: types.StudyEvent{
			Type: "SUBMIT",
			Response: types.SurveyResponse{
				Key:       "wwekly",
				Responses: []types.SurveyItemResponse{},
			},
		},
	}
	t.Run("no survey item response found", func(t *testing.T) {
		exp := types.Expression{Name: "responseHasKeysAny", Data: []types.ExpressionArg{
			{DType: "str", Str: "weekly.G1.Q1"},
			{DType: "str", Str: "rg.mcg"},
			{DType: "str", Str: "1"},
			{DType: "str", Str: "2"},
		}}
		testEvalContext.Event.Response.Responses = []types.SurveyItemResponse{
			{Key: "weekly.G1.Q2", Response: &types.ResponseItem{Key: "rg", Items: []types.ResponseItem{{Key: "mcg", Items: []types.ResponseItem{
				{Key: "0"},
			}}}}},
		}
		ret, err := ExpressionEval(exp, testEvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}

	})
	t.Run("with response item found, but no response parent group", func(t *testing.T) {
		exp := types.Expression{Name: "responseHasKeysAny", Data: []types.ExpressionArg{
			{DType: "str", Str: "weekly.G1.Q1"},
			{DType: "str", Str: "rg.mcg"},
			{DType: "str", Str: "1"},
			{DType: "str", Str: "2"},
		}}
		testEvalContext.Event.Response.Responses = []types.SurveyItemResponse{
			{Key: "weekly.G1.Q1", Response: &types.ResponseItem{Key: "rg", Items: []types.ResponseItem{{Key: "scg", Items: []types.ResponseItem{
				{Key: "0"},
			}}}}},
		}
		ret, err := ExpressionEval(exp, testEvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}

	})

	t.Run("response group does not include any", func(t *testing.T) {
		exp := types.Expression{Name: "responseHasKeysAny", Data: []types.ExpressionArg{
			{DType: "str", Str: "weekly.G1.Q1"},
			{DType: "str", Str: "rg.mcg"},
			{DType: "str", Str: "1"},
			{DType: "str", Str: "2"},
		}}
		testEvalContext.Event.Response.Responses = []types.SurveyItemResponse{
			{Key: "weekly.G1.Q1", Response: &types.ResponseItem{Key: "rg", Items: []types.ResponseItem{{Key: "mcg", Items: []types.ResponseItem{
				{Key: "0"},
				{Key: "3"},
			}}}}},
		}
		ret, err := ExpressionEval(exp, testEvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}

	})

	t.Run("response group includes all and other responses", func(t *testing.T) {
		exp := types.Expression{Name: "responseHasKeysAny", Data: []types.ExpressionArg{
			{DType: "str", Str: "weekly.G1.Q1"},
			{DType: "str", Str: "rg.mcg"},
			{DType: "str", Str: "1"},
			{DType: "str", Str: "2"},
		}}
		testEvalContext.Event.Response.Responses = []types.SurveyItemResponse{
			{Key: "weekly.G1.Q1", Response: &types.ResponseItem{Key: "rg", Items: []types.ResponseItem{{Key: "mcg", Items: []types.ResponseItem{
				{Key: "0"},
				{Key: "1"},
				{Key: "2"},
			}}}}},
		}
		ret, err := ExpressionEval(exp, testEvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if !ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}

	})
	t.Run("response group includes only of the multiple options", func(t *testing.T) {
		exp := types.Expression{Name: "responseHasKeysAny", Data: []types.ExpressionArg{
			{DType: "str", Str: "weekly.G1.Q1"},
			{DType: "str", Str: "rg.mcg"},
			{DType: "str", Str: "1"},
			{DType: "str", Str: "2"},
		}}
		testEvalContext.Event.Response.Responses = []types.SurveyItemResponse{
			{Key: "weekly.G1.Q1", Response: &types.ResponseItem{Key: "rg", Items: []types.ResponseItem{{Key: "mcg", Items: []types.ResponseItem{
				{Key: "0"},
				{Key: "1"},
			}}}}},
		}
		ret, err := ExpressionEval(exp, testEvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if !ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}

	})

}

func TestEvalLastSubmissionDateOlderThan(t *testing.T) {
	t.Run("with not older", func(t *testing.T) {
		exp := types.Expression{Name: "lastSubmissionDateOlderThan", Data: []types.ExpressionArg{
			{DType: "exp", Exp: &types.Expression{Name: "timestampWithOffset", Data: []types.ExpressionArg{
				{DType: "num", Num: -10},
			}}},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{StudyStatus: types.PARTICIPANT_STUDY_STATUS_ACTIVE,
				LastSubmissions: map[string]int64{
					"s1": time.Now().Unix() - 2,
				},
			},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})

	t.Run("with specific survey is older", func(t *testing.T) {
		exp := types.Expression{Name: "lastSubmissionDateOlderThan", Data: []types.ExpressionArg{
			{DType: "exp", Exp: &types.Expression{Name: "timestampWithOffset", Data: []types.ExpressionArg{
				{DType: "num", Num: -10},
			}}},
			{DType: "str", Str: "s2"},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{StudyStatus: types.PARTICIPANT_STUDY_STATUS_ACTIVE,
				LastSubmissions: map[string]int64{
					"s1": time.Now().Unix() - 2,
					"s2": time.Now().Unix() - 20,
				}},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if !ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})

	t.Run("with only one type of survey is older", func(t *testing.T) {
		exp := types.Expression{Name: "lastSubmissionDateOlderThan", Data: []types.ExpressionArg{
			{DType: "num", Num: 10},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{
				StudyStatus: types.PARTICIPANT_STUDY_STATUS_ACTIVE,
				LastSubmissions: map[string]int64{
					"s1": time.Now().Unix() - 2,
					"s2": time.Now().Unix() - 20,
				},
			},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})

	t.Run("with all types are older", func(t *testing.T) {
		exp := types.Expression{Name: "lastSubmissionDateOlderThan", Data: []types.ExpressionArg{
			{DType: "exp", Exp: &types.Expression{Name: "timestampWithOffset", Data: []types.ExpressionArg{
				{DType: "num", Num: -10},
			}}},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{
				StudyStatus: types.PARTICIPANT_STUDY_STATUS_ACTIVE,
				LastSubmissions: map[string]int64{
					"s1": time.Now().Unix() - 25,
					"s2": time.Now().Unix() - 20,
				},
			},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if !ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})
}

// Comparisons
func TestEvalEq(t *testing.T) {
	t.Run("for eq numbers", func(t *testing.T) {
		exp := types.Expression{Name: "eq", Data: []types.ExpressionArg{
			{DType: "num", Num: 23},
			{DType: "num", Num: 23},
		}}
		EvalContext := EvalContext{
			Event: types.StudyEvent{Type: "TIMER"},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if !ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})

	t.Run("for not equal numbers", func(t *testing.T) {
		exp := types.Expression{Name: "eq", Data: []types.ExpressionArg{
			{DType: "num", Num: 13},
			{DType: "num", Num: 23},
		}}
		EvalContext := EvalContext{
			Event: types.StudyEvent{Type: "enter"},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})

	t.Run("for equal strings", func(t *testing.T) {
		exp := types.Expression{Name: "eq", Data: []types.ExpressionArg{
			{DType: "str", Str: "enter"},
			{DType: "str", Str: "enter"},
		}}
		EvalContext := EvalContext{
			Event: types.StudyEvent{Type: "enter"},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if !ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})

	t.Run("for not equal strings", func(t *testing.T) {
		exp := types.Expression{Name: "eq", Data: []types.ExpressionArg{
			{DType: "str", Str: "enter"},
			{DType: "str", Str: "time..."},
		}}
		EvalContext := EvalContext{
			Event: types.StudyEvent{Type: "enter"},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})
}

func TestEvalLT(t *testing.T) {
	t.Run("2 < 2", func(t *testing.T) {
		exp := types.Expression{Name: "lt", Data: []types.ExpressionArg{
			{DType: "num", Num: 2},
			{DType: "num", Num: 2},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})

	t.Run("2 < 1", func(t *testing.T) {
		exp := types.Expression{Name: "lt", Data: []types.ExpressionArg{
			{DType: "num", Num: 2},
			{DType: "num", Num: 1},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})

	t.Run("1 < 2", func(t *testing.T) {
		exp := types.Expression{Name: "lt", Data: []types.ExpressionArg{
			{DType: "num", Num: 1},
			{DType: "num", Num: 2},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if !ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})

	t.Run("a < b", func(t *testing.T) {
		exp := types.Expression{Name: "lt", Data: []types.ExpressionArg{
			{DType: "str", Str: "a"},
			{DType: "str", Str: "b"},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if !ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})

	t.Run("b < b", func(t *testing.T) {
		exp := types.Expression{Name: "lt", Data: []types.ExpressionArg{
			{DType: "str", Str: "b"},
			{DType: "str", Str: "b"},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})

	t.Run("b < a", func(t *testing.T) {
		exp := types.Expression{Name: "lt", Data: []types.ExpressionArg{
			{DType: "str", Str: "b"},
			{DType: "str", Str: "a"},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})
}

func TestEvalLTE(t *testing.T) {
	t.Run("2 <= 2", func(t *testing.T) {
		exp := types.Expression{Name: "lte", Data: []types.ExpressionArg{
			{DType: "num", Num: 2},
			{DType: "num", Num: 2},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if !ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})

	t.Run("2 <= 1", func(t *testing.T) {
		exp := types.Expression{Name: "lte", Data: []types.ExpressionArg{
			{DType: "num", Num: 2},
			{DType: "num", Num: 1},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})

	t.Run("1 <= 2", func(t *testing.T) {
		exp := types.Expression{Name: "lte", Data: []types.ExpressionArg{
			{DType: "num", Num: 1},
			{DType: "num", Num: 2},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if !ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})

	t.Run("a <= b", func(t *testing.T) {
		exp := types.Expression{Name: "lte", Data: []types.ExpressionArg{
			{DType: "str", Str: "a"},
			{DType: "str", Str: "b"},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if !ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})

	t.Run("b <= b", func(t *testing.T) {
		exp := types.Expression{Name: "lte", Data: []types.ExpressionArg{
			{DType: "str", Str: "b"},
			{DType: "str", Str: "b"},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if !ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})

	t.Run("b <= a", func(t *testing.T) {
		exp := types.Expression{Name: "lte", Data: []types.ExpressionArg{
			{DType: "str", Str: "b"},
			{DType: "str", Str: "a"},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})
}

func TestEvalGT(t *testing.T) {
	t.Run("2 > 2", func(t *testing.T) {
		exp := types.Expression{Name: "gt", Data: []types.ExpressionArg{
			{DType: "num", Num: 2},
			{DType: "num", Num: 2},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})

	t.Run("2 > 1", func(t *testing.T) {
		exp := types.Expression{Name: "gt", Data: []types.ExpressionArg{
			{DType: "num", Num: 2},
			{DType: "num", Num: 1},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if !ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})

	t.Run("1 > 2", func(t *testing.T) {
		exp := types.Expression{Name: "gt", Data: []types.ExpressionArg{
			{DType: "num", Num: 1},
			{DType: "num", Num: 2},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})

	t.Run("a > b", func(t *testing.T) {
		exp := types.Expression{Name: "gt", Data: []types.ExpressionArg{
			{DType: "str", Str: "a"},
			{DType: "str", Str: "b"},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})

	t.Run("b > b", func(t *testing.T) {
		exp := types.Expression{Name: "gt", Data: []types.ExpressionArg{
			{DType: "str", Str: "b"},
			{DType: "str", Str: "b"},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})

	t.Run("b > a", func(t *testing.T) {
		exp := types.Expression{Name: "gt", Data: []types.ExpressionArg{
			{DType: "str", Str: "b"},
			{DType: "str", Str: "a"},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if !ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})
}

func TestEvalGTE(t *testing.T) {
	t.Run("2 >= 2", func(t *testing.T) {
		exp := types.Expression{Name: "gte", Data: []types.ExpressionArg{
			{DType: "num", Num: 2},
			{DType: "num", Num: 2},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if !ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})

	t.Run("2 >= 1", func(t *testing.T) {
		exp := types.Expression{Name: "gte", Data: []types.ExpressionArg{
			{DType: "num", Num: 2},
			{DType: "num", Num: 1},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if !ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})

	t.Run("1 >= 2", func(t *testing.T) {
		exp := types.Expression{Name: "gte", Data: []types.ExpressionArg{
			{DType: "num", Num: 1},
			{DType: "num", Num: 2},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})

	t.Run("a >= b", func(t *testing.T) {
		exp := types.Expression{Name: "gte", Data: []types.ExpressionArg{
			{DType: "str", Str: "a"},
			{DType: "str", Str: "b"},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})

	t.Run("b >= b", func(t *testing.T) {
		exp := types.Expression{Name: "gte", Data: []types.ExpressionArg{
			{DType: "str", Str: "b"},
			{DType: "str", Str: "b"},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if !ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})

	t.Run("b >= a", func(t *testing.T) {
		exp := types.Expression{Name: "gte", Data: []types.ExpressionArg{
			{DType: "str", Str: "b"},
			{DType: "str", Str: "a"},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if !ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})
}

// Logic operators
func TestEvalAND(t *testing.T) {
	t.Run("0 && 0 ", func(t *testing.T) {
		exp := types.Expression{Name: "and", Data: []types.ExpressionArg{
			{DType: "num", Num: 0},
			{DType: "num", Num: 0},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})

	t.Run("1 && 0 ", func(t *testing.T) {
		exp := types.Expression{Name: "and", Data: []types.ExpressionArg{
			{DType: "num", Num: 1},
			{DType: "num", Num: 0},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})

	t.Run("0 && 1 ", func(t *testing.T) {
		exp := types.Expression{Name: "and", Data: []types.ExpressionArg{
			{DType: "num", Num: 0},
			{DType: "num", Num: 1},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})

	t.Run("1 && 1 ", func(t *testing.T) {
		exp := types.Expression{Name: "and", Data: []types.ExpressionArg{
			{DType: "num", Num: 1},
			{DType: "num", Num: 1},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if !ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})
}

func TestEvalOR(t *testing.T) {
	t.Run("0 || 0 ", func(t *testing.T) {
		exp := types.Expression{Name: "or", Data: []types.ExpressionArg{
			{DType: "num", Num: 0},
			{DType: "num", Num: 0},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})

	t.Run("1 || 0 ", func(t *testing.T) {
		exp := types.Expression{Name: "or", Data: []types.ExpressionArg{
			{DType: "num", Num: 1},
			{DType: "num", Num: 0},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if !ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})

	t.Run("0 || 1 ", func(t *testing.T) {
		exp := types.Expression{Name: "or", Data: []types.ExpressionArg{
			{DType: "num", Num: 0},
			{DType: "num", Num: 1},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if !ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})

	t.Run("1 || 1 ", func(t *testing.T) {
		exp := types.Expression{Name: "or", Data: []types.ExpressionArg{
			{DType: "num", Num: 1},
			{DType: "num", Num: 1},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if !ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})
}

func TestEvalNOT(t *testing.T) {
	t.Run("0", func(t *testing.T) {
		exp := types.Expression{Name: "not", Data: []types.ExpressionArg{
			{DType: "num", Num: 0},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if !ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})
	t.Run("1", func(t *testing.T) {
		exp := types.Expression{Name: "not", Data: []types.ExpressionArg{
			{DType: "num", Num: 1},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected value: %b", ret)
		}
	})
}

func TestEvalTimestampWithOffset(t *testing.T) {
	t.Run("T + 0", func(t *testing.T) {
		exp := types.Expression{Name: "timestampWithOffset", Data: []types.ExpressionArg{
			{DType: "num", Num: 0},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		resTS := int64(ret.(float64))
		if resTS > time.Now().Unix()+1 || resTS < time.Now().Unix()-1 {
			t.Errorf("unexpected value: %d - expected ca. %d", ret, time.Now().Unix()+0)
		}
	})

	t.Run("T + 10", func(t *testing.T) {
		exp := types.Expression{Name: "timestampWithOffset", Data: []types.ExpressionArg{
			{DType: "num", Num: 10},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		resTS := int64(ret.(float64))
		if resTS > time.Now().Unix()+11 || resTS < time.Now().Unix()+9 {
			t.Errorf("unexpected value: %d - expected ca. %d", ret, time.Now().Unix()+10)
		}
	})

	t.Run("T - 10", func(t *testing.T) {
		exp := types.Expression{Name: "timestampWithOffset", Data: []types.ExpressionArg{
			{DType: "num", Num: -10},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		resTS := int64(ret.(float64))
		if resTS < time.Now().Unix()-11 || resTS > time.Now().Unix()-9 {
			t.Errorf("unexpected value: %d - expected ca. %d", ret, time.Now().Unix()-10)
		}
	})

	t.Run("T + No num", func(t *testing.T) {
		exp := types.Expression{Name: "timestampWithOffset", Data: []types.ExpressionArg{
			{DType: "str", Str: "0"},
		}}
		EvalContext := EvalContext{}
		_, err := ExpressionEval(exp, EvalContext)
		if err == nil {
			t.Errorf("unexpected lack of error: parameter 1 was not num")
			return
		}
	})

	t.Run("R + 0", func(t *testing.T) {
		r := time.Now().Unix() - 31536000
		exp := types.Expression{Name: "timestampWithOffset", Data: []types.ExpressionArg{
			{DType: "num", Num: 0},
			{DType: "num", Num: float64(r)},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		resTS := int64(ret.(float64))
		if resTS > r+1 || resTS < r-1 {
			t.Errorf("unexpected value: %d - expected ca. %d", ret, r+0)
		}
	})

	t.Run("R + 10", func(t *testing.T) {
		r := time.Now().Unix() - 31536000
		exp := types.Expression{Name: "timestampWithOffset", Data: []types.ExpressionArg{
			{DType: "num", Num: 10},
			{DType: "num", Num: float64(r)},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		resTS := int64(ret.(float64))
		if resTS > r+11 || resTS < r+9 {
			t.Errorf("unexpected value: %d - expected ca. %d", ret, r+10)
		}
	})

	t.Run("R - 10", func(t *testing.T) {
		r := time.Now().Unix() - 31536000
		exp := types.Expression{Name: "timestampWithOffset", Data: []types.ExpressionArg{
			{DType: "num", Num: -10},
			{DType: "num", Num: float64(r)},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		resTS := int64(ret.(float64))
		if resTS > r-9 || resTS < r-11 {
			t.Errorf("unexpected value: %d - expected ca. %d", ret, r-10)
		}
	})

	t.Run("R + No num", func(t *testing.T) {
		r := time.Now().Unix() - 31536000
		exp := types.Expression{Name: "timestampWithOffset", Data: []types.ExpressionArg{
			{DType: "str", Str: "0"},
			{DType: "num", Num: float64(r)},
		}}
		EvalContext := EvalContext{}
		_, err := ExpressionEval(exp, EvalContext)
		if err == nil {
			t.Errorf("unexpected lack of error: parameter 1 was not num")
			return
		}
	})

	t.Run("No num + 10", func(t *testing.T) {
		exp := types.Expression{Name: "timestampWithOffset", Data: []types.ExpressionArg{
			{DType: "num", Num: 10},
			{DType: "str", Str: "1"},
		}}
		EvalContext := EvalContext{}
		_, err := ExpressionEval(exp, EvalContext)
		if err == nil {
			t.Errorf("unexpected lack of error: parameter 2 was not num")
			return
		}
	})

	t.Run("No num + No num", func(t *testing.T) {
		exp := types.Expression{Name: "timestampWithOffset", Data: []types.ExpressionArg{
			{DType: "str", Str: "0"},
			{DType: "str", Str: "1"},
		}}
		EvalContext := EvalContext{}
		_, err := ExpressionEval(exp, EvalContext)
		if err == nil {
			t.Errorf("unexpected lack of error: parameters 1 & 2 were not num")
			return
		}
	})

	t.Run("Valid Exp", func(t *testing.T) {
		exp := types.Expression{Name: "timestampWithOffset", Data: []types.ExpressionArg{
			{
				DType: "exp", Exp: &types.Expression{
					Name: "timestampWithOffset", Data: []types.ExpressionArg{
						{DType: "num", Num: -float64(time.Now().Unix())},
					}},
			}}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		resTS := int64(ret.(float64))
		if resTS-time.Now().Unix() > 1 {
			t.Errorf("unexpected value: %d, expected %d", resTS, time.Now().Unix())
		}
	})

	t.Run("Valid Exp + Valid Exp", func(t *testing.T) {
		exp := types.Expression{Name: "timestampWithOffset", Data: []types.ExpressionArg{
			{DType: "exp", Exp: &types.Expression{
				Name: "timestampWithOffset", Data: []types.ExpressionArg{
					{DType: "num", Num: -float64(time.Now().Unix())},
				}},
			},
			{DType: "exp", Exp: &types.Expression{
				Name: "timestampWithOffset", Data: []types.ExpressionArg{
					{DType: "num", Num: -float64(time.Now().Unix())},
				}},
			},
		}}
		EvalContext := EvalContext{}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		resTS := int64(ret.(float64))
		if resTS > 1 {
			t.Errorf("unexpected value: %d, expected %d", resTS, 0)
		}
	})

	t.Run("Not Valid Exp + Valid Exp", func(t *testing.T) {
		exp := types.Expression{Name: "timestampWithOffset", Data: []types.ExpressionArg{
			{DType: "exp", Exp: &types.Expression{
				Name: "or", Data: []types.ExpressionArg{
					{DType: "num", Num: 1},
					{DType: "num", Num: 1},
				}},
			},
			{DType: "exp", Exp: &types.Expression{
				Name: "timestampWithOffset", Data: []types.ExpressionArg{
					{DType: "num", Num: -float64(time.Now().Unix())},
				}},
			},
		}}
		EvalContext := EvalContext{}
		_, err := ExpressionEval(exp, EvalContext)
		if err == nil {
			t.Errorf("unexpected lack of error")
			return
		}
	})

	t.Run("Valid Exp + Not Valid Exp", func(t *testing.T) {
		exp := types.Expression{Name: "timestampWithOffset", Data: []types.ExpressionArg{
			{DType: "exp", Exp: &types.Expression{
				Name: "timestampWithOffset", Data: []types.ExpressionArg{
					{DType: "num", Num: -float64(time.Now().Unix())},
				}},
			},
			{DType: "exp", Exp: &types.Expression{
				Name: "or", Data: []types.ExpressionArg{
					{DType: "num", Num: 1},
					{DType: "num", Num: 1},
				}},
			},
		}}
		EvalContext := EvalContext{}
		_, err := ExpressionEval(exp, EvalContext)
		if err == nil {
			t.Errorf("unexpected lack of error")
			return
		}
	})
}

func TestEvalHasMessageTypeAssigned(t *testing.T) {
	t.Run("participant has no messages", func(t *testing.T) {
		exp := types.Expression{Name: "hasMessageTypeAssigned", Data: []types.ExpressionArg{
			{DType: "str", Str: "testMessage"},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{
				Messages: []types.ParticipantMessage{},
			},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		resTS := ret.(bool)
		if resTS {
			t.Errorf("unexpected value: %v", ret)
		}
	})

	t.Run("participant has messages but none that are looked for", func(t *testing.T) {
		exp := types.Expression{Name: "hasMessageTypeAssigned", Data: []types.ExpressionArg{
			{DType: "str", Str: "testMessage"},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{
				Messages: []types.ParticipantMessage{
					{Type: "testMessage2", ScheduledFor: 100},
				},
			},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		resTS := ret.(bool)
		if resTS {
			t.Errorf("unexpected value: %v", ret)
		}
	})

	t.Run("participant has messages and one is the one looked for", func(t *testing.T) {
		exp := types.Expression{Name: "hasMessageTypeAssigned", Data: []types.ExpressionArg{
			{DType: "str", Str: "testMessage"},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{
				Messages: []types.ParticipantMessage{
					{Type: "testMessage2", ScheduledFor: 100},
					{Type: "testMessage", ScheduledFor: 200},
				},
			},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		resTS := ret.(bool)
		if !resTS {
			t.Errorf("unexpected value: %v", ret)
		}
	})
}

func TestEvalGetMessageNextTime(t *testing.T) {
	t.Run("participant has no messages", func(t *testing.T) {
		exp := types.Expression{Name: "getMessageNextTime", Data: []types.ExpressionArg{
			{DType: "str", Str: "testMessage"},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err == nil {
			t.Error("should return error")
			return
		}
		resTS := ret.(int64)
		if resTS != 0 {
			t.Errorf("unexpected value: %d", ret)
		}
	})

	t.Run("participant has messages but none that are looked for", func(t *testing.T) {
		exp := types.Expression{Name: "getMessageNextTime", Data: []types.ExpressionArg{
			{DType: "str", Str: "testMessage"},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{
				Messages: []types.ParticipantMessage{
					{Type: "testMessage2", ScheduledFor: 100},
				},
			},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err == nil {
			t.Error("should return error")
			return
		}
		resTS := ret.(int64)
		if resTS != 0 {
			t.Errorf("unexpected value: %d", ret)
		}
	})

	t.Run("participant has messages and one is the one looked for", func(t *testing.T) {
		exp := types.Expression{Name: "getMessageNextTime", Data: []types.ExpressionArg{
			{DType: "str", Str: "testMessage"},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{
				Messages: []types.ParticipantMessage{
					{Type: "testMessage2", ScheduledFor: 50},
					{Type: "testMessage", ScheduledFor: 100},
				},
			},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		resTS := ret.(int64)
		if resTS != 100 {
			t.Errorf("unexpected value: %d", ret)
		}
	})

	t.Run("participant has messages and two from the specified type", func(t *testing.T) {
		exp := types.Expression{Name: "getMessageNextTime", Data: []types.ExpressionArg{
			{DType: "str", Str: "testMessage"},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{
				Messages: []types.ParticipantMessage{
					{Type: "testMessage1", ScheduledFor: 100},
					{Type: "testMessage", ScheduledFor: 200},
					{Type: "testMessage", ScheduledFor: 400},
				},
			},
		}
		ret, err := ExpressionEval(exp, EvalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		resTS := ret.(int64)
		if resTS != 200 {
			t.Errorf("unexpected value: %d", ret)
		}
	})
}
