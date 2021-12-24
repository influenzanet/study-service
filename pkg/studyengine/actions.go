package studyengine

import (
	"errors"
	"time"

	"github.com/influenzanet/study-service/pkg/dbs/studydb"
	"github.com/influenzanet/study-service/pkg/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StudyDBService interface {
	FindSurveyResponses(instanceID string, studyKey string, query studydb.ResponseQuery) (responses []types.SurveyResponse, err error)
}

func ActionEval(action types.Expression, oldState types.ParticipantState, event types.StudyEvent, dbService StudyDBService) (newState types.ParticipantState, err error) {
	if event.Type == "SUBMIT" {
		oldState, err = updateLastSubmissionForSurvey(oldState, event)
		if err != nil {
			return oldState, err
		}
	}

	switch action.Name {
	case "IF":
		newState, err = ifAction(action, oldState, event, dbService)
	case "DO":
		newState, err = doAction(action, oldState, event, dbService)
	case "IFTHEN":
		newState, err = ifThenAction(action, oldState, event, dbService)
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
	case "ADD_MESSAGE":
		newState, err = addMessage(action, oldState, event)
	case "REMOVE_ALL_MESSAGES":
		newState, err = removeAllMessages(action, oldState, event)
	case "REMOVE_MESSAGES_BY_KEY":
		newState, err = removeMessagesByKey(action, oldState, event)
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

func updateLastSubmissionForSurvey(oldState types.ParticipantState, event types.StudyEvent) (newState types.ParticipantState, err error) {
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

func checkCondition(condition types.ExpressionArg, EvalContext EvalContext) bool {
	if !condition.IsExpression() {
		return condition.Num != 0
	}
	val, err := ExpressionEval(*condition.Exp, EvalContext)
	bVal, ok := val.(bool)
	return bVal && ok && err == nil
}

// ifAction is used to conditionally perform actions
func ifAction(action types.Expression, oldState types.ParticipantState, event types.StudyEvent, dbService StudyDBService) (newState types.ParticipantState, err error) {
	newState = oldState
	if len(action.Data) < 2 {
		return newState, errors.New("ifAction must have at least two arguments")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState,
		DbService:        dbService,
	}
	var task types.ExpressionArg
	if checkCondition(action.Data[0], EvalContext) {
		task = action.Data[1]
	} else if len(action.Data) == 3 {
		task = action.Data[2]
	}

	if task.IsExpression() {
		newState, err = ActionEval(*task.Exp, newState, event, dbService)
		if err != nil {
			return newState, err
		}
	}
	return
}

// doAction to perform a list of actions
func doAction(action types.Expression, oldState types.ParticipantState, event types.StudyEvent, dbService StudyDBService) (newState types.ParticipantState, err error) {
	newState = oldState
	for _, action := range action.Data {
		if action.IsExpression() {
			newState, err = ActionEval(*action.Exp, newState, event, dbService)
			if err != nil {
				return newState, err
			}
		}
	}
	return
}

// ifThenAction is used to conditionally perform a sequence of actions
func ifThenAction(action types.Expression, oldState types.ParticipantState, event types.StudyEvent, dbService StudyDBService) (newState types.ParticipantState, err error) {
	newState = oldState
	if len(action.Data) < 1 {
		return newState, errors.New("ifThenAction must have at least one argument")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState,
		DbService:        dbService,
	}
	if !checkCondition(action.Data[0], EvalContext) {
		return
	}
	for _, action := range action.Data[1:] {
		if action.IsExpression() {
			newState, err = ActionEval(*action.Exp, newState, event, dbService)
			if err != nil {
				return newState, err
			}
		}
	}
	return
}

// updateStudyStatusAction is used to update if user is active in the study
func updateStudyStatusAction(action types.Expression, oldState types.ParticipantState, event types.StudyEvent) (newState types.ParticipantState, err error) {
	newState = oldState
	if len(action.Data) != 1 {
		return newState, errors.New("updateStudyStatusAction must have exactly one argument")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState,
	}
	k, err := EvalContext.expressionArgResolver(action.Data[0])
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
func updateFlagAction(action types.Expression, oldState types.ParticipantState, event types.StudyEvent) (newState types.ParticipantState, err error) {
	newState = oldState
	if len(action.Data) != 2 {
		return newState, errors.New("updateFlagAction must have exactly two arguments")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState,
	}
	k, err := EvalContext.expressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}
	v, err := EvalContext.expressionArgResolver(action.Data[1])
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
	} else {
		newState.Flags = make(map[string]string)
		for k, v := range oldState.Flags {
			newState.Flags[k] = v
		}
	}
	newState.Flags[key] = value
	return
}

// removeFlagAction is used to update one of the string flags from the participant state
func removeFlagAction(action types.Expression, oldState types.ParticipantState, event types.StudyEvent) (newState types.ParticipantState, err error) {
	newState = oldState
	if len(action.Data) != 1 {
		return newState, errors.New("removeFlagAction must have exactly one argument")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState,
	}
	k, err := EvalContext.expressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}

	key, ok := k.(string)
	if !ok {
		return newState, errors.New("could not parse key")
	}

	if newState.Flags != nil {
		newState.Flags = make(map[string]string)
		for k, v := range oldState.Flags {
			newState.Flags[k] = v
		}
	}

	delete(newState.Flags, key)
	return
}

// addNewSurveyAction appends a new AssignedSurvey for the participant state
func addNewSurveyAction(action types.Expression, oldState types.ParticipantState, event types.StudyEvent) (newState types.ParticipantState, err error) {
	newState = oldState
	if len(action.Data) != 4 {
		return newState, errors.New("addNewSurveyAction must have exactly four arguments")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState,
	}
	k, err := EvalContext.expressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}
	start, err := EvalContext.expressionArgResolver(action.Data[1])
	if err != nil {
		return newState, err
	}
	end, err := EvalContext.expressionArgResolver(action.Data[2])
	if err != nil {
		return newState, err
	}
	c, err := EvalContext.expressionArgResolver(action.Data[3])
	if err != nil {
		return newState, err
	}

	surveyKey, ok1 := k.(string)
	validFrom, ok2 := start.(float64)
	validUntil, ok3 := end.(float64)
	category, ok4 := c.(string)

	if !ok1 || !ok2 || !ok3 || !ok4 {
		return newState, errors.New("could not parse arguments")
	}

	newSurvey := types.AssignedSurvey{
		SurveyKey:  surveyKey,
		ValidFrom:  int64(validFrom),
		ValidUntil: int64(validUntil),
		Category:   category,
	}
	newState.AssignedSurveys = make([]types.AssignedSurvey, len(oldState.AssignedSurveys))
	copy(newState.AssignedSurveys, oldState.AssignedSurveys)

	newState.AssignedSurveys = append(newState.AssignedSurveys, newSurvey)
	return
}

// removeAllSurveys clear the assigned survey list
func removeAllSurveys(action types.Expression, oldState types.ParticipantState, event types.StudyEvent) (newState types.ParticipantState, err error) {
	newState = oldState
	if len(action.Data) > 0 {
		return newState, errors.New("removeAllSurveys must not have arguments")
	}

	newState.AssignedSurveys = []types.AssignedSurvey{}
	return
}

// removeSurveyByKey removes the first or last occurence of a survey
func removeSurveyByKey(action types.Expression, oldState types.ParticipantState, event types.StudyEvent) (newState types.ParticipantState, err error) {
	newState = oldState
	if len(action.Data) != 2 {
		return newState, errors.New("removeSurveyByKey must have exactly two arguments")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState,
	}
	k, err := EvalContext.expressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}
	pos, err := EvalContext.expressionArgResolver(action.Data[1])
	if err != nil {
		return newState, err
	}

	surveyKey, ok1 := k.(string)
	position, ok2 := pos.(string)

	if !ok1 || !ok2 {
		return newState, errors.New("could not parse arguments")
	}

	as := []types.AssignedSurvey{}
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
func removeSurveysByKey(action types.Expression, oldState types.ParticipantState, event types.StudyEvent) (newState types.ParticipantState, err error) {
	newState = oldState
	if len(action.Data) != 1 {
		return newState, errors.New("removeSurveysByKey must have exactly one argument")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState,
	}
	k, err := EvalContext.expressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}

	surveyKey, ok1 := k.(string)

	if !ok1 {
		return newState, errors.New("could not parse arguments")
	}

	as := []types.AssignedSurvey{}
	for _, surv := range newState.AssignedSurveys {
		if surv.SurveyKey != surveyKey {
			as = append(as, surv)
		}
	}
	newState.AssignedSurveys = as
	return
}

// addMessage
func addMessage(action types.Expression, oldState types.ParticipantState, event types.StudyEvent) (newState types.ParticipantState, err error) {
	newState = oldState
	if len(action.Data) != 2 {
		return newState, errors.New("addMessage must have exactly two arguments")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState,
	}
	arg1, err := EvalContext.expressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}
	arg2, err := EvalContext.expressionArgResolver(action.Data[1])
	if err != nil {
		return newState, err
	}

	messageType, ok1 := arg1.(string)
	timestamp, ok2 := arg2.(float64)

	if !ok1 || !ok2 {
		return newState, errors.New("could not parse arguments")
	}

	newMessage := types.ParticipantMessage{
		ID:           primitive.NewObjectID().Hex(),
		Type:         messageType,
		ScheduledFor: int64(timestamp),
	}
	newState.Messages = make([]types.ParticipantMessage, len(oldState.Messages))
	copy(newState.Messages, oldState.Messages)

	newState.Messages = append(newState.Messages, newMessage)
	return
}

// removeAllMessages
func removeAllMessages(action types.Expression, oldState types.ParticipantState, event types.StudyEvent) (newState types.ParticipantState, err error) {
	newState = oldState

	newState.Messages = []types.ParticipantMessage{}
	return
}

// removeSurveysByKey removes all the surveys with a specific key
func removeMessagesByKey(action types.Expression, oldState types.ParticipantState, event types.StudyEvent) (newState types.ParticipantState, err error) {
	newState = oldState
	if len(action.Data) != 1 {
		return newState, errors.New("removeMessagesByKey must have exactly one argument")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState,
	}
	k, err := EvalContext.expressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}

	messageType, ok1 := k.(string)

	if !ok1 {
		return newState, errors.New("could not parse arguments")
	}

	messages := []types.ParticipantMessage{}
	for _, msg := range newState.Messages {
		if msg.Type != messageType {
			messages = append(messages, msg)
		}
	}
	newState.Messages = messages
	return
}

// addReport finds and appends a SurveyItemResponse to the reports array
func addReport(action types.Expression, oldState types.ParticipantState, event types.StudyEvent) (newState types.ParticipantState, err error) {
	newState = oldState
	if len(action.Data) != 1 {
		return newState, errors.New("addReport must have exactly one argument")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState,
	}
	k, err := EvalContext.expressionArgResolver(action.Data[0])
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
func removeAllReports(action types.Expression, oldState types.ParticipantState, event types.StudyEvent) (newState types.ParticipantState, err error) {
	newState = oldState
	if len(action.Data) > 0 {
		return newState, errors.New("removeAllReports must not have arguments")
	}
	newState.Reports = []types.SurveyItemResponse{}
	return
}

// removeReportByKey removes the first or last appearance of the report with specific key
func removeReportByKey(action types.Expression, oldState types.ParticipantState, event types.StudyEvent) (newState types.ParticipantState, err error) {
	newState = oldState
	if len(action.Data) != 2 {
		return newState, errors.New("removeReportByKey must have exactly two arguments")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState,
	}
	k, err := EvalContext.expressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}
	pos, err := EvalContext.expressionArgResolver(action.Data[1])
	if err != nil {
		return newState, err
	}

	itemKey, ok1 := k.(string)
	position, ok2 := pos.(string)

	if !ok1 || !ok2 {
		return newState, errors.New("could not parse arguments")
	}

	sr := []types.SurveyItemResponse{}
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
func removeReportsByKey(action types.Expression, oldState types.ParticipantState, event types.StudyEvent) (newState types.ParticipantState, err error) {
	newState = oldState
	if len(action.Data) != 1 {
		return newState, errors.New("removeReportsByKey must have exactly one argument")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState,
	}
	k, err := EvalContext.expressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}

	itemKey, ok1 := k.(string)

	if !ok1 {
		return newState, errors.New("could not parse arguments")
	}

	sr := []types.SurveyItemResponse{}
	for _, surv := range newState.Reports {
		if surv.Key != itemKey {
			sr = append(sr, surv)
		}
	}
	newState.Reports = sr
	return
}
