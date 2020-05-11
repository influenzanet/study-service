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
		evalContext := evalContext{
			event: types.StudyEvent{Type: "ENTER"},
		}
		ret, err := ExpressionEval(exp, evalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if !ret.(bool) {
			t.Errorf("unexpected type or value: %s", ret)
		}
	})

	t.Run("for not matching", func(t *testing.T) {
		evalContext := evalContext{
			event: types.StudyEvent{Type: "enter"},
		}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{
			event: types.StudyEvent{Type: "SUBMIT"},
		}
		ret, err := ExpressionEval(exp, evalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected type or value: %s", ret)
		}
	})

	t.Run("not matching key", func(t *testing.T) {
		evalContext := evalContext{
			event: types.StudyEvent{
				Type: "SUBMIT",
				Response: types.SurveyResponse{
					Key:       "intake",
					Responses: []types.SurveyItemResponse{},
				},
			},
		}
		ret, err := ExpressionEval(exp, evalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(bool) {
			t.Errorf("unexpected type or value: %s", ret)
		}
	})

	t.Run("for matching key", func(t *testing.T) {
		evalContext := evalContext{
			event: types.StudyEvent{
				Type: "SUBMIT",
				Response: types.SurveyResponse{
					Key:       "weekly",
					Responses: []types.SurveyItemResponse{},
				},
			},
		}
		ret, err := ExpressionEval(exp, evalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if !ret.(bool) {
			t.Errorf("unexpected type or value: %s", ret)
		}
	})
}

func TestEvalGetParticipantState(t *testing.T) {
	t.Run("with normal state", func(t *testing.T) { t.Error("test unimplemented") })
}

// Comparisons
func TestEvalEq(t *testing.T) {
	t.Run("for eq numbers", func(t *testing.T) {
		exp := types.Expression{Name: "eq", Data: []types.ExpressionArg{
			{DType: "num", Num: 23},
			{DType: "num", Num: 23},
		}}
		evalContext := evalContext{
			event: types.StudyEvent{Type: "TIMER"},
		}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{
			event: types.StudyEvent{Type: "enter"},
		}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{
			event: types.StudyEvent{Type: "enter"},
		}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{
			event: types.StudyEvent{Type: "enter"},
		}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
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
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(int64) > time.Now().Unix()+11 || ret.(int64) < time.Now().Unix()+9 {
			t.Errorf("unexpected value: %d - expected ca. %d", ret, time.Now().Unix()+10)
		}
	})
	t.Run("T - 10", func(t *testing.T) {
		exp := types.Expression{Name: "timestampWithOffset", Data: []types.ExpressionArg{
			{DType: "num", Num: -10},
		}}
		evalContext := evalContext{}
		ret, err := ExpressionEval(exp, evalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(int64) < time.Now().Unix()-11 || ret.(int64) > time.Now().Unix()-9 {
			t.Errorf("unexpected value: %d - expected ca. %d", ret, time.Now().Unix()-10)
		}
	})
}
