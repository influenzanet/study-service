package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// ParticipantState defines the datamodel for current state of the participant in a study as stored in the database
type ParticipantState struct {
	ID              primitive.ObjectID    `bson:"_id,omitempty" json:"id,omitempty"`
	ParticipantID   string                `bson:"participantID"` // reference to the study specific participant ID
	Flags           ParticipantStateFlags `bson:"flags"`
	AssignedSurveys []AssignedSurvey      `bson:"assignedSurveys"`
	Reports         []SurveyItemResponse  `bson:"reports"`
	LastSubmissions map[string]int64      `bson:"lastSubmission"` // surveyKey with timestamp
}

// ParticipantStateFlags provide a quick overview of general state infos about the user
type ParticipantStateFlags struct {
	Status string `bson:"status"` // shows if participant is active in the study - possible values: "active", "inactive", "paused"
}
