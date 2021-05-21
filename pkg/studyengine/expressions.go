package studyengine

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/influenzanet/study-service/pkg/dbs/studydb"
	"github.com/influenzanet/study-service/pkg/types"
)

// EvalContext contains all the data that can be looked up by expressions
type EvalContext struct {
	Event            types.StudyEvent
	ParticipantState types.ParticipantState
	DbService        *studydb.StudyDBService
}

func ExpressionEval(expression types.Expression, evalCtx EvalContext) (val interface{}, err error) {
	switch expression.Name {
	case "checkEventType":
		val, err = evalCtx.checkEventType(expression)
	// Response checkers:
	case "checkSurveyResponseKey":
		val, err = evalCtx.checkSurveyResponseKey(expression)
	case "responseHasKeysAny":
		val, err = evalCtx.responseHasKeysAny(expression)
	case "responseHasOnlyKeysOtherThan":
		val, err = evalCtx.responseHasOnlyKeysOtherThan(expression)
	case "getResponseValueAsNum":
		val, err = evalCtx.getResponseValueAsNum(expression)
	case "getResponseValueAsStr":
		val, err = evalCtx.getResponseValueAsStr(expression)
	// Old responses:
	case "checkConditionForOldResponses":
		val, err = evalCtx.checkConditionForOldResponses(expression)
	// Participant state:
	case "getStudyEntryTime":
		val, err = evalCtx.getStudyEntryTime(expression)
	case "hasSurveyKeyAssigned":
		val, err = evalCtx.hasSurveyKeyAssigned(expression)
	case "getSurveyKeyAssignedFrom":
		val, err = evalCtx.getSurveyKeyAssignedFrom(expression)
	case "getSurveyKeyAssignedUntil":
		val, err = evalCtx.getSurveyKeyAssignedUntil(expression)
	case "hasStudyStatus":
		val, err = evalCtx.hasStudyStatus(expression)
	case "hasParticipantFlag":
		val, err = evalCtx.hasParticipantFlag(expression)
	case "lastSubmissionDateOlderThan":
		val, err = evalCtx.lastSubmissionDateOlderThan(expression)
	// Logical and comparisions:
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
	// Other
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

func (ctx EvalContext) checkConditionForOldResponses(exp types.Expression) (val bool, err error) {
	if ctx.DbService == nil {
		return val, errors.New("checkConditionForOldResponses: DB connection not available")
	}
	if len(exp.Data) < 3 {
		return val, fmt.Errorf("checkConditionForOldResponses: unexpected numbers of arguments: %d", len(exp.Data))
	}

	oldEvalContext := EvalContext{
		ParticipantState: ctx.ParticipantState,
		Event: types.StudyEvent{
			Response: types.SurveyResponse{},
		},
	}
	ExpressionEval()
}

func (ctx EvalContext) getStudyEntryTime(exp types.Expression) (t float64, err error) {
	return float64(ctx.ParticipantState.EnteredAt), nil
}

func (ctx EvalContext) hasSurveyKeyAssigned(exp types.Expression) (val bool, err error) {
	if len(exp.Data) != 1 || !exp.Data[0].IsString() {
		return val, errors.New("unexpected number or wrong type of argument")
	}
	arg1, err := ctx.expressionArgResolver(exp.Data[0])
	if err != nil {
		return val, err
	}
	arg1Val, ok := arg1.(string)
	if !ok {
		return val, errors.New("could not cast argument")
	}

	for _, survey := range ctx.ParticipantState.AssignedSurveys {
		if survey.SurveyKey == arg1Val {
			val = true
			return
		}
	}
	return
}

func (ctx EvalContext) getSurveyKeyAssignedFrom(exp types.Expression) (val float64, err error) {
	if len(exp.Data) != 1 || !exp.Data[0].IsString() {
		return val, errors.New("unexpected number or wrong type of argument")
	}
	arg1, err := ctx.expressionArgResolver(exp.Data[0])
	if err != nil {
		return val, err
	}
	arg1Val, ok := arg1.(string)
	if !ok {
		return val, errors.New("could not cast argument")
	}

	for _, survey := range ctx.ParticipantState.AssignedSurveys {
		if survey.SurveyKey == arg1Val {
			val = float64(survey.ValidFrom)
			return
		}
	}

	return -1, nil
}

func (ctx EvalContext) getSurveyKeyAssignedUntil(exp types.Expression) (val float64, err error) {
	if len(exp.Data) != 1 || !exp.Data[0].IsString() {
		return val, errors.New("unexpected number or wrong type of argument")
	}
	arg1, err := ctx.expressionArgResolver(exp.Data[0])
	if err != nil {
		return val, err
	}
	arg1Val, ok := arg1.(string)
	if !ok {
		return val, errors.New("could not cast argument")
	}

	for _, survey := range ctx.ParticipantState.AssignedSurveys {
		if survey.SurveyKey == arg1Val {
			val = float64(survey.ValidUntil)
			return
		}
	}

	return -1, nil
}

func (ctx EvalContext) hasParticipantFlag(exp types.Expression) (val bool, err error) {
	if len(exp.Data) != 2 {
		return val, errors.New("unexpected numbers of arguments")
	}

	if exp.Data[0].IsNumber() || exp.Data[1].IsNumber() {
		return val, errors.New("unexpected argument types")
	}

	arg1, err := ctx.expressionArgResolver(exp.Data[0])
	if err != nil {
		return val, err
	}
	arg1Val, ok := arg1.(string)
	if !ok {
		return val, errors.New("could not cast argument 1")
	}

	arg2, err := ctx.expressionArgResolver(exp.Data[1])
	if err != nil {
		return val, err
	}
	arg2Val, ok := arg2.(string)
	if !ok {
		return val, errors.New("could not cast argument 2")
	}

	value, ok := ctx.ParticipantState.Flags[arg1Val]
	if !ok || value != arg2Val {
		return false, nil
	}
	return true, nil
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
	responseOfInterest, err := findSurveyItemResponse(ctx.Event.Response.Responses, arg1Val)
	if err != nil {
		// Item not found
		return false, nil
	}

	responseParentGroup, err := findResponseObject(responseOfInterest, arg2Val)
	if err != nil {
		// Item not found
		return false, nil
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

func (ctx EvalContext) responseHasOnlyKeysOtherThan(exp types.Expression) (val bool, err error) {
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
	responseOfInterest, err := findSurveyItemResponse(ctx.Event.Response.Responses, arg1Val)
	if err != nil {
		// Item not found
		return false, nil
	}

	responseParentGroup, err := findResponseObject(responseOfInterest, arg2Val)
	if err != nil {
		// Item not found
		return false, nil
	}

	if len(responseParentGroup.Items) < 1 {
		return false, nil
	}

	// Check if any of the target in response
	anyFound := true
	for _, target := range targetKeys {
		for _, item := range responseParentGroup.Items {
			if item.Key == target {
				anyFound = false
				break
			}
		}
		if anyFound {
			break
		}
	}
	return anyFound, nil
}

func (ctx EvalContext) getResponseValueAsNum(exp types.Expression) (val float64, err error) {
	if len(exp.Data) != 2 {
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

	// find survey item:
	surveyItem, err := findSurveyItemResponse(ctx.Event.Response.Responses, arg1Val)
	if err != nil {
		// Item not found
		return 0, errors.New("item not found")
	}

	responseObject, err := findResponseObject(surveyItem, arg2Val)
	if err != nil {
		// Item not found
		return 0, errors.New("item not found")
	}

	val, err = strconv.ParseFloat(responseObject.Value, 64)
	return
}

func (ctx EvalContext) getResponseValueAsStr(exp types.Expression) (val string, err error) {
	if len(exp.Data) != 2 {
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

	// find survey item:
	surveyItem, err := findSurveyItemResponse(ctx.Event.Response.Responses, arg1Val)
	if err != nil {
		// Item not found
		return "", errors.New("item not found")
	}

	responseObject, err := findResponseObject(surveyItem, arg2Val)
	if err != nil {
		// Item not found
		return "", errors.New("item not found")
	}
	val = responseObject.Value
	return
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
	if len(exp.Data) != 1 && len(exp.Data) != 2 {
		return t, errors.New("should have one or two arguments")
	}

	arg1, err1 := ctx.expressionArgResolver(exp.Data[0])
	if err1 != nil {
		return t, err1
	}
	if reflect.TypeOf(arg1).Kind() != reflect.Float64 {
		return t, errors.New("argument 1 should be resolved as type number (float64)")
	}
	delta := int64(arg1.(float64))

	referenceTime := time.Now().Unix()
	if len(exp.Data) == 2 {
		arg2, err2 := ctx.expressionArgResolver(exp.Data[1])
		if err2 != nil {
			return t, err2
		}
		if reflect.TypeOf(arg2).Kind() != reflect.Float64 {
			return t, errors.New("argument 2 should be resolved as type number (float64)")
		}

		referenceTime = int64(arg2.(float64))
	}

	t = float64(referenceTime + delta)
	return
}
