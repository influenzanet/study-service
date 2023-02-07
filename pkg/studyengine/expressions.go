package studyengine

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/coneno/logger"
	"github.com/influenzanet/study-service/pkg/dbs/studydb"
	"github.com/influenzanet/study-service/pkg/types"
)

// EvalContext contains all the data that can be looked up by expressions
type EvalContext struct {
	Event            types.StudyEvent
	ParticipantState types.ParticipantState
	Configs          ActionConfigs
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
	case "getSelectedKeys":
		val, err = evalCtx.getSelectedKeys(expression)
	case "countResponseItems":
		val, err = evalCtx.countResponseItems(expression)
	case "hasResponseKey":
		val, err = evalCtx.hasResponseKey(expression)
	case "hasResponseKeyWithValue":
		val, err = evalCtx.hasResponseKeyWithValue(expression)
	// Old responses:
	case "checkConditionForOldResponses":
		val, err = evalCtx.checkConditionForOldResponses(expression)
	// Participant state:
	case "getStudyEntryTime":
		val, err = evalCtx.getStudyEntryTime(expression, false)
	case "hasSurveyKeyAssigned":
		val, err = evalCtx.hasSurveyKeyAssigned(expression, false)
	case "getSurveyKeyAssignedFrom":
		val, err = evalCtx.getSurveyKeyAssignedFrom(expression, false)
	case "getSurveyKeyAssignedUntil":
		val, err = evalCtx.getSurveyKeyAssignedUntil(expression, false)
	case "hasStudyStatus":
		val, err = evalCtx.hasStudyStatus(expression, false)
	case "hasParticipantFlag":
		val, err = evalCtx.hasParticipantFlag(expression, false)
	case "hasParticipantFlagKey":
		val, err = evalCtx.hasParticipantFlagKey(expression, false)
	case "getParticipantFlagValue":
		val, err = evalCtx.getParticipantFlagValue(expression, false)
	case "lastSubmissionDateOlderThan":
		val, err = evalCtx.lastSubmissionDateOlderThan(expression, false)
	case "hasMessageTypeAssigned":
		val, err = evalCtx.hasMessageTypeAssigned(expression, false)
	case "getMessageNextTime":
		val, err = evalCtx.getMessageNextTime(expression, false)
	// exprssions for merge participant states:
	case "incomingState:getStudyEntryTime":
		val, err = evalCtx.getStudyEntryTime(expression, true)
	case "incomingState:hasSurveyKeyAssigned":
		val, err = evalCtx.hasSurveyKeyAssigned(expression, true)
	case "incomingState:getSurveyKeyAssignedFrom":
		val, err = evalCtx.getSurveyKeyAssignedFrom(expression, true)
	case "incomingState:getSurveyKeyAssignedUntil":
		val, err = evalCtx.getSurveyKeyAssignedUntil(expression, true)
	case "incomingState:hasStudyStatus":
		val, err = evalCtx.hasStudyStatus(expression, true)
	case "incomingState:hasParticipantFlag":
		val, err = evalCtx.hasParticipantFlag(expression, true)
	case "incomingState:hasParticipantFlagKey":
		val, err = evalCtx.hasParticipantFlagKey(expression, true)
	case "incomingState:getParticipantFlagValue":
		val, err = evalCtx.getParticipantFlagValue(expression, true)
	case "incomingState:lastSubmissionDateOlderThan":
		val, err = evalCtx.lastSubmissionDateOlderThan(expression, true)
	case "incomingState:hasMessageTypeAssigned":
		val, err = evalCtx.hasMessageTypeAssigned(expression, true)
	case "incomingState:getMessageNextTime":
		val, err = evalCtx.getMessageNextTime(expression, true)
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
	case "externalEventEval":
		val, err = evalCtx.externalEventEval(expression)
	default:
		err = fmt.Errorf("expression name not known: %s", expression.Name)
		logger.Debug.Println(err)
		return
	}
	return
}

func (ctx EvalContext) expressionArgResolver(arg types.ExpressionArg) (interface{}, error) {
	switch arg.DType {
	case "num":
		return arg.Num, nil
	case "exp":
		if arg.Exp == nil {
			return nil, errors.New("missing argument - expected expression, but was empty")
		}
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

func (ctx EvalContext) hasStudyStatus(exp types.Expression, withIncomingParticipantState bool) (val bool, err error) {
	pState := ctx.ParticipantState
	if withIncomingParticipantState {
		pState = ctx.Event.MergeWithParticipant
	}
	if len(exp.Data) != 1 {
		return val, errors.New("unexpected numbers of arguments")
	}

	arg1Val, err := ctx.mustGetStrValue(exp.Data[0])
	if err != nil {
		return val, err
	}

	return pState.StudyStatus == arg1Val, nil
}

func (ctx EvalContext) checkConditionForOldResponses(exp types.Expression) (val bool, err error) {
	if ctx.Configs.DBService == nil {
		return val, errors.New("checkConditionForOldResponses: DB connection not available in the context")
	}
	if ctx.Event.InstanceID == "" || ctx.Event.StudyKey == "" {
		return val, errors.New("checkConditionForOldResponses: instanceID or study key missing from context")
	}

	argNum := len(exp.Data)
	if argNum < 1 || argNum > 5 {
		return val, fmt.Errorf("checkConditionForOldResponses: unexpected numbers of arguments: %d", len(exp.Data))
	}

	arg1 := exp.Data[0]
	if !arg1.IsExpression() {
		return val, errors.New("checkConditionForOldResponses: first argument must be an expression")
	}
	condition := arg1.Exp
	if condition == nil {
		return val, errors.New("checkConditionForOldResponses: first argument must be an expression")
	}

	checkFor := "all"
	checkForCount := 1
	surveyKey := ""
	since := int64(0)
	until := int64(0)
	if argNum > 1 {
		arg1, err := ctx.expressionArgResolver(exp.Data[1])
		if err != nil {
			return val, err
		}
		switch arg1Val := arg1.(type) {
		case string:
			checkFor = arg1Val
		case float64:
			checkFor = "count"
			checkForCount = int(arg1Val)
		default:
			return val, fmt.Errorf("type unknown %T", arg1Val)
		}
	}
	if argNum > 2 {
		surveyKey, err = ctx.mustGetStrValue(exp.Data[2])
		if err != nil {
			return val, err
		}
	}
	if argNum > 3 {
		arg4, err := ctx.expressionArgResolver(exp.Data[3])
		if err != nil {
			return val, err
		}
		arg4Val, ok := arg4.(float64)
		if ok {
			since = int64(arg4Val)
		}

	}
	if argNum > 4 {
		arg5, err := ctx.expressionArgResolver(exp.Data[4])
		if err != nil {
			return val, err
		}
		arg5Val, ok := arg5.(float64)
		if ok {
			until = int64(arg5Val)
		}
	}

	responses, err := ctx.Configs.DBService.FindSurveyResponses(ctx.Event.InstanceID, ctx.Event.StudyKey, studydb.ResponseQuery{
		ParticipantID: ctx.ParticipantState.ParticipantID,
		SurveyKey:     surveyKey,
		Since:         since,
		Until:         until,
	})
	if err != nil {
		return val, err
	}

	counter := 0
	result := false
	for _, resp := range responses {
		oldEvalContext := EvalContext{
			ParticipantState: ctx.ParticipantState,
			Event: types.StudyEvent{
				Response: resp,
			},
		}

		expResult, err := ExpressionEval(*condition, oldEvalContext)
		if err != nil {
			return false, err
		}
		val := expResult.(bool)

		switch checkFor {
		case "all":
			if val {
				result = true
			} else {
				result = false
				break
			}
		case "any":
			if val {
				result = true
				break
			}
		case "count":
			if val {
				counter += 1
				if counter >= checkForCount {
					result = true
					break
				}
			}
		}
	}

	return result, nil
}

func (ctx EvalContext) getStudyEntryTime(exp types.Expression, withIncomingParticipantState bool) (t float64, err error) {
	pState := ctx.ParticipantState
	if withIncomingParticipantState {
		pState = ctx.Event.MergeWithParticipant
	}
	return float64(pState.EnteredAt), nil
}

func (ctx EvalContext) hasSurveyKeyAssigned(exp types.Expression, withIncomingParticipantState bool) (val bool, err error) {
	pState := ctx.ParticipantState
	if withIncomingParticipantState {
		pState = ctx.Event.MergeWithParticipant
	}

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

	for _, survey := range pState.AssignedSurveys {
		if survey.SurveyKey == arg1Val {
			val = true
			return
		}
	}
	return
}

func (ctx EvalContext) getSurveyKeyAssignedFrom(exp types.Expression, withIncomingParticipantState bool) (val float64, err error) {
	pState := ctx.ParticipantState
	if withIncomingParticipantState {
		pState = ctx.Event.MergeWithParticipant
	}

	if len(exp.Data) != 1 || !exp.Data[0].IsString() {
		return val, errors.New("unexpected number or wrong type of argument")
	}

	arg1Val, err := ctx.mustGetStrValue(exp.Data[0])
	if err != nil {
		return val, err
	}

	for _, survey := range pState.AssignedSurveys {
		if survey.SurveyKey == arg1Val {
			val = float64(survey.ValidFrom)
			return
		}
	}

	return -1, nil
}

func (ctx EvalContext) getSurveyKeyAssignedUntil(exp types.Expression, withIncomingParticipantState bool) (val float64, err error) {
	pState := ctx.ParticipantState
	if withIncomingParticipantState {
		pState = ctx.Event.MergeWithParticipant
	}

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

	for _, survey := range pState.AssignedSurveys {
		if survey.SurveyKey == arg1Val {
			val = float64(survey.ValidUntil)
			return
		}
	}

	return -1, nil
}

func (ctx EvalContext) hasParticipantFlagKey(exp types.Expression, withIncomingParticipantState bool) (val bool, err error) {
	pState := ctx.ParticipantState
	if withIncomingParticipantState {
		pState = ctx.Event.MergeWithParticipant
	}
	if len(exp.Data) != 1 {
		return val, errors.New("unexpected numbers of arguments")
	}

	if exp.Data[0].IsNumber() {
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

	_, ok = pState.Flags[arg1Val]
	if !ok {
		return false, nil
	}
	return true, nil
}

func (ctx EvalContext) getParticipantFlagValue(exp types.Expression, withIncomingParticipantState bool) (val string, err error) {
	pState := ctx.ParticipantState
	if withIncomingParticipantState {
		pState = ctx.Event.MergeWithParticipant
	}
	if len(exp.Data) != 1 {
		return val, errors.New("unexpected numbers of arguments")
	}

	if exp.Data[0].IsNumber() {
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

	res, ok := pState.Flags[arg1Val]
	if !ok {
		return "", nil
	}
	return res, nil
}

func (ctx EvalContext) hasParticipantFlag(exp types.Expression, withIncomingParticipantState bool) (val bool, err error) {
	pState := ctx.ParticipantState
	if withIncomingParticipantState {
		pState = ctx.Event.MergeWithParticipant
	}
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

	value, ok := pState.Flags[arg1Val]
	if !ok || value != arg2Val {
		return false, nil
	}
	return true, nil
}

func (ctx EvalContext) lastSubmissionDateOlderThan(exp types.Expression, withIncomingParticipantState bool) (val bool, err error) {
	pState := ctx.ParticipantState
	if withIncomingParticipantState {
		pState = ctx.Event.MergeWithParticipant
	}
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
	refTime := int64(arg1Val)

	if len(exp.Data) == 2 {
		arg2Val, err := ctx.mustGetStrValue(exp.Data[1])
		if err != nil {
			return val, err
		}
		lastTs, ok := pState.LastSubmissions[arg2Val]
		if !ok {
			return false, nil
		}
		return lastTs < refTime, nil

	} else {
		for _, lastTs := range pState.LastSubmissions {
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

func (ctx EvalContext) getSelectedKeys(exp types.Expression) (val string, err error) {
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

	keys := []string{}
	for _, item := range responseObject.Items {
		keys = append(keys, item.Key)
	}
	val = strings.Join(keys, ";")
	return
}

func (ctx EvalContext) countResponseItems(exp types.Expression) (val float64, err error) {
	if len(exp.Data) != 2 {
		return val, errors.New("unexpected numbers of arguments")
	}

	itemKey, err := ctx.mustGetStrValue(exp.Data[0])
	if err != nil {
		return val, err
	}

	responseGroupKey, err := ctx.mustGetStrValue(exp.Data[1])
	if err != nil {
		return val, err
	}

	// find survey item:
	surveyItem, err := findSurveyItemResponse(ctx.Event.Response.Responses, itemKey)
	if err != nil {
		// Item not found
		return -1.0, errors.New("item not found")
	}

	responseObject, err := findResponseObject(surveyItem, responseGroupKey)
	if err != nil {
		// Item not found
		return -1.0, errors.New("item not found")
	}

	val = float64(len(responseObject.Items))
	return
}

func (ctx EvalContext) hasResponseKey(exp types.Expression) (val bool, err error) {
	if len(exp.Data) != 2 {
		return val, errors.New("unexpected numbers of arguments")
	}

	itemKey, err := ctx.mustGetStrValue(exp.Data[0])
	if err != nil {
		return val, err
	}

	responseGroupKey, err := ctx.mustGetStrValue(exp.Data[1])
	if err != nil {
		return val, err
	}

	// find survey item:
	surveyItem, err := findSurveyItemResponse(ctx.Event.Response.Responses, itemKey)
	if err != nil {
		// Item not found
		return false, nil
	}

	_, err = findResponseObject(surveyItem, responseGroupKey)
	if err != nil {
		// Item not found
		return false, nil
	}
	return true, nil
}

func (ctx EvalContext) hasResponseKeyWithValue(exp types.Expression) (val bool, err error) {
	if len(exp.Data) != 3 {
		return val, errors.New("unexpected numbers of arguments")
	}

	itemKey, err := ctx.mustGetStrValue(exp.Data[0])
	if err != nil {
		return val, err
	}

	responseKey, err := ctx.mustGetStrValue(exp.Data[1])
	if err != nil {
		return val, err
	}

	value, err := ctx.mustGetStrValue(exp.Data[2])
	if err != nil {
		return val, err
	}

	// find survey item:
	surveyItem, err := findSurveyItemResponse(ctx.Event.Response.Responses, itemKey)
	if err != nil {
		// Item not found
		return false, nil
	}

	responseObject, err := findResponseObject(surveyItem, responseKey)
	if err != nil {
		// Item not found
		return false, nil
	}

	val = responseObject.Value == value
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

func (ctx EvalContext) hasMessageTypeAssigned(exp types.Expression, withIncomingParticipantState bool) (val bool, err error) {
	pState := ctx.ParticipantState
	if withIncomingParticipantState {
		pState = ctx.Event.MergeWithParticipant
	}
	if len(exp.Data) != 1 {
		return val, errors.New("should have at exactly one argument")
	}
	arg1, err := ctx.expressionArgResolver(exp.Data[0])
	if err != nil {
		return val, err
	}
	for _, m := range pState.Messages {
		if m.Type == arg1 {
			return true, nil
		}
	}
	return false, nil
}

func (ctx EvalContext) getMessageNextTime(exp types.Expression, withIncomingParticipantState bool) (val int64, err error) {
	pState := ctx.ParticipantState
	if withIncomingParticipantState {
		pState = ctx.Event.MergeWithParticipant
	}
	if len(exp.Data) != 1 {
		return val, errors.New("should have at exactly one argument")
	}
	arg1, err := ctx.expressionArgResolver(exp.Data[0])
	if err != nil {
		return val, err
	}
	msgType := arg1.(string)
	nextTime := int64(0)
	for _, m := range pState.Messages {
		if m.Type == msgType {
			if nextTime == 0 || nextTime > m.ScheduledFor {
				nextTime = m.ScheduledFor
			}
		}
	}
	if nextTime == 0 {
		return 0, errors.New("no message for this type found")
	}
	return nextTime, nil
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
			logger.Debug.Printf("exp in 'or' returned error: %v", err)
			continue
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

func (ctx EvalContext) externalEventEval(exp types.Expression) (val interface{}, err error) {
	if len(exp.Data) != 1 {
		return val, errors.New("should have one argument")
	}

	serviceName, err := ctx.mustGetStrValue(exp.Data[0])
	if err != nil {
		return val, err
	}

	serviceConfig, err := getExternalServicesConfigByName(ctx.Configs.ExternalServiceConfigs, serviceName)
	if err != nil {
		logger.Error.Println(err)
		return val, err
	}

	payload := ExternalEventPayload{
		ParticipantState: ctx.ParticipantState,
		EventType:        ctx.Event.Type,
		StudyKey:         ctx.Event.StudyKey,
		InstanceID:       ctx.Event.InstanceID,
		Response:         ctx.Event.Response,
	}
	response, err := runHTTPcall(serviceConfig.URL, serviceConfig.APIKey, payload)
	if err != nil {
		logger.Error.Println(err)
		return val, err
	}

	logger.Debug.Printf("%s replied: %v", serviceName, response)

	// if relevant, update participant state:
	value := response["value"]
	if exp.ReturnType == "float" {
		return value.(float64), nil
	}
	return value, nil
}

func (ctx EvalContext) mustGetStrValue(arg types.ExpressionArg) (string, error) {
	arg1, err := ctx.expressionArgResolver(arg)
	if err != nil {
		return "", err
	}
	val, ok := arg1.(string)
	if !ok {
		return "", errors.New("could not cast argument")
	}
	return val, nil
}
