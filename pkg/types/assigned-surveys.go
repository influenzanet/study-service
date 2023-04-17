package types

import api "github.com/influenzanet/study-service/pkg/api"

const (
	ASSIGNED_SURVEY_CATEGORY_PRIO   = "prio"
	ASSIGNED_SURVEY_CATEGORY_NORMAL = "normal"
	ASSIGNED_SURVEY_CATEGORY_QUICK  = "quick"
	ASSIGNED_SURVEY_CATEGORY_UPDATE = "update"
)

// AssignedSurvey is an object about what surveys are assigned to the participant
type AssignedSurvey struct {
	SurveyKey  string `bson:"surveyKey"`  // reference to the survey object
	ValidFrom  int64  `bson:"validFrom"`  // survey should be only visible after this timestamp
	ValidUntil int64  `bson:"validUntil"` // survey should be submitted before this timestamp
	Category   string `bson:"category"`   // how to display the survey (see ASSIGNED_SURVEY_CATEGORY_* constants)
}

func (as AssignedSurvey) ToAPI() *api.AssignedSurvey {
	return &api.AssignedSurvey{
		SurveyKey:  as.SurveyKey,
		ValidFrom:  as.ValidFrom,
		ValidUntil: as.ValidUntil,
		Category:   as.Category,
	}
}

func AssignedSurveyFromAPI(as *api.AssignedSurvey) AssignedSurvey {
	if as == nil {
		return AssignedSurvey{}
	}
	return AssignedSurvey{
		SurveyKey:  as.SurveyKey,
		ValidFrom:  as.ValidFrom,
		ValidUntil: as.ValidUntil,
		Category:   as.Category,
	}
}
