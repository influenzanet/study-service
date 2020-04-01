package studyengine

import (
	"errors"

	"github.com/influenzanet/study-service/models"
)

type evalContext struct {
	event            models.StudyEvent
	participantState models.ParticipantState
}

func ExpressionEval(expression models.Expression, evalCtx evalContext) (val interface{}, err error) {
	switch expression.Name {
	case "eq":
		val, err = evalCtx.eq(expression)
	case "getEventType":
		val = evalCtx.getEventType()
	default:
		err = errors.New("expression name not known")
		return
	}
	return
}

// getEventType returns type from the event struct
func (ctx evalContext) getEventType() string {
	return ctx.event.Type
}

func (ctx evalContext) eq(exp models.Expression) (val bool, err error) {
	if len(exp.Data) != 2 {
		return val, errors.New("not expected numbers of arguments")
	}

	arg1 := exp.Data[0]
	if arg1.IsExpression() {
		arg1Val, err := ExpressionEval(arg1.Exp, ctx)
		strVal, ok := arg1Val.(string)
		if err != nil || !ok {
			return val, err
		}
		arg1.DType = "str"
		arg1.Str = strVal
	}

	arg2 := exp.Data[1]
	if arg2.IsExpression() {
		arg2Val, err := ExpressionEval(arg2.Exp, ctx)
		strVal, ok := arg2Val.(string)
		if err != nil || !ok {
			return val, err
		}
		arg2.DType = "str"
		arg2.Str = strVal
	}

	if arg1.DType != arg2.DType {
		return val, errors.New("data type of arguments don't match")
	}

	switch arg1.DType {
	case "num":
		return arg1.Num == arg2.Num, nil
	case "str":
		return arg1.Str == arg2.Str, nil
	}
	return
}
