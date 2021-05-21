package types

type StudyEvent struct {
	InstanceID string
	StudyKey   string
	Type       string         // what kind of event (TIMER, SUBMISSION, ENTER etc.)
	Response   SurveyResponse // if something is submitted during the event is added here
}
