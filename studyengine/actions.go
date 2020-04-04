package studyengine

import (
	"errors"
	"time"

	"github.com/influenzanet/study-service/models"
)

func ActionEval(action models.Expression, oldState models.ParticipantState, event models.StudyEvent) (newState models.ParticipantState, err error) {
	if event.Type == "SUBMIT" {
		oldState, err = updateLastSubmissionForSurvey(oldState, event)
		if err != nil {
			return oldState, err
		}
	}

	switch action.Name {
	case "IFTHEN":
		newState, err = ifThenAction(action, oldState, event)
	case "UPDATE_STUDY_STATUS":
		newState, err = updateStudyStatusAction(action, oldState, event)
	case "UPDATE_FLAG":
		newState, err = updateFlagAction(action, oldState, event)
	case "REMOVE_FLAG":
		newState, err = removeFlagAction(action, oldState, event)
	case "ADD_NEW_SURVEY":
		newState, err = addNewSurveyAction(action, oldState, event)
	case "REMOVE_ALL_SURVEYS":
		newState, err = removeAllSurveys(action, oldState, event)
	case "REMOVE_SURVEY_BY_KEY":
		newState, err = removeSurveyByKey(action, oldState, event)
	case "REMOVE_SURVEYS_BY_KEY":
		newState, err = removeSurveysByKey(action, oldState, event)
	case "ADD_REPORT":
		newState, err = addReport(action, oldState, event)
	case "REMOVE_ALL_REPORTS":
		newState, err = removeAllReports(action, oldState, event)
	case "REMOVE_REPORT_BY_KEY":
		newState, err = removeReportByKey(action, oldState, event)
	case "REMOVE_REPORTS_BY_KEY":
		newState, err = removeReportsByKey(action, oldState, event)
	default:
		newState = oldState
		err = errors.New("action name not known")
	}
	return
}

func updateLastSubmissionForSurvey(oldState models.ParticipantState, event models.StudyEvent) (newState models.ParticipantState, err error) {
	newState = oldState
	if event.Response.Key == "" {
		return newState, errors.New("no response key found")
	}
	if newState.LastSubmissions == nil {
		newState.LastSubmissions = map[string]int64{}
	}
	newState.LastSubmissions[event.Response.Key] = time.Now().Unix()
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

// updateStudyStatusAction is used to update if user is active in the study
func updateStudyStatusAction(action models.Expression, oldState models.ParticipantState, event models.StudyEvent) (newState models.ParticipantState, err error) {
	newState = oldState
	if len(action.Data) != 1 {
		return newState, errors.New("updateStudyStatusAction must have exactly one argument")
	}
	evalContext := evalContext{
		event:            event,
		participantState: newState,
	}
	k, err := evalContext.expressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}

	status, ok := k.(string)
	if !ok {
		return newState, errors.New("could not parse argument")
	}

	newState.StudyStatus = status
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

	if newState.Flags == nil {
		newState.Flags = map[string]string{}
	}
	newState.Flags[key] = value
	return
}

// removeFlagAction is used to update one of the string flags from the participant state
func removeFlagAction(action models.Expression, oldState models.ParticipantState, event models.StudyEvent) (newState models.ParticipantState, err error) {
	newState = oldState
	if len(action.Data) != 1 {
		return newState, errors.New("removeFlagAction must have exactly one argument")
	}
	evalContext := evalContext{
		event:            event,
		participantState: newState,
	}
	k, err := evalContext.expressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}

	key, ok := k.(string)
	if !ok {
		return newState, errors.New("could not parse key")
	}

	delete(newState.Flags, key)
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

// addReport finds and appends a SurveyItemResponse to the reports array
func addReport(action models.Expression, oldState models.ParticipantState, event models.StudyEvent) (newState models.ParticipantState, err error) {
	newState = oldState
	if len(action.Data) != 1 {
		return newState, errors.New("addReport must have exactly one argument")
	}
	evalContext := evalContext{
		event:            event,
		participantState: newState,
	}
	k, err := evalContext.expressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}

	itemKey, ok1 := k.(string)
	if !ok1 {
		return newState, errors.New("could not parse arguments")
	}

	for _, itemResp := range event.Response.Responses {
		if itemResp.Key == itemKey {
			newState.Reports = append(newState.Reports, itemResp)
			break
		}
	}
	return
}

// removeAllReports clears the reports array
func removeAllReports(action models.Expression, oldState models.ParticipantState, event models.StudyEvent) (newState models.ParticipantState, err error) {
	newState = oldState
	if len(action.Data) > 0 {
		return newState, errors.New("removeAllReports must not have arguments")
	}
	newState.Reports = []models.SurveyItemResponse{}
	return
}

// removeReportByKey removes the first or last appearance of the report with specific key
func removeReportByKey(action models.Expression, oldState models.ParticipantState, event models.StudyEvent) (newState models.ParticipantState, err error) {
	newState = oldState
	if len(action.Data) != 2 {
		return newState, errors.New("removeReportByKey must have exactly two arguments")
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

	itemKey, ok1 := k.(string)
	position, ok2 := pos.(string)

	if !ok1 || !ok2 {
		return newState, errors.New("could not parse arguments")
	}

	sr := []models.SurveyItemResponse{}
	switch position {
	case "first":
		found := false
		for _, surv := range newState.Reports {
			if surv.Key == itemKey {
				if !found {
					found = true
					continue
				}
			}
			sr = append(sr, surv)
		}
	case "last":
		ind := -1
		for i, surv := range newState.Reports {
			if surv.Key == itemKey {
				ind = i
			}
		}
		if ind < 0 {
			sr = newState.Reports
		} else {
			sr = append(newState.Reports[:ind], newState.Reports[ind+1:]...)
		}

	default:
		return newState, errors.New("position not known")
	}
	newState.Reports = sr
	return
}

// removeReportsByKey removes all responses with a specific key
func removeReportsByKey(action models.Expression, oldState models.ParticipantState, event models.StudyEvent) (newState models.ParticipantState, err error) {
	newState = oldState
	if len(action.Data) != 1 {
		return newState, errors.New("removeReportsByKey must have exactly one argument")
	}
	evalContext := evalContext{
		event:            event,
		participantState: newState,
	}
	k, err := evalContext.expressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}

	itemKey, ok1 := k.(string)

	if !ok1 {
		return newState, errors.New("could not parse arguments")
	}

	sr := []models.SurveyItemResponse{}
	for _, surv := range newState.Reports {
		if surv.Key != itemKey {
			sr = append(sr, surv)
		}
	}
	newState.Reports = sr
	return
}
