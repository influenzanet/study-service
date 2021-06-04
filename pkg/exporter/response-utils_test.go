package exporter

import (
	"testing"

	studyAPI "github.com/influenzanet/study-service/pkg/api"
)

func TestFindSurveyVersion(t *testing.T) {
	t.Run("with no versions available", func(t *testing.T) {
		_, err := findSurveyVersion("id1", 100, []SurveyVersionPreview{})
		if err == nil {
			t.Error("should fail with error")
		}
	})

	testVersions := []SurveyVersionPreview{
		{VersionID: "id1", Published: 0, Unpublished: 50},
		{VersionID: "id2", Published: 50, Unpublished: 120},
		{VersionID: "id3", Published: 120, Unpublished: 0},
	}

	t.Run("with versionID empty - has no matching version based on timestamp", func(t *testing.T) {
		_, err := findSurveyVersion("", -10, testVersions)
		if err == nil {
			t.Error("should fail with error")
		}
	})

	t.Run("with versionID but no matching version", func(t *testing.T) {
		_, err := findSurveyVersion("otherID", -1, testVersions)
		if err == nil {
			t.Error("should fail with error")
		}
	})

	t.Run("with versionID empty - has matching version based on timestamp", func(t *testing.T) {
		sv, err := findSurveyVersion("", 100, testVersions)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if sv.VersionID != "id2" {
			t.Errorf("unexpected version: %v", sv)
		}
	})

	t.Run("with versionID but no matching version but has matching version based on timestamp", func(t *testing.T) {
		sv, err := findSurveyVersion("otherID", 100, testVersions)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if sv.VersionID != "id2" {
			t.Errorf("unexpected version: %v", sv)
		}
	})

	t.Run("with versionID simply", func(t *testing.T) {
		sv, err := findSurveyVersion("id2", 100, testVersions)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if sv.VersionID != "id2" {
			t.Errorf("unexpected version: %v", sv)
		}
	})
}

func TestGetResponseColumns(t *testing.T) {
	questionOptionSep := "-"

	t.Run("QUESTION_TYPE_EMPTY", func(t *testing.T) {
		cols := getResponseColumns(SurveyQuestion{
			ID:           "test",
			QuestionType: QUESTION_TYPE_EMPTY,
			Responses:    []ResponseDef{},
		}, &studyAPI.SurveyItemResponse{
			Key: "test",
		}, questionOptionSep)
		if len(cols) > 0 {
			t.Errorf("unexpected results: %v", cols)
		}
	})

	t.Run("QUESTION_TYPE_SINGLE_CHOICE with response", func(t *testing.T) {
		cols := getResponseColumns(SurveyQuestion{
			ID:           "test",
			QuestionType: QUESTION_TYPE_SINGLE_CHOICE,
			Responses: []ResponseDef{
				{ID: "scg", ResponseType: QUESTION_TYPE_SINGLE_CHOICE, Options: []ResponseOption{
					{ID: "1", OptionType: OPTION_TYPE_TEXT_INPUT},
					{ID: "2", OptionType: OPTION_TYPE_RADIO},
				}},
			},
		}, &studyAPI.SurveyItemResponse{
			Key: "test",
			Response: &studyAPI.ResponseItem{
				Key: "rg",
				Items: []*studyAPI.ResponseItem{
					{Key: "scg",
						Items: []*studyAPI.ResponseItem{
							{Key: "1", Value: "hello"},
						},
					},
				},
			},
		}, questionOptionSep)
		if len(cols) != 2 {
			t.Errorf("unexpected results: %v", cols)
		}
		if cols["test"] != "1" {
			t.Errorf("unexpected results: %v", cols)
		}
	})

	t.Run("QUESTION_TYPE_SINGLE_CHOICE without response", func(t *testing.T) {
		cols := getResponseColumns(SurveyQuestion{
			ID:           "test",
			QuestionType: QUESTION_TYPE_SINGLE_CHOICE,
			Responses: []ResponseDef{
				{ID: "scg", ResponseType: QUESTION_TYPE_SINGLE_CHOICE, Options: []ResponseOption{
					{ID: "1", OptionType: OPTION_TYPE_RADIO},
					{ID: "2", OptionType: OPTION_TYPE_RADIO},
					{ID: "3", OptionType: OPTION_TYPE_TEXT_INPUT},
					{ID: "4", OptionType: OPTION_TYPE_DATE_INPUT},
				}},
			},
		}, nil, questionOptionSep)
		if len(cols) != 3 {
			t.Errorf("unexpected results: %v", cols)
		}
	})

	t.Run("QUESTION_TYPE_SINGLE_CHOICE multiple response groups", func(t *testing.T) {
		cols := getResponseColumns(SurveyQuestion{
			ID:           "test",
			QuestionType: QUESTION_TYPE_SINGLE_CHOICE,
			Responses: []ResponseDef{
				{ID: "scg1", ResponseType: QUESTION_TYPE_SINGLE_CHOICE, Options: []ResponseOption{
					{ID: "1", OptionType: OPTION_TYPE_RADIO},
					{ID: "2", OptionType: OPTION_TYPE_RADIO},
					{ID: "3", OptionType: OPTION_TYPE_TEXT_INPUT},
					{ID: "4", OptionType: OPTION_TYPE_DATE_INPUT},
				}},
				{ID: "scg2", ResponseType: QUESTION_TYPE_SINGLE_CHOICE, Options: []ResponseOption{
					{ID: "1", OptionType: OPTION_TYPE_RADIO},
					{ID: "2", OptionType: OPTION_TYPE_RADIO},
					{ID: "3", OptionType: OPTION_TYPE_TEXT_INPUT},
					{ID: "4", OptionType: OPTION_TYPE_DATE_INPUT},
				}},
			},
		}, &studyAPI.SurveyItemResponse{
			Key: "test",
			Response: &studyAPI.ResponseItem{
				Key: "rg",
				Items: []*studyAPI.ResponseItem{
					{Key: "scg1",
						Items: []*studyAPI.ResponseItem{
							{Key: "4", Value: "hello"},
						},
					},
				},
			},
		}, questionOptionSep)
		if len(cols) != 6 {
			t.Errorf("unexpected results: %v", cols)
		}
	})

	t.Run("QUESTION_TYPE_MULTIPLE_CHOICE with response", func(t *testing.T) {
		cols := getResponseColumns(SurveyQuestion{
			ID:           "test",
			QuestionType: QUESTION_TYPE_MULTIPLE_CHOICE,
			Responses: []ResponseDef{
				{ID: "mcg", ResponseType: QUESTION_TYPE_MULTIPLE_CHOICE, Options: []ResponseOption{
					{ID: "1", OptionType: OPTION_TYPE_CHECKBOX},
					{ID: "2", OptionType: OPTION_TYPE_CHECKBOX},
					{ID: "3", OptionType: OPTION_TYPE_TEXT_INPUT},
					{ID: "4", OptionType: OPTION_TYPE_DATE_INPUT},
				}},
			},
		}, &studyAPI.SurveyItemResponse{
			Key: "test",
			Response: &studyAPI.ResponseItem{
				Key: "rg",
				Items: []*studyAPI.ResponseItem{
					{Key: "mcg",
						Items: []*studyAPI.ResponseItem{
							{Key: "1"},
							{Key: "4", Value: "hello"},
						},
					},
				},
			},
		}, questionOptionSep)
		if len(cols) != 6 {
			t.Errorf("unexpected results: %v", cols)
		}
		if cols["test-1"] != "TRUE" {
			t.Errorf("unexpected results: %v", cols)
		}
	})

	t.Run("QUESTION_TYPE_MULTIPLE_CHOICE without response", func(t *testing.T) {
		cols := getResponseColumns(SurveyQuestion{
			ID:           "test",
			QuestionType: QUESTION_TYPE_MULTIPLE_CHOICE,
			Responses: []ResponseDef{
				{ID: "mcg1", ResponseType: QUESTION_TYPE_MULTIPLE_CHOICE, Options: []ResponseOption{
					{ID: "1", OptionType: OPTION_TYPE_CHECKBOX},
					{ID: "2", OptionType: OPTION_TYPE_CHECKBOX},
					{ID: "3", OptionType: OPTION_TYPE_TEXT_INPUT},
					{ID: "4", OptionType: OPTION_TYPE_DATE_INPUT},
				}},
			},
		}, nil, questionOptionSep)
		if len(cols) != 6 {
			t.Errorf("unexpected results: %v", cols)
		}
	})

	t.Run("QUESTION_TYPE_MULTIPLE_CHOICE with group", func(t *testing.T) {
		cols := getResponseColumns(SurveyQuestion{
			ID:           "test",
			QuestionType: QUESTION_TYPE_MULTIPLE_CHOICE,
			Responses: []ResponseDef{
				{ID: "mcg1", ResponseType: QUESTION_TYPE_MULTIPLE_CHOICE, Options: []ResponseOption{
					{ID: "1", OptionType: OPTION_TYPE_CHECKBOX},
					{ID: "2", OptionType: OPTION_TYPE_CHECKBOX},
					{ID: "3", OptionType: OPTION_TYPE_TEXT_INPUT},
					{ID: "4", OptionType: OPTION_TYPE_DATE_INPUT},
				}},
				{ID: "mcg2", ResponseType: QUESTION_TYPE_MULTIPLE_CHOICE, Options: []ResponseOption{
					{ID: "1", OptionType: OPTION_TYPE_CHECKBOX},
					{ID: "2", OptionType: OPTION_TYPE_CHECKBOX},
					{ID: "3", OptionType: OPTION_TYPE_TEXT_INPUT},
					{ID: "4", OptionType: OPTION_TYPE_DATE_INPUT},
				}},
			},
		}, &studyAPI.SurveyItemResponse{
			Key: "test",
			Response: &studyAPI.ResponseItem{
				Key: "rg",
				Items: []*studyAPI.ResponseItem{
					{Key: "mcg1",
						Items: []*studyAPI.ResponseItem{
							{Key: "4", Value: "hello"},
						},
					},
				},
			},
		}, questionOptionSep)
		if len(cols) != 12 {
			t.Errorf("unexpected results: %v", cols)
		}
	})

	t.Run("QUESTION_TYPE_TEXT_INPUT with response", func(t *testing.T) {
		cols := getResponseColumns(SurveyQuestion{
			ID:           "test",
			QuestionType: QUESTION_TYPE_TEXT_INPUT,
			Responses: []ResponseDef{
				{ID: "inp", ResponseType: QUESTION_TYPE_TEXT_INPUT}},
		}, &studyAPI.SurveyItemResponse{
			Key: "test",
			Response: &studyAPI.ResponseItem{
				Key: "rg",
				Items: []*studyAPI.ResponseItem{
					{Key: "inp", Value: "hello"},
				},
			},
		}, questionOptionSep)
		if len(cols) != 1 {
			t.Errorf("unexpected results: %v", cols)
			return
		}
		if cols["test"] != "hello" {
			t.Errorf("unexpected results: %v", cols)
		}
	})

	t.Run("QUESTION_TYPE_TEXT_INPUT without response", func(t *testing.T) {
		cols := getResponseColumns(SurveyQuestion{
			ID:           "test",
			QuestionType: QUESTION_TYPE_TEXT_INPUT,
			Responses: []ResponseDef{
				{ID: "inp1", ResponseType: QUESTION_TYPE_TEXT_INPUT},
				{ID: "inp2", ResponseType: QUESTION_TYPE_TEXT_INPUT},
			},
		}, nil, questionOptionSep)
		if len(cols) != 2 {
			t.Errorf("unexpected results: %v", cols)
			return
		}
	})

	t.Run("QUESTION_TYPE_TEXT_INPUT group", func(t *testing.T) {
		cols := getResponseColumns(SurveyQuestion{
			ID:           "test",
			QuestionType: QUESTION_TYPE_TEXT_INPUT,
			Responses: []ResponseDef{
				{ID: "inp1", ResponseType: QUESTION_TYPE_TEXT_INPUT},
				{ID: "inp2", ResponseType: QUESTION_TYPE_TEXT_INPUT},
			},
		}, &studyAPI.SurveyItemResponse{
			Key: "test",
			Response: &studyAPI.ResponseItem{
				Key: "rg",
				Items: []*studyAPI.ResponseItem{
					{Key: "inp1", Value: "hello"},
				},
			},
		}, questionOptionSep)
		if len(cols) != 2 {
			t.Errorf("unexpected results: %v", cols)
			return
		}
		if cols["test-inp1"] != "hello" {
			t.Errorf("unexpected results: %v", cols)
		}
	})

	t.Run("QUESTION_TYPE_NUMBER_INPUT with response", func(t *testing.T) {
		cols := getResponseColumns(SurveyQuestion{
			ID:           "test",
			QuestionType: QUESTION_TYPE_NUMBER_INPUT,
			Responses: []ResponseDef{
				{ID: "inp", ResponseType: QUESTION_TYPE_NUMBER_INPUT}},
		}, &studyAPI.SurveyItemResponse{
			Key: "test",
			Response: &studyAPI.ResponseItem{
				Key: "rg",
				Items: []*studyAPI.ResponseItem{
					{Key: "inp", Value: "1327", Dtype: "number"},
				},
			},
		}, questionOptionSep)
		if len(cols) != 1 {
			t.Errorf("unexpected results: %v", cols)
			return
		}
		if cols["test"] != "1327" {
			t.Errorf("unexpected results: %v", cols)
		}
	})

	t.Run("QUESTION_TYPE_NUMBER_INPUT without response", func(t *testing.T) {
		cols := getResponseColumns(SurveyQuestion{
			ID:           "test",
			QuestionType: QUESTION_TYPE_NUMBER_INPUT,
			Responses: []ResponseDef{
				{ID: "inp1", ResponseType: QUESTION_TYPE_NUMBER_INPUT},
				{ID: "inp2", ResponseType: QUESTION_TYPE_NUMBER_INPUT},
			},
		}, nil, questionOptionSep)
		if len(cols) != 2 {
			t.Errorf("unexpected results: %v", cols)
			return
		}
	})

	t.Run("QUESTION_TYPE_DATE_INPUT with response", func(t *testing.T) {
		cols := getResponseColumns(SurveyQuestion{
			ID:           "test",
			QuestionType: QUESTION_TYPE_DATE_INPUT,
			Responses: []ResponseDef{
				{ID: "inp", ResponseType: QUESTION_TYPE_DATE_INPUT}},
		}, &studyAPI.SurveyItemResponse{
			Key: "test",
			Response: &studyAPI.ResponseItem{
				Key: "rg",
				Items: []*studyAPI.ResponseItem{
					{Key: "inp", Value: "1327", Dtype: "date"},
				},
			},
		}, questionOptionSep)
		if len(cols) != 1 {
			t.Errorf("unexpected results: %v", cols)
			return
		}
		if cols["test"] != "1327" {
			t.Errorf("unexpected results: %v", cols)
		}
	})

	t.Run("QUESTION_TYPE_DATE_INPUT without response", func(t *testing.T) {
		cols := getResponseColumns(SurveyQuestion{
			ID:           "test",
			QuestionType: QUESTION_TYPE_DATE_INPUT,
			Responses: []ResponseDef{
				{ID: "inp1", ResponseType: QUESTION_TYPE_DATE_INPUT},
				{ID: "inp2", ResponseType: QUESTION_TYPE_DATE_INPUT},
			},
		}, nil, questionOptionSep)
		if len(cols) != 2 {
			t.Errorf("unexpected results: %v", cols)
			return
		}
	})

	t.Run("QUESTION_TYPE_DROPDOWN with response", func(t *testing.T) {
		cols := getResponseColumns(SurveyQuestion{
			ID:           "test",
			QuestionType: QUESTION_TYPE_DROPDOWN,
			Responses: []ResponseDef{
				{ID: "ddg", ResponseType: QUESTION_TYPE_DROPDOWN, Options: []ResponseOption{
					{ID: "1", OptionType: OPTION_TYPE_DROPDOWN_OPTION},
					{ID: "2", OptionType: OPTION_TYPE_DROPDOWN_OPTION},
				}},
			},
		}, &studyAPI.SurveyItemResponse{
			Key: "test",
			Response: &studyAPI.ResponseItem{
				Key: "rg",
				Items: []*studyAPI.ResponseItem{
					{Key: "ddg",
						Items: []*studyAPI.ResponseItem{
							{Key: "1"},
						},
					},
				},
			},
		}, questionOptionSep)
		if len(cols) != 1 {
			t.Errorf("unexpected results: %v", cols)
		}
		if cols["test"] != "1" {
			t.Errorf("unexpected results: %v", cols)
		}
	})

	t.Run("QUESTION_TYPE_DROPDOWN without response", func(t *testing.T) {
		cols := getResponseColumns(SurveyQuestion{
			ID:           "test",
			QuestionType: QUESTION_TYPE_DROPDOWN,
			Responses: []ResponseDef{
				{ID: "ddg", ResponseType: QUESTION_TYPE_DROPDOWN, Options: []ResponseOption{
					{ID: "1", OptionType: OPTION_TYPE_DROPDOWN_OPTION},
					{ID: "2", OptionType: OPTION_TYPE_DROPDOWN_OPTION},
				}},
			},
		}, nil, questionOptionSep)
		if len(cols) != 1 {
			t.Errorf("unexpected results: %v", cols)
		}
	})

	t.Run("QUESTION_TYPE_DROPDOWN group", func(t *testing.T) {
		cols := getResponseColumns(SurveyQuestion{
			ID:           "test",
			QuestionType: QUESTION_TYPE_DROPDOWN,
			Responses: []ResponseDef{
				{ID: "ddg1", ResponseType: QUESTION_TYPE_DROPDOWN, Options: []ResponseOption{
					{ID: "1", OptionType: OPTION_TYPE_DROPDOWN_OPTION},
					{ID: "2", OptionType: OPTION_TYPE_DROPDOWN_OPTION},
				}},
				{ID: "ddg2", ResponseType: QUESTION_TYPE_DROPDOWN, Options: []ResponseOption{
					{ID: "1", OptionType: OPTION_TYPE_DROPDOWN_OPTION},
					{ID: "2", OptionType: OPTION_TYPE_DROPDOWN_OPTION},
				}},
			},
		}, &studyAPI.SurveyItemResponse{
			Key: "test",
			Response: &studyAPI.ResponseItem{
				Key: "rg",
				Items: []*studyAPI.ResponseItem{
					{Key: "ddg1",
						Items: []*studyAPI.ResponseItem{
							{Key: "1"},
						},
					},
				},
			},
		}, questionOptionSep)
		if len(cols) != 2 {
			t.Errorf("unexpected results: %v", cols)
		}
	})

	t.Run("QUESTION_TYPE_LIKERT with response", func(t *testing.T) {
		cols := getResponseColumns(SurveyQuestion{
			ID:           "test",
			QuestionType: QUESTION_TYPE_LIKERT,
			Responses: []ResponseDef{
				{ID: "likert", ResponseType: QUESTION_TYPE_LIKERT, Options: []ResponseOption{
					{ID: "1", OptionType: OPTION_TYPE_RADIO},
					{ID: "2", OptionType: OPTION_TYPE_RADIO},
					{ID: "3", OptionType: OPTION_TYPE_RADIO},
				}},
			},
		}, &studyAPI.SurveyItemResponse{
			Key: "test",
			Response: &studyAPI.ResponseItem{
				Key: "rg",
				Items: []*studyAPI.ResponseItem{
					{Key: "likert",
						Items: []*studyAPI.ResponseItem{
							{Key: "1"},
						},
					},
				},
			},
		}, questionOptionSep)
		if len(cols) != 1 {
			t.Errorf("unexpected results: %v", cols)
		}
		if cols["test"] != "1" {
			t.Errorf("unexpected results: %v", cols)
		}
	})

	t.Run("QUESTION_TYPE_LIKERT group with response", func(t *testing.T) {
		cols := getResponseColumns(SurveyQuestion{
			ID:           "test",
			QuestionType: QUESTION_TYPE_LIKERT,
			Responses: []ResponseDef{
				{ID: "likert1", ResponseType: QUESTION_TYPE_LIKERT, Options: []ResponseOption{
					{ID: "1", OptionType: OPTION_TYPE_RADIO},
					{ID: "2", OptionType: OPTION_TYPE_RADIO},
					{ID: "3", OptionType: OPTION_TYPE_RADIO},
				}},
				{ID: "likert2", ResponseType: QUESTION_TYPE_LIKERT, Options: []ResponseOption{
					{ID: "1", OptionType: OPTION_TYPE_RADIO},
					{ID: "2", OptionType: OPTION_TYPE_RADIO},
					{ID: "3", OptionType: OPTION_TYPE_RADIO},
				}},
			},
		}, &studyAPI.SurveyItemResponse{
			Key: "test",
			Response: &studyAPI.ResponseItem{
				Key: "rg",
				Items: []*studyAPI.ResponseItem{
					{Key: "likert1",
						Items: []*studyAPI.ResponseItem{
							{Key: "1"},
						},
					},
					{Key: "likert2",
						Items: []*studyAPI.ResponseItem{
							{Key: "3"},
						},
					},
				},
			},
		}, questionOptionSep)
		if len(cols) != 2 {
			t.Errorf("unexpected results: %v", cols)
		}
		if cols["test-likert1"] != "1" {
			t.Errorf("unexpected results: %v", cols)
		}
	})

	t.Run("QUESTION_TYPE_LIKERT without response", func(t *testing.T) {
		cols := getResponseColumns(SurveyQuestion{
			ID:           "test",
			QuestionType: QUESTION_TYPE_LIKERT,
			Responses: []ResponseDef{
				{ID: "likert1", ResponseType: QUESTION_TYPE_LIKERT, Options: []ResponseOption{
					{ID: "1", OptionType: OPTION_TYPE_RADIO},
					{ID: "2", OptionType: OPTION_TYPE_RADIO},
					{ID: "3", OptionType: OPTION_TYPE_RADIO},
				}},
				{ID: "likert2", ResponseType: QUESTION_TYPE_LIKERT, Options: []ResponseOption{
					{ID: "1", OptionType: OPTION_TYPE_RADIO},
					{ID: "2", OptionType: OPTION_TYPE_RADIO},
					{ID: "3", OptionType: OPTION_TYPE_RADIO},
				}},
			},
		}, nil, questionOptionSep)
		if len(cols) != 2 {
			t.Errorf("unexpected results: %v", cols)
		}
	})

	t.Run("QUESTION_TYPE_EQ5D_SLIDER with response", func(t *testing.T) {
		cols := getResponseColumns(SurveyQuestion{
			ID:           "test",
			QuestionType: QUESTION_TYPE_EQ5D_SLIDER,
			Responses: []ResponseDef{
				{ID: "inp", ResponseType: QUESTION_TYPE_EQ5D_SLIDER}},
		}, &studyAPI.SurveyItemResponse{
			Key: "test",
			Response: &studyAPI.ResponseItem{
				Key: "rg",
				Items: []*studyAPI.ResponseItem{
					{Key: "inp", Value: "1327"},
				},
			},
		}, questionOptionSep)
		if len(cols) != 1 {
			t.Errorf("unexpected results: %v", cols)
			return
		}
		if cols["test"] != "1327" {
			t.Errorf("unexpected results: %v", cols)
		}
	})

	t.Run("QUESTION_TYPE_EQ5D_SLIDER without response", func(t *testing.T) {
		cols := getResponseColumns(SurveyQuestion{
			ID:           "test",
			QuestionType: QUESTION_TYPE_EQ5D_SLIDER,
			Responses: []ResponseDef{
				{ID: "inp1", ResponseType: QUESTION_TYPE_EQ5D_SLIDER},
				{ID: "inp2", ResponseType: QUESTION_TYPE_EQ5D_SLIDER},
			},
		}, nil, questionOptionSep)
		if len(cols) != 2 {
			t.Errorf("unexpected results: %v", cols)
			return
		}
	})

	t.Run("QUESTION_TYPE_NUMERIC_SLIDER with response", func(t *testing.T) {
		cols := getResponseColumns(SurveyQuestion{
			ID:           "test",
			QuestionType: QUESTION_TYPE_NUMERIC_SLIDER,
			Responses: []ResponseDef{
				{ID: "inp", ResponseType: QUESTION_TYPE_NUMERIC_SLIDER}},
		}, &studyAPI.SurveyItemResponse{
			Key: "test",
			Response: &studyAPI.ResponseItem{
				Key: "rg",
				Items: []*studyAPI.ResponseItem{
					{Key: "inp", Value: "1327"},
				},
			},
		}, questionOptionSep)
		if len(cols) != 1 {
			t.Errorf("unexpected results: %v", cols)
			return
		}
		if cols["test"] != "1327" {
			t.Errorf("unexpected results: %v", cols)
		}
	})

	t.Run("QUESTION_TYPE_NUMERIC_SLIDER without response", func(t *testing.T) {
		cols := getResponseColumns(SurveyQuestion{
			ID:           "test",
			QuestionType: QUESTION_TYPE_NUMERIC_SLIDER,
			Responses: []ResponseDef{
				{ID: "inp1", ResponseType: QUESTION_TYPE_NUMERIC_SLIDER},
				{ID: "inp2", ResponseType: QUESTION_TYPE_NUMERIC_SLIDER},
			},
		}, nil, questionOptionSep)
		if len(cols) != 2 {
			t.Errorf("unexpected results: %v", cols)
			return
		}
	})

	t.Run("QUESTION_TYPE_MATRIX with response", func(t *testing.T) {
		cols := getResponseColumns(SurveyQuestion{
			ID:           "test",
			QuestionType: QUESTION_TYPE_MATRIX,
			Responses: []ResponseDef{
				{ID: "mat.row1", ResponseType: QUESTION_TYPE_MATRIX_RADIO_ROW, Options: []ResponseOption{
					{ID: "1", OptionType: OPTION_TYPE_RADIO},
					{ID: "2", OptionType: OPTION_TYPE_RADIO},
				}},
				{ID: "mat.row2.col1", ResponseType: QUESTION_TYPE_MATRIX_CHECKBOX},
				{ID: "mat.row2.col2", ResponseType: QUESTION_TYPE_MATRIX_DROPDOWN, Options: []ResponseOption{
					{ID: "1", OptionType: OPTION_TYPE_DROPDOWN_OPTION},
					{ID: "2", OptionType: OPTION_TYPE_DROPDOWN_OPTION},
				}},
			},
		}, &studyAPI.SurveyItemResponse{
			Key: "test",
			Response: &studyAPI.ResponseItem{
				Key: "rg",
				Items: []*studyAPI.ResponseItem{
					{Key: "mat", Items: []*studyAPI.ResponseItem{
						{Key: "row1", Items: []*studyAPI.ResponseItem{
							{Key: "1"},
						}},
						{Key: "row2", Items: []*studyAPI.ResponseItem{
							{Key: "col2", Items: []*studyAPI.ResponseItem{
								{Key: "1"},
							}},
						}},
					}},
				},
			},
		}, questionOptionSep)
		if len(cols) != 3 {
			t.Errorf("unexpected results: %v", cols)
			return
		}
		if cols["test-mat.row2.col2"] != "1" {
			t.Errorf("unexpected results: %v", cols)
		}
	})

	t.Run("QUESTION_TYPE_UNKNOWN with response", func(t *testing.T) {
		cols := getResponseColumns(SurveyQuestion{
			ID:           "test",
			QuestionType: QUESTION_TYPE_UNKNOWN,
			Responses: []ResponseDef{
				{ID: "unk1", ResponseType: QUESTION_TYPE_UNKNOWN, Options: []ResponseOption{
					{ID: "1", OptionType: OPTION_TYPE_DATE_INPUT},
					{ID: "2", OptionType: OPTION_TYPE_RADIO},
				}},
				{ID: "unk2", ResponseType: QUESTION_TYPE_UNKNOWN},
			},
		}, &studyAPI.SurveyItemResponse{
			Key: "test",
			Response: &studyAPI.ResponseItem{
				Key: "rg",
				Items: []*studyAPI.ResponseItem{
					{Key: "unk1", Items: []*studyAPI.ResponseItem{
						{Key: "1", Value: "hello"},
					}},
				},
			},
		}, questionOptionSep)
		if len(cols) != 3 {
			t.Errorf("unexpected results: %v", cols)
			return
		}
		if cols["test-unk1.1"] != "hello" {
			t.Errorf("unexpected results: %v", cols)
		}
	})
}

func TestRetrieveResponseItem(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		r := retrieveResponseItem(nil, "")
		if r != nil {
			t.Errorf("unexpected result: %v", r)
		}
	})

	t.Run("retrieve root", func(t *testing.T) {
		r := retrieveResponseItem(&studyAPI.SurveyItemResponse{
			Response: &studyAPI.ResponseItem{
				Key: "rg",
				Items: []*studyAPI.ResponseItem{
					{Key: "input"},
				},
			},
		}, "rg")
		if r == nil {
			t.Error("should find result")
		}
	})

	t.Run("retrieve group", func(t *testing.T) {
		r := retrieveResponseItem(&studyAPI.SurveyItemResponse{
			Response: &studyAPI.ResponseItem{
				Key: "rg",
				Items: []*studyAPI.ResponseItem{
					{Key: "scg", Items: []*studyAPI.ResponseItem{
						{Key: "1"},
						{Key: "2"},
					}},
				},
			},
		}, "rg.scg")
		if r == nil {
			t.Error("should find result")
			return
		}
		if r.Key != "scg" || len(r.Items) != 2 {
			t.Errorf("unexpected result: %v", r)
		}
	})

	t.Run("retrieve item", func(t *testing.T) {
		r := retrieveResponseItem(&studyAPI.SurveyItemResponse{
			Response: &studyAPI.ResponseItem{
				Key: "rg",
				Items: []*studyAPI.ResponseItem{
					{Key: "scg", Items: []*studyAPI.ResponseItem{
						{Key: "1"},
						{Key: "2"},
					}},
				},
			},
		}, "rg.scg.1")
		if r == nil {
			t.Error("should find result")
			return
		}
		if r.Key != "1" || len(r.Items) != 0 {
			t.Errorf("unexpected result: %v", r)
		}
	})

	t.Run("wrong first key", func(t *testing.T) {
		r := retrieveResponseItem(&studyAPI.SurveyItemResponse{
			Response: &studyAPI.ResponseItem{
				Key: "rg",
				Items: []*studyAPI.ResponseItem{
					{Key: "scg", Items: []*studyAPI.ResponseItem{
						{Key: "1"},
						{Key: "2"},
					}},
				},
			},
		}, "wrong.scg.1")
		if r != nil {
			t.Errorf("unexpected result: %v", r)
		}
	})

	t.Run("wrong middle key", func(t *testing.T) {
		r := retrieveResponseItem(&studyAPI.SurveyItemResponse{
			Response: &studyAPI.ResponseItem{
				Key: "rg",
				Items: []*studyAPI.ResponseItem{
					{Key: "scg", Items: []*studyAPI.ResponseItem{
						{Key: "1"},
						{Key: "2"},
					}},
				},
			},
		}, "rg.wrong.1")
		if r != nil {
			t.Errorf("unexpected result: %v", r)
		}
	})

	t.Run("wrong last key", func(t *testing.T) {
		r := retrieveResponseItem(&studyAPI.SurveyItemResponse{
			Response: &studyAPI.ResponseItem{
				Key: "rg",
				Items: []*studyAPI.ResponseItem{
					{Key: "scg", Items: []*studyAPI.ResponseItem{
						{Key: "1"},
						{Key: "2"},
					}},
				},
			},
		}, "rg.scg.wrong")
		if r != nil {
			t.Errorf("unexpected result: %v", r)
		}
	})
}
