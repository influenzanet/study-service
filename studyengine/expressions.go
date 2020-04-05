package studyengine

import (
	"errors"
	"fmt"
	"strings"

	"github.com/influenzanet/study-service/models"
)

// evalContext contains all the data that can be looked up by expressions
type evalContext struct {
	event            models.StudyEvent
	participantState models.ParticipantState
}

func ExpressionEval(expression models.Expression, evalCtx evalContext) (val interface{}, err error) {
	switch expression.Name {
	case "checkEventType":
		val, err = evalCtx.checkEventType(expression)
	case "eq":
		val, err = evalCtx.eq(expression)
	case "lt":
		val, err = evalCtx.lt(expression)
	case "lte":
		val, err = evalCtx.lte(expression)
	case "gt":
		val, err = evalCtx.gt(expression)
	case "gte":
		val, err = evalCtx.gte(expression)
	case "and":
		val, err = evalCtx.and(expression)
	case "or":
		val, err = evalCtx.or(expression)
	case "not":
		val, err = evalCtx.not(expression)
	default:
		err = fmt.Errorf("expression name not known: %s", expression.Name)
		return
	}
	return
}

func (ctx evalContext) expressionArgResolver(arg models.ExpressionArg) (interface{}, error) {
	switch arg.DType {
	case "num":
		return arg.Num, nil
	case "exp":
		return ExpressionEval(arg.Exp, ctx)
	case "str":
		return arg.Str, nil
	default:
		return arg.Str, nil
	}
}

// checkEventType compares the eventType with a string
func (ctx evalContext) checkEventType(exp models.Expression) (val bool, err error) {
	if len(exp.Data) != 1 {
		return val, errors.New("unexpected numbers of arguments")
	}

	arg1, err := ctx.expressionArgResolver(exp.Data[0])
	if err != nil {
		return val, err
	}
	arg1Val, ok := arg1.(string)
	if !ok {
		return val, errors.New("could not cast arguments")
	}

	return ctx.event.Type == arg1Val, nil
}

func (ctx evalContext) eq(exp models.Expression) (val bool, err error) {
	if len(exp.Data) != 2 {
		return val, errors.New("not expected numbers of arguments")
	}

	arg1, err := ctx.expressionArgResolver(exp.Data[0])
	if err != nil {
		return val, err
	}
	arg2, err := ctx.expressionArgResolver(exp.Data[1])
	if err != nil {
		return val, err
	}

	switch arg1Val := arg1.(type) {
	case string:
		arg2Val, ok2 := arg2.(string)
		if !ok2 {
			return val, errors.New("could not cast arguments")
		}
		return strings.Compare(arg1Val, arg2Val) == 0, nil
	case float64:
		arg2Val, ok2 := arg2.(float64)
		if !ok2 {
			return val, errors.New("could not cast arguments")
		}
		return arg1Val == arg2Val, nil
	default:
		return val, fmt.Errorf("I don't know about type %T", arg1Val)
	}
}

func (ctx evalContext) lt(exp models.Expression) (val bool, err error) {
	if len(exp.Data) != 2 {
		return val, errors.New("not expected numbers of arguments")
	}

	arg1, err := ctx.expressionArgResolver(exp.Data[0])
	if err != nil {
		return val, err
	}
	arg2, err := ctx.expressionArgResolver(exp.Data[1])
	if err != nil {
		return val, err
	}

	switch arg1Val := arg1.(type) {
	case string:
		arg2Val, ok2 := arg2.(string)
		if !ok2 {
			return val, errors.New("could not cast arguments")
		}
		return strings.Compare(arg1Val, arg2Val) == -1, nil
	case float64:
		arg2Val, ok2 := arg2.(float64)
		if !ok2 {
			return val, errors.New("could not cast arguments")
		}
		return arg1Val < arg2Val, nil
	default:
		return val, fmt.Errorf("I don't know about type %T", arg1Val)
	}
}

func (ctx evalContext) lte(exp models.Expression) (val bool, err error) {
	if len(exp.Data) != 2 {
		return val, errors.New("not expected numbers of arguments")
	}

	arg1, err := ctx.expressionArgResolver(exp.Data[0])
	if err != nil {
		return val, err
	}
	arg2, err := ctx.expressionArgResolver(exp.Data[1])
	if err != nil {
		return val, err
	}

	switch arg1Val := arg1.(type) {
	case string:
		arg2Val, ok2 := arg2.(string)
		if !ok2 {
			return val, errors.New("could not cast arguments")
		}
		return strings.Compare(arg1Val, arg2Val) <= 0, nil
	case float64:
		arg2Val, ok2 := arg2.(float64)
		if !ok2 {
			return val, errors.New("could not cast arguments")
		}
		return arg1Val <= arg2Val, nil
	default:
		return val, fmt.Errorf("I don't know about type %T", arg1Val)
	}
}

func (ctx evalContext) gt(exp models.Expression) (val bool, err error) {
	if len(exp.Data) != 2 {
		return val, errors.New("not expected numbers of arguments")
	}

	arg1, err := ctx.expressionArgResolver(exp.Data[0])
	if err != nil {
		return val, err
	}
	arg2, err := ctx.expressionArgResolver(exp.Data[1])
	if err != nil {
		return val, err
	}

	switch arg1Val := arg1.(type) {
	case string:
		arg2Val, ok2 := arg2.(string)
		if !ok2 {
			return val, errors.New("could not cast arguments")
		}
		return strings.Compare(arg1Val, arg2Val) == 1, nil
	case float64:
		arg2Val, ok2 := arg2.(float64)
		if !ok2 {
			return val, errors.New("could not cast arguments")
		}
		return arg1Val > arg2Val, nil
	default:
		return val, fmt.Errorf("I don't know about type %T", arg1Val)
	}
}

func (ctx evalContext) gte(exp models.Expression) (val bool, err error) {
	if len(exp.Data) != 2 {
		return val, errors.New("not expected numbers of arguments")
	}

	arg1, err := ctx.expressionArgResolver(exp.Data[0])
	if err != nil {
		return val, err
	}
	arg2, err := ctx.expressionArgResolver(exp.Data[1])
	if err != nil {
		return val, err
	}

	switch arg1Val := arg1.(type) {
	case string:
		arg2Val, ok2 := arg2.(string)
		if !ok2 {
			return val, errors.New("could not cast arguments")
		}
		return strings.Compare(arg1Val, arg2Val) >= 0, nil
	case float64:
		arg2Val, ok2 := arg2.(float64)
		if !ok2 {
			return val, errors.New("could not cast arguments")
		}
		return arg1Val >= arg2Val, nil
	default:
		return val, fmt.Errorf("I don't know about type %T", arg1Val)
	}
}

func (ctx evalContext) and(exp models.Expression) (val bool, err error) {
	if len(exp.Data) < 2 {
		return val, errors.New("should have at least two arguments")
	}

	for _, d := range exp.Data {
		arg1, err := ctx.expressionArgResolver(d)
		if err != nil {
			return val, err
		}
		switch arg1Val := arg1.(type) {
		case bool:
			if !arg1Val {
				return false, nil
			}
		case float64:
			if arg1Val == 0 {
				return false, nil
			}
		}
	}
	return true, nil
}

func (ctx evalContext) or(exp models.Expression) (val bool, err error) {
	if len(exp.Data) < 2 {
		return val, errors.New("should have at least two arguments")
	}

	for _, d := range exp.Data {
		arg1, err := ctx.expressionArgResolver(d)
		if err != nil {
			return val, err
		}
		switch arg1Val := arg1.(type) {
		case bool:
			if arg1Val {
				return true, nil
			}
		case float64:
			if arg1Val > 0 {
				return true, nil
			}
		}
	}
	return false, nil
}

func (ctx evalContext) not(exp models.Expression) (val bool, err error) {
	if len(exp.Data) != 1 {
		return val, errors.New("should have one argument")
	}

	arg1, err := ctx.expressionArgResolver(exp.Data[0])
	if err != nil {
		return val, err
	}
	switch arg1Val := arg1.(type) {
	case bool:
		return !arg1Val, nil
	case float64:
		if arg1Val == 0 {
			return true, nil
		}
		return false, nil
	}
	return
}
