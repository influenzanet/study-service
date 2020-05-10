package types

import api "github.com/influenzanet/study-service/pkg/api"

// Validation
type Validation struct {
	Key  string     `bson:"key"`
	Type string     `bson:"type"`
	Rule Expression `bson:"expression"`
}

func (v Validation) ToAPI() *api.Validation {
	return &api.Validation{
		Key:  v.Key,
		Type: v.Type,
		Rule: v.Rule.ToAPI(),
	}
}

func ValidationFromAPI(v *api.Validation) Validation {
	if v == nil {
		return Validation{}
	}
	return Validation{
		Key:  v.Key,
		Type: v.Type,
		Rule: *ExpressionFromAPI(v.Rule),
	}
}
