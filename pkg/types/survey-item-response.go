package types

import api "github.com/influenzanet/study-service/pkg/api"

type SurveyItemResponse struct {
	Key  string       `bson:"key,omitempty"`
	Meta ResponseMeta `bson:"meta,omitempty"`
	// for item groups:
	Items []SurveyItemResponse `bson:"items,omitempty"`
	// for single items:
	Response *ResponseItem `bson:"response,omitempty"`
}

func (sir SurveyItemResponse) ToAPI() *api.SurveyItemResponse {
	items := make([]*api.SurveyItemResponse, len(sir.Items))
	for i, si := range sir.Items {
		items[i] = si.ToAPI()
	}
	apiResp := &api.SurveyItemResponse{
		Key:   sir.Key,
		Meta:  sir.Meta.ToAPI(),
		Items: items,
	}
	if sir.Response != nil {
		apiResp.Response = sir.Response.ToAPI()
	}
	return apiResp
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
		Response: ResponseItemFromAPI(sir.Response),
	}
}

// ResponseItem
type ResponseItem struct {
	Key   string `bson:"key,omitempty"`
	Value string `bson:"value,omitempty"`
	Dtype string `bson:"dtype,omitempty"`
	// For response option groups:
	Items []ResponseItem `bson:"items,omitempty"`
}

func (rv ResponseItem) ToAPI() *api.ResponseItem {
	items := make([]*api.ResponseItem, len(rv.Items))
	for i, si := range rv.Items {
		items[i] = si.ToAPI()
	}
	return &api.ResponseItem{
		Key:   rv.Key,
		Value: rv.Value,
		Dtype: rv.Dtype,
		Items: items,
	}
}

func ResponseItemFromAPI(rv *api.ResponseItem) *ResponseItem {
	if rv == nil {
		return nil
	}
	items := make([]ResponseItem, len(rv.Items))
	for i, si := range rv.Items {
		items[i] = *ResponseItemFromAPI(si)
	}
	return &ResponseItem{
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
