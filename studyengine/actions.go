package studyengine

import (
	"errors"

	"github.com/influenzanet/study-service/models"
)

func ActionEval(action models.Action, oldState models.ParticipantState, event models.StudyEvent) (newState models.ParticipantState, err error) {
	newState = oldState

	switch action.Name {
	case "IFTHEN":
		newState, err = ifThenAction(action, oldState, event)
	case "UPDATE_FLAG_STATUS":
		err = errors.New("not implemented")
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

func ifThenAction(action models.Action, oldState models.ParticipantState, event models.StudyEvent) (newState models.ParticipantState, err error) {
	newState = oldState
	evalContext := evalContext{
		event:            event,
		participantState: newState,
	}
	if !checkCondition(action.Condition, evalContext) {
		return
	}
	for _, action := range action.Actions {
		newState, err = ActionEval(action, newState, event)
	}
	return
}
