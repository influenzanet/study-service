package studyengine

import (
	"errors"
	"fmt"

	"github.com/influenzanet/study-service/models"
)

func ActionEval(action models.Expression, oldState models.ParticipantState, event models.StudyEvent) (newState models.ParticipantState, err error) {
	newState = oldState

	switch action.Name {
	case "IFTHEN":
		newState, err = ifThenAction(action, oldState, event)
	case "UPDATE_FLAG":
		newState, err = updateFlagAction(action, oldState, event)
	case "ADD_NEW_SURVEY":
		err = errors.New("not implemented")
	case "REMOVE_ALL_SURVEYS":
		err = errors.New("not implemented")
	case "REMOVE_SURVEY_BY_KEY":
		err = errors.New("not implemented")
	case "REMOVE_SURVEYS_BY_KEY":
		err = errors.New("not implemented")
	case "ADD_REPORT":
		err = errors.New("not implemented")
	case "REMOVE_ALL_REPORTS":
		err = errors.New("not implemented")
	case "REMOVE_REPORT_BY_KEY":
		err = errors.New("not implemented")
	case "REMOVE_REPORTS_BY_KEY":
		err = errors.New("not implemented")
	default:
		err = errors.New("action name not known")
	}

	return
}

func checkCondition(condition models.ExpressionArg, evalContext evalContext) bool {
	if !condition.IsExpression() {
		return condition.Num != 0
	}
	val, err := ExpressionEval(condition.Exp, evalContext)
	bVal, ok := val.(bool)
	return bVal && ok && err == nil
}

// ifThenAction is used to conditionally perform a sequence of actions
func ifThenAction(action models.Expression, oldState models.ParticipantState, event models.StudyEvent) (newState models.ParticipantState, err error) {
	newState = oldState
	if len(action.Data) < 1 {
		return newState, errors.New("ifThenAction must have exactly one argument")
	}
	evalContext := evalContext{
		event:            event,
		participantState: newState,
	}
	if !checkCondition(action.Data[0], evalContext) {
		return
	}
	for _, action := range action.Data[1:] {
		if action.IsExpression() {
			newState, err = ActionEval(action.Exp, newState, event)
			if err != nil {
				return newState, err
			}
		}
	}
	return
}

// updateFlagAction is used to update one of the string flags from the participant state
func updateFlagAction(action models.Expression, oldState models.ParticipantState, event models.StudyEvent) (newState models.ParticipantState, err error) {
	newState = oldState
	if len(action.Data) != 2 {
		return newState, errors.New("updateFlagAction must have exactly two arguments")
	}
	evalContext := evalContext{
		event:            event,
		participantState: newState,
	}
	k, err := evalContext.expressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}
	v, err := evalContext.expressionArgResolver(action.Data[1])
	if err != nil {
		return newState, err
	}

	key := k.(string)
	value := v.(string)

	switch key {
	case "status":
		newState.Flags.Status = value
	default:
		return newState, fmt.Errorf("status key not known: %s", key)
	}

	return
}
