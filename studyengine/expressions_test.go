package studyengine

import (
	"testing"

	"github.com/influenzanet/study-service/models"
)

func TestEvalGetEventType(t *testing.T) {
	exp := models.Expression{Name: "getEventType"}

	t.Run("for enter", func(t *testing.T) {
		evalContext := evalContext{
			event: models.StudyEvent{Type: "enter"},
		}
		ret, err := ExpressionEval(exp, evalContext)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if ret.(string) != "enter" {
			t.Errorf("unexpected type or str value: %s", ret)
		}
	})
}

func TestEvalEq(t *testing.T) {
	t.Run("for eq numbers", func(t *testing.T) {
		exp := models.Expression{Name: "eq", Data: []models.ExpressionArg{
			models.ExpressionArg{DType: "num", Num: 23},
			models.ExpressionArg{DType: "num", Num: 23},
		}}
		evalContext := evalContext{
			event: models.StudyEvent{Type: "TIMER"},
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
		exp := models.Expression{Name: "eq", Data: []models.ExpressionArg{
			models.ExpressionArg{DType: "num", Num: 13},
			models.ExpressionArg{DType: "num", Num: 23},
		}}
		evalContext := evalContext{
			event: models.StudyEvent{Type: "enter"},
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
		exp := models.Expression{Name: "eq", Data: []models.ExpressionArg{
			models.ExpressionArg{DType: "exp", Exp: models.Expression{Name: "getEventType"}},
			models.ExpressionArg{DType: "str", Str: "enter"},
		}}
		evalContext := evalContext{
			event: models.StudyEvent{Type: "enter"},
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
		exp := models.Expression{Name: "eq", Data: []models.ExpressionArg{
			models.ExpressionArg{DType: "exp", Exp: models.Expression{Name: "getEventType"}},
			models.ExpressionArg{DType: "str", Str: "time..."},
		}}
		evalContext := evalContext{
			event: models.StudyEvent{Type: "enter"},
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
