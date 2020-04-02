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
		newState, err = addNewSurveyAction(action, oldState, event)
	case "REMOVE_ALL_SURVEYS":
		newState, err = removeAllSurveys(action, oldState, event)
	case "REMOVE_SURVEY_BY_KEY":
		newState, err = removeSurveyByKey(action, oldState, event)
	case "REMOVE_SURVEYS_BY_KEY":
		newState, err = removeSurveysByKey(action, oldState, event)
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

	key, ok := k.(string)
	value, ok2 := v.(string)
	if !ok || !ok2 {
		return newState, errors.New("could not parse key/value")
	}

	switch key {
	case "status":
		newState.Flags.Status = value
	default:
		return newState, fmt.Errorf("status key not known: %s", key)
	}

	return
}

// addNewSurveyAction appends a new AssignedSurvey for the participant state
func addNewSurveyAction(action models.Expression, oldState models.ParticipantState, event models.StudyEvent) (newState models.ParticipantState, err error) {
	newState = oldState
	if len(action.Data) != 3 {
		return newState, errors.New("addNewSurveyAction must have exactly three arguments")
	}
	evalContext := evalContext{
		event:            event,
		participantState: newState,
	}
	k, err := evalContext.expressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}
	start, err := evalContext.expressionArgResolver(action.Data[1])
	if err != nil {
		return newState, err
	}
	end, err := evalContext.expressionArgResolver(action.Data[2])
	if err != nil {
		return newState, err
	}

	surveyKey, ok1 := k.(string)
	validFrom, ok2 := start.(float64)
	validUntil, ok3 := end.(float64)

	if !ok1 || !ok2 || !ok3 {
		return newState, errors.New("could not parse arguments")
	}

	newSurvey := models.AssignedSurvey{
		SurveyKey:  surveyKey,
		ValidFrom:  int64(validFrom),
		ValidUntil: int64(validUntil),
	}
	newState.AssignedSurveys = append(newState.AssignedSurveys, newSurvey)
	return
}

// removeAllSurveys clear the assigned survey list
func removeAllSurveys(action models.Expression, oldState models.ParticipantState, event models.StudyEvent) (newState models.ParticipantState, err error) {
	newState = oldState
	if len(action.Data) > 0 {
		return newState, errors.New("addNewSurveyAction must not have arguments")
	}

	newState.AssignedSurveys = []models.AssignedSurvey{}
	return
}

// removeSurveyByKey removes the first or last occurence of a survey
func removeSurveyByKey(action models.Expression, oldState models.ParticipantState, event models.StudyEvent) (newState models.ParticipantState, err error) {
	newState = oldState
	if len(action.Data) != 2 {
		return newState, errors.New("removeSurveyByKey must have exactly two arguments")
	}
	evalContext := evalContext{
		event:            event,
		participantState: newState,
	}
	k, err := evalContext.expressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}
	pos, err := evalContext.expressionArgResolver(action.Data[1])
	if err != nil {
		return newState, err
	}

	surveyKey, ok1 := k.(string)
	position, ok2 := pos.(string)

	if !ok1 || !ok2 {
		return newState, errors.New("could not parse arguments")
	}

	as := []models.AssignedSurvey{}
	switch position {
	case "first":
		found := false
		for _, surv := range newState.AssignedSurveys {
			if surv.SurveyKey == surveyKey {
				if !found {
					found = true
					continue
				}
			}
			as = append(as, surv)
		}
	case "last":
		ind := -1
		for i, surv := range newState.AssignedSurveys {
			if surv.SurveyKey == surveyKey {
				ind = i
			}
		}
		if ind < 0 {
			as = newState.AssignedSurveys
		} else {
			as = append(newState.AssignedSurveys[:ind], newState.AssignedSurveys[ind+1:]...)
		}

	default:
		return newState, errors.New("position not known")
	}
	newState.AssignedSurveys = as
	return
}

// removeSurveysByKey removes all the surveys with a specific key
func removeSurveysByKey(action models.Expression, oldState models.ParticipantState, event models.StudyEvent) (newState models.ParticipantState, err error) {
	newState = oldState
	if len(action.Data) != 1 {
		return newState, errors.New("removeSurveysByKey must have exactly one argument")
	}
	evalContext := evalContext{
		event:            event,
		participantState: newState,
	}
	k, err := evalContext.expressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}

	surveyKey, ok1 := k.(string)

	if !ok1 {
		return newState, errors.New("could not parse arguments")
	}

	as := []models.AssignedSurvey{}
	for _, surv := range newState.AssignedSurveys {
		if surv.SurveyKey != surveyKey {
			as = append(as, surv)
		}
	}
	newState.AssignedSurveys = as
	return
}
