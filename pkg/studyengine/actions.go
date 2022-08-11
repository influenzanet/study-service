package studyengine

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/coneno/logger"
	"github.com/influenzanet/study-service/pkg/dbs/studydb"
	"github.com/influenzanet/study-service/pkg/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StudyDBService interface {
	FindSurveyResponses(instanceID string, studyKey string, query studydb.ResponseQuery) (responses []types.SurveyResponse, err error)
	DeleteConfidentialResponses(instanceID string, studyKey string, participantID string, key string) (count int64, err error)
	SaveResearcherMessage(instanceID string, studyKey string, message types.StudyMessage) error
}

type ActionData struct {
	PState          types.ParticipantState
	ReportsToCreate map[string]types.Report
}

type ActionConfigs struct {
	DBService              StudyDBService
	ExternalServiceConfigs []types.ExternalService
}

func ActionEval(action types.Expression, oldState ActionData, event types.StudyEvent, configs ActionConfigs) (newState ActionData, err error) {
	if event.Type == "SUBMIT" {
		oldState, err = updateLastSubmissionForSurvey(oldState, event)
		if err != nil {
			return oldState, err
		}
	}

	switch action.Name {
	case "IF":
		newState, err = ifAction(action, oldState, event, configs)
	case "DO":
		newState, err = doAction(action, oldState, event, configs)
	case "IFTHEN":
		newState, err = ifThenAction(action, oldState, event, configs)
	case "UPDATE_STUDY_STATUS":
		newState, err = updateStudyStatusAction(action, oldState, event, configs)
	case "START_NEW_STUDY_SESSION":
		newState, err = startNewStudySession(action, oldState, event, configs)
	case "UPDATE_FLAG":
		newState, err = updateFlagAction(action, oldState, event, configs)
	case "REMOVE_FLAG":
		newState, err = removeFlagAction(action, oldState, event, configs)
	case "ADD_NEW_SURVEY":
		newState, err = addNewSurveyAction(action, oldState, event, configs)
	case "REMOVE_ALL_SURVEYS":
		newState, err = removeAllSurveys(action, oldState, event, configs)
	case "REMOVE_SURVEY_BY_KEY":
		newState, err = removeSurveyByKey(action, oldState, event, configs)
	case "REMOVE_SURVEYS_BY_KEY":
		newState, err = removeSurveysByKey(action, oldState, event, configs)
	case "ADD_MESSAGE":
		newState, err = addMessage(action, oldState, event, configs)
	case "REMOVE_ALL_MESSAGES":
		newState, err = removeAllMessages(action, oldState, event, configs)
	case "REMOVE_MESSAGES_BY_TYPE":
		newState, err = removeMessagesByType(action, oldState, event, configs)
	case "NOTIFY_RESEARCHER":
		newState, err = notifyResearcher(action, oldState, event, configs)
	case "INIT_REPORT":
		newState, err = initReport(action, oldState, event, configs)
	case "UPDATE_REPORT_DATA":
		newState, err = updateReportData(action, oldState, event, configs)
	case "REMOVE_REPORT_DATA":
		newState, err = removeReportData(action, oldState, event, configs)
	case "CANCEL_REPORT":
		newState, err = cancelReport(action, oldState, event, configs)
	case "REMOVE_CONFIDENTIAL_RESPONSE_BY_KEY":
		newState, err = removeConfidentialResponseByKey(action, oldState, event, configs)
	case "REMOVE_ALL_CONFIDENTIAL_RESPONSES":
		newState, err = removeAllConfidentialResponses(action, oldState, event, configs)
	case "EXTERNAL_EVENT_HANDLER":
		newState, err = externalEventHandler(action, oldState, event, configs)
	default:
		newState = oldState
		err = errors.New("action name not known")
	}
	if err != nil {
		logger.Debug.Printf("error when running action: %v, %v", action, err)
	}
	return
}

func updateLastSubmissionForSurvey(oldState ActionData, event types.StudyEvent) (newState ActionData, err error) {
	newState = oldState
	if event.Response.Key == "" {
		return newState, errors.New("no response key found")
	}
	if newState.PState.LastSubmissions == nil {
		newState.PState.LastSubmissions = map[string]int64{}
	}
	newState.PState.LastSubmissions[event.Response.Key] = time.Now().Unix()
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
func ifAction(action types.Expression, oldState ActionData, event types.StudyEvent, configs ActionConfigs) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) < 2 {
		return newState, errors.New("ifAction must have at least two arguments")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
		Configs:          configs,
	}
	var task types.ExpressionArg
	if checkCondition(action.Data[0], EvalContext) {
		task = action.Data[1]
	} else if len(action.Data) == 3 {
		task = action.Data[2]
	}

	if task.IsExpression() {
		newState, err = ActionEval(*task.Exp, newState, event, configs)
		if err != nil {
			return newState, err
		}
	}
	return
}

// doAction to perform a list of actions
func doAction(action types.Expression, oldState ActionData, event types.StudyEvent, configs ActionConfigs) (newState ActionData, err error) {
	newState = oldState
	for _, action := range action.Data {
		if action.IsExpression() {
			newState, err = ActionEval(*action.Exp, newState, event, configs)
			if err != nil {
				return newState, err
			}
		}
	}
	return
}

// ifThenAction is used to conditionally perform a sequence of actions
func ifThenAction(action types.Expression, oldState ActionData, event types.StudyEvent, configs ActionConfigs) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) < 1 {
		return newState, errors.New("ifThenAction must have at least one argument")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
		Configs:          configs,
	}
	if !checkCondition(action.Data[0], EvalContext) {
		return
	}
	for _, action := range action.Data[1:] {
		if action.IsExpression() {
			newState, err = ActionEval(*action.Exp, newState, event, configs)
			if err != nil {
				logger.Debug.Printf("ifThen: %v", err)
			}
		}
	}
	return
}

// updateStudyStatusAction is used to update if user is active in the study
func updateStudyStatusAction(action types.Expression, oldState ActionData, event types.StudyEvent, configs ActionConfigs) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) != 1 {
		return newState, errors.New("updateStudyStatusAction must have exactly one argument")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
		Configs:          configs,
	}
	k, err := EvalContext.expressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}

	status, ok := k.(string)
	if !ok {
		return newState, errors.New("could not parse argument")
	}

	newState.PState.StudyStatus = status
	return
}

// startNewStudySession is used to generate a new study session ID
func startNewStudySession(action types.Expression, oldState ActionData, event types.StudyEvent, configs ActionConfigs) (newState ActionData, err error) {
	newState = oldState

	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		logger.Error.Println(err)
	}

	newState.PState.CurrentStudySession = strconv.FormatInt(time.Now().Unix(), 16) + hex.EncodeToString(bytes)
	return
}

// updateFlagAction is used to update one of the string flags from the participant state
func updateFlagAction(action types.Expression, oldState ActionData, event types.StudyEvent, configs ActionConfigs) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) != 2 {
		return newState, errors.New("updateFlagAction must have exactly two arguments")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
		Configs:          configs,
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
	if !ok {
		return newState, errors.New("could not parse flag key")
	}

	value := ""
	switch flagVal := v.(type) {
	case string:
		value = flagVal
	case float64:
		value = fmt.Sprintf("%f", flagVal)
	case bool:
		value = fmt.Sprintf("%t", flagVal)
	}

	if newState.PState.Flags == nil {
		newState.PState.Flags = map[string]string{}
	} else {
		newState.PState.Flags = make(map[string]string)
		for k, v := range oldState.PState.Flags {
			newState.PState.Flags[k] = v
		}
	}
	newState.PState.Flags[key] = value
	return
}

// removeFlagAction is used to update one of the string flags from the participant state
func removeFlagAction(action types.Expression, oldState ActionData, event types.StudyEvent, configs ActionConfigs) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) != 1 {
		return newState, errors.New("removeFlagAction must have exactly one argument")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
		Configs:          configs,
	}
	k, err := EvalContext.expressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}

	key, ok := k.(string)
	if !ok {
		return newState, errors.New("could not parse key")
	}

	if newState.PState.Flags != nil {
		newState.PState.Flags = make(map[string]string)
		for k, v := range oldState.PState.Flags {
			newState.PState.Flags[k] = v
		}
	}

	delete(newState.PState.Flags, key)
	return
}

// addNewSurveyAction appends a new AssignedSurvey for the participant state
func addNewSurveyAction(action types.Expression, oldState ActionData, event types.StudyEvent, configs ActionConfigs) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) != 4 {
		return newState, errors.New("addNewSurveyAction must have exactly four arguments")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
		Configs:          configs,
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
	newState.PState.AssignedSurveys = make([]types.AssignedSurvey, len(oldState.PState.AssignedSurveys))
	copy(newState.PState.AssignedSurveys, oldState.PState.AssignedSurveys)

	newState.PState.AssignedSurveys = append(newState.PState.AssignedSurveys, newSurvey)
	return
}

// removeAllSurveys clear the assigned survey list
func removeAllSurveys(action types.Expression, oldState ActionData, event types.StudyEvent, configs ActionConfigs) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) > 0 {
		return newState, errors.New("removeAllSurveys must not have arguments")
	}

	newState.PState.AssignedSurveys = []types.AssignedSurvey{}
	return
}

// removeSurveyByKey removes the first or last occurence of a survey
func removeSurveyByKey(action types.Expression, oldState ActionData, event types.StudyEvent, configs ActionConfigs) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) != 2 {
		return newState, errors.New("removeSurveyByKey must have exactly two arguments")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
		Configs:          configs,
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
		for _, surv := range newState.PState.AssignedSurveys {
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
		for i, surv := range newState.PState.AssignedSurveys {
			if surv.SurveyKey == surveyKey {
				ind = i
			}
		}
		if ind < 0 {
			as = newState.PState.AssignedSurveys
		} else {
			as = append(newState.PState.AssignedSurveys[:ind], newState.PState.AssignedSurveys[ind+1:]...)
		}

	default:
		return newState, errors.New("position not known")
	}
	newState.PState.AssignedSurveys = as
	return
}

// removeSurveysByKey removes all the surveys with a specific key
func removeSurveysByKey(action types.Expression, oldState ActionData, event types.StudyEvent, configs ActionConfigs) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) != 1 {
		return newState, errors.New("removeSurveysByKey must have exactly one argument")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
		Configs:          configs,
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
	for _, surv := range newState.PState.AssignedSurveys {
		if surv.SurveyKey != surveyKey {
			as = append(as, surv)
		}
	}
	newState.PState.AssignedSurveys = as
	return
}

// addMessage
func addMessage(action types.Expression, oldState ActionData, event types.StudyEvent, configs ActionConfigs) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) != 2 {
		return newState, errors.New("addMessage must have exactly two arguments")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
		Configs:          configs,
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
	newState.PState.Messages = make([]types.ParticipantMessage, len(oldState.PState.Messages))
	copy(newState.PState.Messages, oldState.PState.Messages)

	newState.PState.Messages = append(newState.PState.Messages, newMessage)
	return
}

// removeAllMessages
func removeAllMessages(action types.Expression, oldState ActionData, event types.StudyEvent, configs ActionConfigs) (newState ActionData, err error) {
	newState = oldState

	newState.PState.Messages = []types.ParticipantMessage{}
	return
}

// removeSurveysByKey removes all the surveys with a specific key
func removeMessagesByType(action types.Expression, oldState ActionData, event types.StudyEvent, configs ActionConfigs) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) != 1 {
		return newState, errors.New("removeMessagesByType must have exactly one argument")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
		Configs:          configs,
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
	for _, msg := range newState.PState.Messages {
		if msg.Type != messageType {
			messages = append(messages, msg)
		}
	}
	newState.PState.Messages = messages
	return
}

// notifyResearcher can save a specific message with a payload, that should be sent out to the researcher
func notifyResearcher(action types.Expression, oldState ActionData, event types.StudyEvent, configs ActionConfigs) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) < 1 {
		return newState, errors.New("notifyResearcher must have at least one argument")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
		Configs:          configs,
	}
	k, err := EvalContext.expressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}

	messageType, ok1 := k.(string)
	if !ok1 {
		return newState, errors.New("could not parse arguments")
	}

	payload := map[string]string{}

	for i := 1; i < len(action.Data)-1; i = i + 2 {
		k, err := EvalContext.expressionArgResolver(action.Data[i])
		if err != nil {
			return newState, err
		}
		v, err := EvalContext.expressionArgResolver(action.Data[i+1])
		if err != nil {
			return newState, err
		}

		key, ok := k.(string)
		if !ok {
			return newState, errors.New("could not parse key")
		}
		value, ok := v.(string)
		if !ok {
			return newState, errors.New("could not parse value")
		}

		payload[key] = value
	}

	message := types.StudyMessage{
		Type:          messageType,
		ParticipantID: oldState.PState.ParticipantID,
		Payload:       payload,
	}

	err = configs.DBService.SaveResearcherMessage(event.InstanceID, event.StudyKey, message)
	if err != nil {
		logger.Error.Printf("unexpected error when saving researcher message: %v", err)
	}
	return
}

// init one empty report for the current event - if report already existing, reset report to empty report
func initReport(action types.Expression, oldState ActionData, event types.StudyEvent, configs ActionConfigs) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) != 1 {
		return newState, errors.New("initReport must have exactly one argument")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
		Configs:          configs,
	}
	k, err := EvalContext.expressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}

	reportKey, ok1 := k.(string)
	if !ok1 {
		return newState, errors.New("could not parse arguments")
	}

	newState.ReportsToCreate[reportKey] = types.Report{
		Key:           reportKey,
		ParticipantID: oldState.PState.ParticipantID,
		Timestamp:     time.Now().Truncate(time.Minute).Unix(),
	}
	return
}

// update one data entry in the report with the key, if report was not initialised, init one directly
func updateReportData(action types.Expression, oldState ActionData, event types.StudyEvent, configs ActionConfigs) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) < 3 {
		return newState, errors.New("updateReportData must have at least 3 arguments")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
		Configs:          configs,
	}
	k, err := EvalContext.expressionArgResolver(action.Data[0])
	if err != nil {
		logger.Error.Printf("unexpected error: %v", err)
		return newState, err
	}

	reportKey, ok1 := k.(string)
	if !ok1 {
		return newState, errors.New("could not parse arguments")
	}

	// If report not initialized yet, init report:
	report, hasKey := newState.ReportsToCreate[reportKey]
	if !hasKey {
		report = types.Report{
			Key:           reportKey,
			ParticipantID: oldState.PState.ParticipantID,
			Timestamp:     time.Now().Truncate(time.Minute).Unix(),
		}
	}

	// Get attribute Key
	a, err := EvalContext.expressionArgResolver(action.Data[1])
	if err != nil {
		logger.Error.Printf("unexpected error: %v", err)
		return newState, err
	}
	attributeKey, ok1 := a.(string)
	if !ok1 {
		return newState, errors.New("could not parse arguments")
	}

	// Get value
	v, err := EvalContext.expressionArgResolver(action.Data[2])
	if err != nil {
		logger.Debug.Printf("value couldn't be retrieved, skipping update: %v", err)
		return newState, err
	}

	dType := ""
	if len(action.Data) > 3 {
		// Set dtype
		d, err := EvalContext.expressionArgResolver(action.Data[3])
		if err != nil {
			return newState, err
		}
		dtype, ok1 := d.(string)
		if !ok1 {
			return newState, errors.New("could not parse arguments")
		}
		dType = dtype
	}

	value := ""
	switch flagVal := v.(type) {
	case string:
		value = flagVal
	case float64:
		if dType == "int" {
			value = fmt.Sprintf("%d", int(flagVal))
		} else {
			value = fmt.Sprintf("%f", flagVal)
		}
	case bool:
		value = fmt.Sprintf("%t", flagVal)
	}

	newData := types.ReportData{
		Key:   attributeKey,
		Value: value,
		Dtype: dType,
	}

	index := -1
	for i, d := range report.Data {
		if d.Key == attributeKey {
			index = i
			break
		}
	}

	if index < 0 {
		report.Data = append(report.Data, newData)
	} else {
		report.Data[index] = newData
	}

	newState.ReportsToCreate[reportKey] = report
	return
}

// remove one data entry in the report with the key
func removeReportData(action types.Expression, oldState ActionData, event types.StudyEvent, configs ActionConfigs) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) != 2 {
		return newState, errors.New("removeReportData must have exactly two arguments")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
		Configs:          configs,
	}
	k, err := EvalContext.expressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}

	reportKey, ok1 := k.(string)
	if !ok1 {
		return newState, errors.New("could not parse arguments")
	}

	// If report not initialized yet, init report:
	report, hasKey := newState.ReportsToCreate[reportKey]
	if !hasKey {
		// nothing to do
		return newState, nil
	}

	// Get attribute Key
	a, err := EvalContext.expressionArgResolver(action.Data[1])
	if err != nil {
		return newState, err
	}
	attributeKey, ok1 := a.(string)
	if !ok1 {
		return newState, errors.New("could not parse arguments")
	}

	index := -1
	for i, d := range report.Data {
		if d.Key == attributeKey {
			index = i
			break
		}
	}

	if index > -1 {
		report.Data = append(report.Data[:index], report.Data[index+1:]...)
	}

	newState.ReportsToCreate[reportKey] = report
	return
}

// remove the report from this event
func cancelReport(action types.Expression, oldState ActionData, event types.StudyEvent, configs ActionConfigs) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) != 1 {
		return newState, errors.New("updateReportData must have exactly 1 argument")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
		Configs:          configs,
	}
	k, err := EvalContext.expressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}

	reportKey, ok1 := k.(string)
	if !ok1 {
		return newState, errors.New("could not parse arguments")
	}

	_, hasKey := newState.ReportsToCreate[reportKey]
	if hasKey {
		delete(newState.ReportsToCreate, reportKey)
	}
	return
}

// delete confidential responses for this participant for a particular key only
func removeConfidentialResponseByKey(action types.Expression, oldState ActionData, event types.StudyEvent, configs ActionConfigs) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) != 1 {
		return newState, errors.New("removeConfidentialResponseByKey must have exactly 1 argument")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
		Configs:          configs,
	}
	k, err := EvalContext.expressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}

	key, ok1 := k.(string)
	if !ok1 {
		return newState, errors.New("could not parse arguments")
	}

	_, err = configs.DBService.DeleteConfidentialResponses(event.InstanceID, event.StudyKey, event.ParticipantIDForConfidentialResponses, key)
	if err != nil {
		logger.Error.Printf("unexpected error: %v", err)
	}
	return
}

// delete confidential responses for this participant
func removeAllConfidentialResponses(action types.Expression, oldState ActionData, event types.StudyEvent, configs ActionConfigs) (newState ActionData, err error) {
	newState = oldState
	_, err = configs.DBService.DeleteConfidentialResponses(event.InstanceID, event.StudyKey, event.ParticipantIDForConfidentialResponses, "")
	if err != nil {
		logger.Error.Printf("unexpected error: %v", err)
	}
	return
}

// call external service to handle event
func externalEventHandler(action types.Expression, oldState ActionData, event types.StudyEvent, configs ActionConfigs) (newState ActionData, err error) {
	newState = oldState

	if len(action.Data) != 1 {
		msg := "externalEventHandler must have exactly 1 argument"
		logger.Error.Printf(msg)
		return newState, errors.New(msg)
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
		Configs:          configs,
	}
	k, err := EvalContext.expressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}

	serviceName, ok1 := k.(string)
	if !ok1 {
		return newState, errors.New("could not parse arguments")
	}

	serviceConfig, err := getExternalServicesConfigByName(configs.ExternalServiceConfigs, serviceName)
	if err != nil {
		logger.Error.Println(err)
		return newState, err
	}

	payload := ExternalEventPayload{
		ParticipantState: newState.PState,
		EventType:        event.Type,
		StudyKey:         event.StudyKey,
		InstanceID:       event.InstanceID,
		Response:         event.Response,
	}
	response, err := runHTTPcall(serviceConfig.URL, serviceConfig.APIKey, payload)
	if err != nil {
		logger.Error.Printf("error when handling response for '%s': %v", serviceName, err)
		return newState, err
	}

	// if relevant, update participant state:
	pState, hasKey := response["pState"]
	if hasKey {
		newState.PState = pState.(types.ParticipantState)
		logger.Debug.Printf("updated participant state from external service")
	}

	// collect reports if any:
	reportsToCreate, hasKey := response["reportsToCreate"]
	if hasKey {
		reportsToCreate := reportsToCreate.(map[string]types.Report)
		for key, value := range reportsToCreate {
			newState.ReportsToCreate[key] = value
		}
		logger.Debug.Printf("updated reports list from external service")
	}
	return
}
