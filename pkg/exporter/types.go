package exporter

import "github.com/influenzanet/study-service/pkg/api"

const (
	QUESTION_TYPE_SINGLE_CHOICE                   = "single_choice"
	QUESTION_TYPE_MULTIPLE_CHOICE                 = "multiple_choice"
	QUESTION_TYPE_TEXT_INPUT                      = "text"
	QUESTION_TYPE_NUMBER_INPUT                    = "number"
	QUESTION_TYPE_DATE_INPUT                      = "date"
	QUESTION_TYPE_DROPDOWN                        = "dropdown"
	QUESTION_TYPE_LIKERT                          = "likert"
	QUESTION_TYPE_LIKERT_GROUP                    = "likert_group"
	QUESTION_TYPE_EQ5D_SLIDER                     = "eq5d_slider"
	QUESTION_TYPE_NUMERIC_SLIDER                  = "slider"
	QUESTION_TYPE_MATRIX                          = "matrix"
	QUESTION_TYPE_MATRIX_RADIO_ROW                = "matrix_radio_row"
	QUESTION_TYPE_MATRIX_DROPDOWN                 = "matrix_dropdown"
	QUESTION_TYPE_MATRIX_INPUT                    = "matrix_input"
	QUESTION_TYPE_MATRIX_NUMBER_INPUT             = "matrix_number_input"
	QUESTION_TYPE_MATRIX_CHECKBOX                 = "matrix_checkbox"
	QUESTION_TYPE_RESPONSIVE_SINGLE_CHOICE_ARRAY  = "responsive_single_choice_array"
	QUESTION_TYPE_RESPONSIVE_BIPOLAR_LIKERT_ARRAY = "responsive_bipolar_likert_array"
	QUESTION_TYPE_UNKNOWN                         = "unknown"
	QUESTION_TYPE_EMPTY                           = "empty"
)

const (
	OPTION_TYPE_DROPDOWN_OPTION = "option"
	OPTION_TYPE_RADIO           = "radio"
	OPTION_TYPE_CHECKBOX        = "checkbox"
	OPTION_TYPE_TEXT_INPUT      = "text"
	OPTION_TYPE_DATE_INPUT      = "date"
	OPTION_TYPE_NUMBER_INPUT    = "number"
)

const (
	RESPONSE_ROOT_KEY = "rg"
)

const (
	OPEN_FIELD_COL_SUFFIX = "open"
	TRUE_VALUE            = "TRUE"
	FALSE_VALUE           = "FALSE"
)

type SurveyVersionPreview struct {
	VersionID   string
	Published   int64
	Unpublished int64
	Questions   []SurveyQuestion
}

func (sv SurveyVersionPreview) ToAPI() *api.SurveyVersionPreview {
	v := &api.SurveyVersionPreview{
		VersionId:   sv.VersionID,
		Published:   sv.Published,
		Unpublished: sv.Unpublished,
	}
	v.Questions = make([]*api.SurveyQuestionPreview, len(sv.Questions))
	for i, r := range sv.Questions {
		v.Questions[i] = r.ToAPI()
	}
	return v
}

type SurveyQuestion struct {
	ID           string
	Title        string
	QuestionType string
	Responses    []ResponseDef
}

func (q SurveyQuestion) ToAPI() *api.SurveyQuestionPreview {
	v := &api.SurveyQuestionPreview{
		Key:          q.ID,
		QuestionType: q.QuestionType,
		Title:        q.Title,
	}
	v.Responses = make([]*api.ResponseDefPreview, len(q.Responses))
	for i, r := range q.Responses {
		v.Responses[i] = r.ToAPI()
	}
	return v
}

type ResponseDef struct {
	ID           string
	ResponseType string
	Label        string
	Options      []ResponseOption
}

func (rd ResponseDef) ToAPI() *api.ResponseDefPreview {
	v := &api.ResponseDefPreview{
		Key:           rd.ID,
		ResponseTypes: rd.ResponseType,
		Label:         rd.Label,
	}
	v.Options = make([]*api.ResponseOptionPreview, len(rd.Options))
	for i, o := range rd.Options {
		v.Options[i] = o.ToAPI()
	}
	return v
}

type ResponseOption struct {
	ID         string
	OptionType string
	Label      string
}

func (ro ResponseOption) ToAPI() *api.ResponseOptionPreview {
	return &api.ResponseOptionPreview{
		Key:        ro.ID,
		OptionType: ro.OptionType,
		Label:      ro.Label,
	}
}

type ParsedResponse struct {
	ID            string
	ParticipantID string
	OpenedAt      int64
	SubmittedAt   int64
	Version       string
	Context       map[string]string // e.g. Language, or engine version
	Responses     map[string]string
	Meta          ResponseMeta
}

type ResponseMeta struct {
	Initialised map[string]string
	Displayed   map[string]string
	Responded   map[string]string
	ItemVersion map[string]string
	Position    map[string]int32
}

type IncludeMeta struct {
	Postion        bool
	ItemVersion    bool
	InitTimes      bool
	DisplayedTimes bool
	ResponsedTimes bool
}
