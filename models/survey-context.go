package models

type SurveyContext struct {
	Mode              string
	PreviousResponses []SurveyResponse
}
