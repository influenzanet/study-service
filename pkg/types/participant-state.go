package types

import "go.mongodb.org/mongo-driver/bson/primitive"

const (
	PARTICIPANT_STUDY_STATUS_ACTIVE = "active"
	PARTICIPANT_STUDY_STATUS_EXITED = "exited"
)

// ParticipantState defines the datamodel for current state of the participant in a study as stored in the database
type ParticipantState struct {
	ID              primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	ParticipantID   string               `bson:"participantID"` // reference to the study specific participant ID
	EnteredAt       int64                `bson:"enteredAt"`
	StudyStatus     string               `bson:"studyStatus"` // shows if participant is active in the study - possible values: "active", "inactive", "paused", "finished"
	Flags           map[string]string    `bson:"flags"`
	AssignedSurveys []AssignedSurvey     `bson:"assignedSurveys"`
	Reports         []SurveyItemResponse `bson:"reports"`
	LastSubmissions map[string]int64     `bson:"lastSubmission"` // surveyKey with timestamp
}
