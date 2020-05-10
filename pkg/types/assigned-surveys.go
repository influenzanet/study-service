package types

import api "github.com/influenzanet/study-service/pkg/api"

// AssignedSurvey is an object about what surveys are assigned to the participant
type AssignedSurvey struct {
	SurveyKey  string `bson:"surveyKey"`  // reference to the survey object
	ValidFrom  int64  `bson:"validFrom"`  // survey should be only visible after this timestamp
	ValidUntil int64  `bson:"validUntil"` // survey should be submitted before this timestamp
}

func (as AssignedSurvey) ToAPI() *api.AssignedSurvey {
	return &api.AssignedSurvey{
		SurveyKey:  as.SurveyKey,
		ValidFrom:  as.ValidFrom,
		ValidUntil: as.ValidUntil,
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
	}
}
