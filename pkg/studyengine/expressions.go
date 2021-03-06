package studyengine

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/influenzanet/study-service/pkg/types"
)

// EvalContext contains all the data that can be looked up by expressions
type EvalContext struct {
	Event            types.StudyEvent
	ParticipantState types.ParticipantState
}

func ExpressionEval(expression types.Expression, evalCtx EvalContext) (val interface{}, err error) {
	switch expression.Name {
	case "checkEventType":
		val, err = evalCtx.checkEventType(expression)
	case "checkSurveyResponseKey":
		val, err = evalCtx.checkSurveyResponseKey(expression)
	case "hasStudyStatus":
		val, err = evalCtx.hasStudyStatus(expression)
	case "lastSubmissionDateOlderThan":
		val, err = evalCtx.lastSubmissionDateOlderThan(expression)
	case "responseHasKeysAny":
		val, err = evalCtx.responseHasKeysAny(expression)
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
	case "timestampWithOffset":
		val, err = evalCtx.timestampWithOffset(expression)
	default:
		err = fmt.Errorf("expression name not known: %s", expression.Name)
		return
	}
	return
}

func (ctx EvalContext) expressionArgResolver(arg types.ExpressionArg) (interface{}, error) {
	switch arg.DType {
	case "num":
		return arg.Num, nil
	case "exp":
		return ExpressionEval(*arg.Exp, ctx)
	case "str":
		return arg.Str, nil
	default:
		return arg.Str, nil
	}
}

// checkEventType compares the eventType with a string
func (ctx EvalContext) checkEventType(exp types.Expression) (val bool, err error) {
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

	return ctx.Event.Type == arg1Val, nil
}

// checkSurveyResponseKey compares the key of the submitted survey response (if any)
func (ctx EvalContext) checkSurveyResponseKey(exp types.Expression) (val bool, err error) {
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

	return ctx.Event.Response.Key == arg1Val, nil
}

func (ctx EvalContext) hasStudyStatus(exp types.Expression) (val bool, err error) {
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

	return ctx.ParticipantState.StudyStatus == arg1Val, nil
}

func (ctx EvalContext) lastSubmissionDateOlderThan(exp types.Expression) (val bool, err error) {
	if len(exp.Data) != 1 && len(exp.Data) != 2 {
		return val, errors.New("unexpected numbers of arguments")
	}

	arg1, err := ctx.expressionArgResolver(exp.Data[0])
	if err != nil {
		return val, err
	}
	arg1Val, ok := arg1.(float64)
	if !ok {
		return val, errors.New("could not cast argument 1")
	}

	refTime := time.Now().Unix() - int64(arg1Val)
	if len(exp.Data) == 2 {
		arg2, err := ctx.expressionArgResolver(exp.Data[1])
		if err != nil {
			return val, err
		}
		arg2Val, ok := arg2.(string)
		if !ok {
			return val, errors.New("could not cast arguments")
		}
		lastTs, ok := ctx.ParticipantState.LastSubmissions[arg2Val]
		if !ok {
			return false, nil
		}
		return lastTs < refTime, nil

	} else {
		for _, lastTs := range ctx.ParticipantState.LastSubmissions {
			if lastTs > refTime {
				return false, nil
			}
		}
	}
	return true, nil
}

func (ctx EvalContext) responseHasKeysAny(exp types.Expression) (val bool, err error) {
	if len(exp.Data) < 3 {
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
	arg2, err := ctx.expressionArgResolver(exp.Data[1])
	if err != nil {
		return val, err
	}
	arg2Val, ok := arg2.(string)
	if !ok {
		return val, errors.New("could not cast arguments")
	}

	targetKeys := []string{}
	for _, d := range exp.Data[2:] {
		arg, err := ctx.expressionArgResolver(d)
		if err != nil {
			return val, err
		}
		argVal, ok := arg.(string)
		if !ok {
			return val, errors.New("could not cast arguments")
		}
		targetKeys = append(targetKeys, argVal)
	}

	// find survey item:
	responseOfInterest := &types.SurveyItemResponse{}
	for _, response := range ctx.Event.Response.Responses {
		if response.Key == arg1Val {
			responseOfInterest = &response
			break
		}
	}
	if responseOfInterest.Key == "" || responseOfInterest.Response == nil {
		// Item not found
		return false, nil
	}
	responseParentGroup := &types.ResponseItem{}
	for i, k := range strings.Split(arg2Val, ".") {
		if i == 0 {
			if responseOfInterest.Response.Key != k {
				// item not found:
				return false, nil
			}
			responseParentGroup = responseOfInterest.Response
			continue
		}

		found := false
		for _, item := range responseParentGroup.Items {
			if item.Key == k {
				found = true
				responseParentGroup = &item
				break
			}
		}
		if !found {
			// item not found:
			return false, nil
		}
	}

	// Check if any of the target in response
	anyFound := false
	for _, target := range targetKeys {
		for _, item := range responseParentGroup.Items {
			if item.Key == target {
				anyFound = true
				break
			}
		}
		if anyFound {
			break
		}
	}
	return anyFound, nil
}

func (ctx EvalContext) eq(exp types.Expression) (val bool, err error) {
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

func (ctx EvalContext) lt(exp types.Expression) (val bool, err error) {
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

func (ctx EvalContext) lte(exp types.Expression) (val bool, err error) {
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

func (ctx EvalContext) gt(exp types.Expression) (val bool, err error) {
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

func (ctx EvalContext) gte(exp types.Expression) (val bool, err error) {
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

func (ctx EvalContext) and(exp types.Expression) (val bool, err error) {
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

func (ctx EvalContext) or(exp types.Expression) (val bool, err error) {
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

func (ctx EvalContext) not(exp types.Expression) (val bool, err error) {
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

func (ctx EvalContext) timestampWithOffset(exp types.Expression) (t float64, err error) {
	if len(exp.Data) != 1 {
		return t, errors.New("should have one argument")
	}

	arg1, err := ctx.expressionArgResolver(exp.Data[0])
	if err != nil {
		return t, err
	}
	delta := int64(arg1.(float64))
	t = float64(time.Now().Unix() + delta)
	return
}
