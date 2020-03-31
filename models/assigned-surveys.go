package models

// AssignedSurvey is an object about what surveys are assigned to the participant
type AssignedSurvey struct {
	SurveyKey  string `bson:"surveyKey"`  // reference to the survey object
	ValidFrom  int64  `bson:"validFrom"`  // survey should be only visible after this timestamp
	ValidUntil int64  `bson:"validUntil"` // survey should be submitted before this timestamp
}
