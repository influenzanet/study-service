package models

import api "github.com/influenzanet/study-service/api"

type SurveyItemResponse struct {
	Key  string       `bson:"key"`
	Meta ResponseMeta `bson:"meta"`
	// for item groups:
	Items []SurveyItemResponse `bson:"items"`
	// for single items:
	Response ResponseValue `bson:"response"`
}

func (sir SurveyItemResponse) ToAPI() *api.SurveyItemResponse {
	items := make([]*api.SurveyItemResponse, len(sir.Items))
	for i, si := range sir.Items {
		items[i] = si.ToAPI()
	}
	return &api.SurveyItemResponse{
		Key:      sir.Key,
		Meta:     sir.Meta.ToAPI(),
		Items:    items,
		Response: sir.Response.ToAPI(),
	}
}

func SurveyItemResponseFromAPI(sir *api.SurveyItemResponse) SurveyItemResponse {
	if sir == nil {
		return SurveyItemResponse{}
	}
	items := make([]SurveyItemResponse, len(sir.Items))
	for i, si := range sir.Items {
		items[i] = SurveyItemResponseFromAPI(si)
	}
	return SurveyItemResponse{
		Key:      sir.Key,
		Meta:     ResponseMetaFromAPI(sir.Meta),
		Items:    items,
		Response: ResponseValueFromAPI(sir.Response),
	}
}

// ResponseValue
type ResponseValue struct {
	Key   string `bson:"key"`
	Value string `bson:"value"`
	Dtype string `bson:"dtype"`
	// For response option groups:
	Items []ResponseValue `bson:"items"`
}

func (rv ResponseValue) ToAPI() *api.ResponseValue {
	items := make([]*api.ResponseValue, len(rv.Items))
	for i, si := range rv.Items {
		items[i] = si.ToAPI()
	}
	return &api.ResponseValue{
		Key:   rv.Key,
		Value: rv.Value,
		Dtype: rv.Dtype,
		Items: items,
	}
}

func ResponseValueFromAPI(rv *api.ResponseValue) ResponseValue {
	if rv == nil {
		return ResponseValue{}
	}
	items := make([]ResponseValue, len(rv.Items))
	for i, si := range rv.Items {
		items[i] = ResponseValueFromAPI(si)
	}
	return ResponseValue{
		Key:   rv.Key,
		Value: rv.Value,
		Dtype: rv.Dtype,
		Items: items,
	}
}

// ResponseMeta
type ResponseMeta struct {
	Position   int32  `bson:"position"`
	LocaleCode string `bson:"localeCode"`
	Version    int32  `bson:"version"`
	// timestamps:
	Rendered  []int64 `bson:"rendered"`
	Displayed []int64 `bson:"displayed"`
	Responded []int64 `bson:"responded"`
}

func (rm ResponseMeta) ToAPI() *api.ResponseMeta {
	return &api.ResponseMeta{
		Position:   rm.Position,
		LocaleCode: rm.LocaleCode,
		Version:    rm.Version,
		Rendered:   rm.Rendered,
		Displayed:  rm.Displayed,
		Responded:  rm.Responded,
	}

}

func ResponseMetaFromAPI(rm *api.ResponseMeta) ResponseMeta {
	if rm == nil {
		return ResponseMeta{}
	}
	return ResponseMeta{
		Position:   rm.Position,
		LocaleCode: rm.LocaleCode,
		Version:    rm.Version,
		Rendered:   rm.Rendered,
		Displayed:  rm.Displayed,
		Responded:  rm.Responded,
	}
}
