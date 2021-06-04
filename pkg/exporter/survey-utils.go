package exporter

import (
	"errors"
	"log"
	"strings"

	studyAPI "github.com/influenzanet/study-service/pkg/api"
)

func surveyDefToVersionPreview(original *studyAPI.SurveyVersion, prefLang string) SurveyVersionPreview {
	sp := SurveyVersionPreview{
		VersionID:   original.VersionId,
		Published:   original.Published,
		Unpublished: original.Unpublished,
		Questions:   []SurveyQuestion{},
	}

	sp.Questions = extractQuestions(original.SurveyDefinition, prefLang)
	return sp
}

func extractQuestions(root *studyAPI.SurveyItem, prefLang string) []SurveyQuestion {
	questions := []SurveyQuestion{}
	if root == nil {
		return questions
	}
	for _, item := range root.Items {
		if item.Type == "pageBreak" {
			continue
		}

		if isItemGroup(item) {
			questions = append(questions, extractQuestions(item, prefLang)...)
			continue
		}

		rg := getResponseGroupComponent(item)
		if rg == nil {
			continue
		}

		responses, qType := extractResponses(rg, prefLang)

		titleComp := getTitleComponent(item)
		title := ""
		if titleComp != nil {
			var err error
			title, err = getTranslation(titleComp.Content, prefLang)
			if err != nil {
				log.Printf("Question %s title error: %v", item.Key, err)
			}
		}

		question := SurveyQuestion{
			ID:           item.Key,
			Title:        title,
			QuestionType: qType,
			Responses:    responses,
		}
		questions = append(questions, question)
	}
	return questions
}

func isItemGroup(item *studyAPI.SurveyItem) bool {
	return item != nil && len(item.Items) > 0
}

func getResponseGroupComponent(question *studyAPI.SurveyItem) *studyAPI.ItemComponent {
	if question.Components == nil {
		return nil
	}
	for _, c := range question.Components.Items {
		if c.Role == "responseGroup" {
			return c
		}
	}
	return nil
}

func getTitleComponent(question *studyAPI.SurveyItem) *studyAPI.ItemComponent {
	if question.Components == nil {
		return nil
	}
	for _, c := range question.Components.Items {
		if c.Role == "title" {
			return c
		}
	}
	return nil
}

func getTranslation(content []*studyAPI.LocalisedObject, lang string) (string, error) {
	if len(content) < 1 {
		return "", errors.New("translations missing")
	}

	for _, translation := range content {
		if translation.Code == lang {
			mergedText := ""
			for _, p := range translation.Parts {
				mergedText += p.GetStr()
			}
			return mergedText, nil
		}
	}
	return "", errors.New("translation missing")
}

func extractResponses(rg *studyAPI.ItemComponent, lang string) ([]ResponseDef, string) {
	if rg == nil {
		return []ResponseDef{}, QUESTION_TYPE_EMPTY
	}

	responses := []ResponseDef{}
	for _, item := range rg.Items {
		r := mapToResponseDef(item, rg.Key, lang)
		responses = append(responses, r...)

	}

	qType := getQuestionType(responses)
	return responses, qType

}

func mapToResponseDef(rItem *studyAPI.ItemComponent, parentKey string, lang string) []ResponseDef {
	if rItem == nil {
		log.Println("mapToResponseDef: unexpected nil input")
		return []ResponseDef{}
	}

	key := rItem.Key
	responseDef := ResponseDef{
		ID: key,
	}

	switch rItem.Role {
	case "singleChoiceGroup":
		for _, o := range rItem.Items {
			label, err := getTranslation(o.Content, lang)
			if err != nil {
				log.Printf("mapToResponseDef: label not found for: %v", o)
			}
			option := ResponseOption{
				ID:    o.Key,
				Label: label,
			}
			switch o.Role {
			case "option":
				option.OptionType = OPTION_TYPE_RADIO
			case "input":
				option.OptionType = OPTION_TYPE_TEXT_INPUT
			case "dateInput":
				option.OptionType = OPTION_TYPE_DATE_INPUT
			case "numberInput":
				option.OptionType = OPTION_TYPE_NUMBER_INPUT
			}
			responseDef.Options = append(responseDef.Options, option)
		}
		responseDef.ResponseType = QUESTION_TYPE_SINGLE_CHOICE
		return []ResponseDef{responseDef}
	case "multipleChoiceGroup":
		for _, o := range rItem.Items {
			label, err := getTranslation(o.Content, lang)
			if err != nil {
				log.Printf("mapToResponseDef: label not found for: %v", o)
			}
			option := ResponseOption{
				ID:    o.Key,
				Label: label,
			}
			switch o.Role {
			case "option":
				option.OptionType = OPTION_TYPE_CHECKBOX
			case "input":
				option.OptionType = OPTION_TYPE_TEXT_INPUT
			case "dateInput":
				option.OptionType = OPTION_TYPE_DATE_INPUT
			case "numberInput":
				option.OptionType = OPTION_TYPE_NUMBER_INPUT
			}
			responseDef.Options = append(responseDef.Options, option)
		}
		responseDef.ResponseType = QUESTION_TYPE_MULTIPLE_CHOICE
		return []ResponseDef{responseDef}
	case "dropDownGroup":
		for _, o := range rItem.Items {
			label, err := getTranslation(o.Content, lang)
			if err != nil {
				log.Printf("mapToResponseDef: label not found for: %v", o)
			}
			option := ResponseOption{
				ID:    o.Key,
				Label: label,
			}
			option.OptionType = OPTION_TYPE_DROPDOWN_OPTION
			responseDef.Options = append(responseDef.Options, option)
		}
		responseDef.ResponseType = QUESTION_TYPE_DROPDOWN
		return []ResponseDef{responseDef}
	case "input":
		label, err := getTranslation(rItem.Content, lang)
		if err != nil {
			log.Printf("mapToResponseDef: label not found for: %v", rItem)
		}
		responseDef.Label = label
		responseDef.ResponseType = QUESTION_TYPE_TEXT_INPUT
		return []ResponseDef{responseDef}
	case "multilineTextInput":
		label, err := getTranslation(rItem.Content, lang)
		if err != nil {
			log.Printf("mapToResponseDef: label not found for: %v", rItem)
		}
		responseDef.Label = label
		responseDef.ResponseType = QUESTION_TYPE_TEXT_INPUT
		return []ResponseDef{responseDef}
	case "numberInput":
		label, err := getTranslation(rItem.Content, lang)
		if err != nil {
			log.Printf("mapToResponseDef: label not found for: %v", rItem)
		}
		responseDef.Label = label
		responseDef.ResponseType = QUESTION_TYPE_NUMBER_INPUT
		return []ResponseDef{responseDef}
	case "dateInput":
		label, err := getTranslation(rItem.Content, lang)
		if err != nil {
			log.Printf("mapToResponseDef: label not found for: %v", rItem)
		}
		responseDef.Label = label
		responseDef.ResponseType = QUESTION_TYPE_DATE_INPUT
		return []ResponseDef{responseDef}
	case "eq5d-health-indicator":
		responseDef.Label = ""
		responseDef.ResponseType = QUESTION_TYPE_EQ5D_SLIDER
		return []ResponseDef{responseDef}
	case "sliderNumeric":
		label, err := getTranslation(rItem.Content, lang)
		if err != nil {
			log.Printf("mapToResponseDef: label not found for: %v", rItem)
		}
		responseDef.Label = label
		responseDef.ResponseType = QUESTION_TYPE_NUMERIC_SLIDER
		return []ResponseDef{responseDef}
	case "likert":
		for _, o := range rItem.Items {
			label, err := getTranslation(o.Content, lang)
			if err != nil {
				log.Printf("mapToResponseDef: label not found for: %v", o)
			}
			option := ResponseOption{
				ID:    o.Key,
				Label: label,
			}
			option.OptionType = OPTION_TYPE_RADIO
			responseDef.Options = append(responseDef.Options, option)
		}
		responseDef.ResponseType = QUESTION_TYPE_LIKERT
		return []ResponseDef{responseDef}
	case "likertGroup":
		responses := []ResponseDef{}
		for _, likertComp := range rItem.Items {
			if likertComp.Role != "likert" {
				continue
			}
			subKey := likertComp.Key
			currentResponseDef := ResponseDef{
				ID:           subKey,
				ResponseType: QUESTION_TYPE_LIKERT,
			}
			for _, o := range likertComp.Items {
				option := ResponseOption{
					ID: o.Key,
				}
				option.OptionType = OPTION_TYPE_RADIO
				currentResponseDef.Options = append(currentResponseDef.Options, option)
			}
			responses = append(responses, currentResponseDef)
		}
		return responses

	case "matrix":
		responses := []ResponseDef{}
		for _, row := range rItem.Items {
			rowKey := key + "." + row.Key
			if row.Role == "responseRow" {
				for _, col := range row.Items {
					cellKey := rowKey + "." + col.Key
					currentResponseDef := ResponseDef{
						ID: cellKey,
					}
					if col.Role == "dropDownGroup" {
						for _, o := range col.Items {
							dL, err := getTranslation(o.Content, lang)
							if err != nil {
								log.Printf("mapToResponseDef: label not found for: %v", o)
							}
							option := ResponseOption{
								ID:    o.Key,
								Label: dL,
							}
							option.OptionType = OPTION_TYPE_DROPDOWN_OPTION
							currentResponseDef.Options = append(currentResponseDef.Options, option)
						}
						currentResponseDef.ResponseType = QUESTION_TYPE_MATRIX_DROPDOWN
					} else if col.Role == "input" {
						label, err := getTranslation(col.Content, lang)
						if err != nil {
							log.Printf("mapToResponseDef: label not found for: %v", col)
						}
						currentResponseDef.ResponseType = QUESTION_TYPE_MATRIX_INPUT
						currentResponseDef.Label = label
					} else if col.Role == "check" {
						currentResponseDef.ResponseType = QUESTION_TYPE_MATRIX_CHECKBOX
					} else if col.Role == "numberInput" {
						label, err := getTranslation(col.Content, lang)
						if err != nil {
							log.Printf("mapToResponseDef: label not found for: %v", col)
						}
						currentResponseDef.ResponseType = QUESTION_TYPE_MATRIX_NUMBER_INPUT
						currentResponseDef.Label = label
					} else {
						log.Printf("mapToResponseDef: matrix cell role %s ignored.", col.Role)
						continue
					}
					responses = append(responses, currentResponseDef)
				}
			} else if row.Role == "radioRow" {
				currentResponseDef := ResponseDef{
					ID:           rowKey,
					ResponseType: QUESTION_TYPE_MATRIX_RADIO_ROW,
				}
				for _, o := range row.Items {
					if o.Role == "label" {
						label, err := getTranslation(o.Content, lang)
						if err != nil {
							log.Printf("mapToResponseDef: label not found for: %v", o)
						}
						currentResponseDef.Label = label
					} else {
						option := ResponseOption{
							ID: o.Key,
						}
						option.OptionType = OPTION_TYPE_RADIO
						currentResponseDef.Options = append(currentResponseDef.Options, option)
					}
				}
				responses = append(responses, currentResponseDef)
			}
		}
		return responses
	default:
		log.Printf("mapToResponseDef: component with role is ignored: %s [%s]", rItem.Role, key)
		return []ResponseDef{}
	}
}

func getQuestionType(responses []ResponseDef) string {
	var qType string
	if len(responses) < 1 {
		qType = QUESTION_TYPE_EMPTY
	} else if len(responses) == 1 {
		qType = responses[0].ResponseType
	} else {
		// mixed or map to something specific (e.g., if all the same...)
		qType = responses[0].ResponseType

		// Check for matrix questions:
		if strings.Contains(qType, QUESTION_TYPE_MATRIX) {
			return QUESTION_TYPE_MATRIX
		}

		// Check for other questions, that contain same subtype
		for _, r := range responses {
			if qType != r.ResponseType {
				return QUESTION_TYPE_UNKNOWN
			}
		}
	}

	return qType
}
