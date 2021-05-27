package exporter

import (
	"bytes"
	"testing"

	studyAPI "github.com/influenzanet/study-service/pkg/api"
)

/*
"likertGroup"
"eq5d-health-indicator"
"multipleChoiceGroup"
"singleChoiceGroup"

"responseGroup"
*/

func TestResponseExporter(t *testing.T) {
	testLang := "en"
	questionOptionSep := "-"
	testSurveyDef := &studyAPI.SurveyItem{
		Key: "weekly",
		Items: []*studyAPI.SurveyItem{
			mockQuestion("weekly.Q1", testLang, "Title of Q1", mockSingleChoiceGroup(testLang, []MockOpionDef{
				{Key: "1", Role: "option", Label: "Yes"},
				{Key: "2", Role: "option", Label: "No"},
				{Key: "3", Role: "input", Label: "Other"},
			})),
			mockQuestion("weekly.Q2", testLang, "Title of Q2", mockMultipleChoiceGroup(testLang, []MockOpionDef{
				{Key: "1", Role: "option", Label: "Option 1"},
				{Key: "2", Role: "option", Label: "Option 2"},
				{Key: "3", Role: "input", Label: "Other"},
			})),
			{Key: "weekly.G1", Items: []*studyAPI.SurveyItem{
				mockQuestion("weekly.G1.Q1", testLang, "Title of Group 1's Q1", mockLikertGroup(testLang, []MockOpionDef{
					{Key: "cat1", Label: "Category 1"},
					{Key: "cat2", Label: "Category 2"},
				}, []string{
					"o1", "o2", "o3",
				})),
			}},
		},
	}

	t.Run("with with missing surveyDef", func(t *testing.T) {
		_, err := NewResponseExporter(nil, "en", true, questionOptionSep)
		if err == nil {
			t.Error("error expected")
			return
		}
		if err.Error() != "current survey definition not found" {
			t.Errorf("unexpected error message: %v", err)
			return
		}
	})

	t.Run("with with missing current", func(t *testing.T) {
		testSurvey := studyAPI.Survey{
			Id:      "test-id",
			Current: nil,
			History: []*studyAPI.SurveyVersion{},
		}

		_, err := NewResponseExporter(&testSurvey, "en", true, questionOptionSep)
		if err == nil {
			t.Error("error expected")
			return
		}
		if err.Error() != "current survey definition not found" {
			t.Errorf("unexpected error message: %v", err)
			return
		}
	})

	t.Run("with with one version", func(t *testing.T) {
		testSurvey := studyAPI.Survey{
			Id: "test-id",
			Current: &studyAPI.SurveyVersion{
				Published:        10,
				VersionId:        "1",
				SurveyDefinition: testSurveyDef,
			},
		}

		rp, err := NewResponseExporter(&testSurvey, "en", true, questionOptionSep)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if len(rp.surveyVersions) != 1 {
			t.Errorf("unexpected number of versions: %d", len(rp.surveyVersions))
			return
		}

		if len(rp.surveyVersions[0].Questions) != 3 {
			t.Errorf("unexpected number of versions: %d", len(rp.surveyVersions[0].Questions))
			return
		}
	})

	t.Run("with with multiple versions", func(t *testing.T) {
		testSurvey := studyAPI.Survey{
			Id: "test-id",
			Current: &studyAPI.SurveyVersion{
				Published:        10,
				VersionId:        "3",
				SurveyDefinition: testSurveyDef,
			},
			History: []*studyAPI.SurveyVersion{
				{
					Published:        2,
					Unpublished:      5,
					VersionId:        "1",
					SurveyDefinition: testSurveyDef,
				},
				{
					Published:        5,
					Unpublished:      10,
					VersionId:        "2",
					SurveyDefinition: testSurveyDef,
				},
			},
		}

		rp, err := NewResponseExporter(&testSurvey, "en", true, questionOptionSep)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if len(rp.surveyVersions) != 3 {
			t.Errorf("unexpected number of versions: %d", len(rp.surveyVersions))
			return
		}
	})
}

func TestGetResponseCSV(t *testing.T) {
	testLang := "en"
	questionOptionSep := "-"
	testSurveyDef := &studyAPI.SurveyItem{
		Key: "weekly",
		Items: []*studyAPI.SurveyItem{
			mockQuestion("weekly.Q1", testLang, "Title of Q1", mockSingleChoiceGroup(testLang, []MockOpionDef{
				{Key: "1", Role: "option", Label: "Yes"},
				{Key: "2", Role: "option", Label: "No"},
				{Key: "3", Role: "input", Label: "Other"},
			})),
			mockQuestion("weekly.Q2", testLang, "Title of Q2", mockMultipleChoiceGroup(testLang, []MockOpionDef{
				{Key: "1", Role: "option", Label: "Option 1"},
				{Key: "2", Role: "option", Label: "Option 2"},
				{Key: "3", Role: "input", Label: "Other"},
			})),
			{Key: "weekly.G1", Items: []*studyAPI.SurveyItem{
				mockQuestion("weekly.G1.Q1", testLang, "Title of Group 1's Q1", mockLikertGroup(testLang, []MockOpionDef{
					{Key: "cat1", Label: "Category 1"},
					{Key: "cat2", Label: "Category 2"},
				}, []string{
					"o1", "o2", "o3",
				})),
			}},
		},
	}
	testSurvey := studyAPI.Survey{
		Id: "surveyIDfromDB",
		Current: &studyAPI.SurveyVersion{
			Published:        10,
			VersionId:        "3",
			SurveyDefinition: testSurveyDef,
		},
	}
	parser, err := NewResponseExporter(&testSurvey, "en", true, questionOptionSep)
	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
		return
	}

	t.Run("with no responses added yet", func(t *testing.T) {
		buf := new(bytes.Buffer)
		err := parser.GetResponsesCSV(buf, nil)
		if err == nil {
			t.Error("should produce error")
		}
	})

	err = parser.AddResponse(&studyAPI.SurveyResponse{Key: "weekly", ParticipantId: "part1", SubmittedAt: 10,
		Context: map[string]string{
			"engineVersion": "v0923",
			"language":      "en",
		},
		VersionId: "3",
		Responses: []*studyAPI.SurveyItemResponse{
			{Key: "weekly.Q1", Meta: &studyAPI.ResponseMeta{}, Response: &studyAPI.ResponseItem{
				Key: "rg",
				Items: []*studyAPI.ResponseItem{
					{Key: "scg", Items: []*studyAPI.ResponseItem{
						{Key: "3", Value: "hello, \"how are you\""},
					}},
				},
			}},
			{Key: "weekly.Q2", Meta: &studyAPI.ResponseMeta{}, Response: &studyAPI.ResponseItem{
				Key: "rg",
				Items: []*studyAPI.ResponseItem{
					{Key: "mcg", Items: []*studyAPI.ResponseItem{
						{Key: "1"},
					}},
				},
			}},
			{Key: "weekly.G1.Q1", Meta: &studyAPI.ResponseMeta{}, Response: &studyAPI.ResponseItem{
				Key: "rg",
				Items: []*studyAPI.ResponseItem{
					{Key: "cat1", Items: []*studyAPI.ResponseItem{
						{Key: "o1"},
					}},
					{Key: "cat2", Items: []*studyAPI.ResponseItem{
						{Key: "o2"},
					}},
				},
			}},
		},
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
		return
	}

	t.Run("with one response added", func(t *testing.T) {
		buf := new(bytes.Buffer)
		err := parser.GetResponsesCSV(buf, nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if buf.String() != "" {
			return
		}
	})

}
