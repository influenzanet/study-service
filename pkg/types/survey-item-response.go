package types

import api "github.com/influenzanet/study-service/pkg/api"

type SurveyItemResponse struct {
	Key  string       `bson:"key,omitempty" json:"key"`
	Meta ResponseMeta `bson:"meta,omitempty" json:"meta"`
	// for item groups:
	Items []SurveyItemResponse `bson:"items,omitempty" json:"items,omitempty"`
	// for single items:
	Response         *ResponseItem `bson:"response,omitempty" json:"response,omitempty"`
	ConfidentialMode string        `bson:"confidentialMode,omitempty" json:"confidentialMode,omitempty"`
	MapToKey         string        `bson:"mapToKey,omitempty" json:"mapToKey,omitempty"` // map to this key for confidential mode
}

func (sir SurveyItemResponse) ToAPI() *api.SurveyItemResponse {
	items := make([]*api.SurveyItemResponse, len(sir.Items))
	for i, si := range sir.Items {
		items[i] = si.ToAPI()
	}
	apiResp := &api.SurveyItemResponse{
		Key:              sir.Key,
		Meta:             sir.Meta.ToAPI(),
		Items:            items,
		ConfidentialMode: sir.ConfidentialMode,
		MapToKey:         sir.MapToKey,
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
		Key:              sir.Key,
		Meta:             ResponseMetaFromAPI(sir.Meta),
		Items:            items,
		Response:         ResponseItemFromAPI(sir.Response),
		ConfidentialMode: sir.ConfidentialMode,
		MapToKey:         sir.MapToKey,
	}
}

// ResponseItem
type ResponseItem struct {
	Key   string `bson:"key,omitempty" json:"key,omitempty"`
	Value string `bson:"value,omitempty" json:"value,omitempty"`
	Dtype string `bson:"dtype,omitempty" json:"dtype,omitempty"`
	// For response option groups:
	Items []*ResponseItem `bson:"items,omitempty" json:"items,omitempty"`
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
	items := make([]*ResponseItem, len(rv.Items))
	for i, si := range rv.Items {
		items[i] = ResponseItemFromAPI(si)
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
	Position   int32  `bson:"position" json:"position"`
	LocaleCode string `bson:"localeCode" json:"localeCode"`
	// timestamps:
	Rendered  []int64 `bson:"rendered" json:"rendered"`
	Displayed []int64 `bson:"displayed" json:"displayed"`
	Responded []int64 `bson:"responded" json:"responded"`
}

func (rm ResponseMeta) ToAPI() *api.ResponseMeta {
	return &api.ResponseMeta{
		Position:   rm.Position,
		LocaleCode: rm.LocaleCode,
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
		Rendered:   rm.Rendered,
		Displayed:  rm.Displayed,
		Responded:  rm.Responded,
	}
}
