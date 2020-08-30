package types

import api "github.com/influenzanet/study-service/pkg/api"

type SurveyContext struct {
	Mode              string
	PreviousResponses []SurveyResponse
	ParticipantFlags  map[string]string
}

func (ctx SurveyContext) ToAPI() *api.SurveyContext {
	previous := make([]*api.SurveyResponse, len(ctx.PreviousResponses))
	for i, r := range ctx.PreviousResponses {
		previous[i] = r.ToAPI()
	}
	return &api.SurveyContext{
		Mode:              ctx.Mode,
		PreviousResponses: previous,
		ParticipantFlags:  ctx.ParticipantFlags,
	}
}

type SurveyContextDef struct {
	Mode              *ExpressionArg `bson:"mode"`
	PreviousResponses []Expression   `bson:"previousResponses"`
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

func SurveyContextDefFromAPI(cd *api.SurveyContextDef) *SurveyContextDef {
	if cd == nil {
		return nil
	}
	previous := make([]Expression, len(cd.PreviousResponses))
	for i, r := range cd.PreviousResponses {
		previous[i] = *ExpressionFromAPI(r)
	}
	return &SurveyContextDef{
		Mode:              ExpressionArgFromAPI(cd.Mode),
		PreviousResponses: previous,
	}
}
