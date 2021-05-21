package types

type StudyEvent struct {
	StudyKey string
	Type     string         // what kind of event (TIMER, SUBMISSION, ENTER etc.)
	Response SurveyResponse // if something is submitted during the event is added here
}
