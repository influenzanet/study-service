package studyengine

import (
	"testing"
	"time"

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
			{DType: "str", Str: "exited"},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{StudyStatus: "active"},
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
			{DType: "str", Str: "active"},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{StudyStatus: "active"},
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
			{DType: "num", Num: 10},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{StudyStatus: "active",
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
			{DType: "num", Num: 10},
			{DType: "str", Str: "s2"},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{StudyStatus: "active",
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
				StudyStatus: "active",
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
			{DType: "num", Num: 10},
		}}
		EvalContext := EvalContext{
			ParticipantState: types.ParticipantState{
				StudyStatus: "active",
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
}
