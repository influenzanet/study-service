package models

import api "github.com/influenzanet/study-service/api"

type SurveyContext struct {
	Mode              string
	PreviousResponses []SurveyResponse
}

type SurveyContextDef struct {
	SurveyKey string `bson:"surveyKey"`

	Mode              ExpressionArg `bson:"mode"`
	PreviousResponses []Expression  `bson:"previousResponses"`
}

func (cd SurveyContextDef) ToAPI() *api.SurveyContextDef {
	previous := make([]*api.Expression, len(cd.PreviousResponses))
	for i, r := range cd.PreviousResponses {
		previous[i] = r.ToAPI()
	}
	return &api.SurveyContextDef{
		Mode:              cd.Mode.ToAPI(),
		PreviousResponses: previous,
	}
}

func SurveyContextDefFromAPI(cd *api.SurveyContextDef) SurveyContextDef {
	if cd == nil {
		return SurveyContextDef{}
	}
	previous := make([]Expression, len(cd.PreviousResponses))
	for i, r := range cd.PreviousResponses {
		previous[i] = ExpressionFromAPI(r)
	}
	return SurveyContextDef{
		Mode:              ExpressionArgFromAPI(cd.Mode),
		PreviousResponses: previous,
	}
}
