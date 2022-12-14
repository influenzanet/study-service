package exporter

import (
	"errors"
	"strings"

	"github.com/coneno/logger"
	"github.com/influenzanet/study-service/pkg/types"
	"github.com/influenzanet/study-service/pkg/utils"
)

func surveyDefToVersionPreview(original *types.Survey, prefLang string, includeItemNames []string, excludeItemNames []string) SurveyVersionPreview {
	sp := SurveyVersionPreview{
		VersionID:   original.VersionID,
		Published:   original.Published,
		Unpublished: original.Unpublished,
		Questions:   []SurveyQuestion{},
	}

	sp.Questions = extractQuestions(&original.SurveyDefinition, prefLang, includeItemNames, excludeItemNames)
	return sp
}

func extractQuestions(root *types.SurveyItem, prefLang string, includeItemNames []string, excludeItemNames []string) []SurveyQuestion {
	questions := []SurveyQuestion{}
	if root == nil {
		return questions
	}
	for _, item := range root.Items {
		if item.Type == "pageBreak" {
			continue
		}
		if item.ConfidentialMode != "" {
			continue
		}
		if isItemGroup(&item) {
			questions = append(questions, extractQuestions(&item, prefLang, includeItemNames, excludeItemNames)...)
			continue
		}

		if len(includeItemNames) > 0 {
			if !utils.ContainsString(includeItemNames, item.Key) {
				continue
			}
		} else if utils.ContainsString(excludeItemNames, item.Key) {
			continue
		}

		rg := getResponseGroupComponent(&item)
		if rg == nil {
			continue
		}

		responses, qType := extractResponses(rg, prefLang)

		titleComp := getTitleComponent(&item)
		title := ""
		if titleComp != nil {
			var err error
			title, err = getPreviewText(titleComp, prefLang)
			if err != nil {
				logger.Debug.Printf("Question %s title error: %v", item.Key, err)
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

func isItemGroup(item *types.SurveyItem) bool {
	return item != nil && len(item.Items) > 0
}

func getResponseGroupComponent(question *types.SurveyItem) *types.ItemComponent {
	if question.Components == nil {
		return nil
	}
	for _, c := range question.Components.Items {
		if c.Role == "responseGroup" {
			return &c
		}
	}
	return nil
}

func getTitleComponent(question *types.SurveyItem) *types.ItemComponent {
	if question.Components == nil {
		return nil
	}
	for _, c := range question.Components.Items {
		if c.Role == "title" {
			return &c
		}
	}
	return nil
}

func getPreviewText(item *types.ItemComponent, lang string) (string, error) {
	if lang == "ignored" {
		return "", nil
	}
	if item == nil {
		return "", errors.New("getPreviewText: item nil")
	}
	if len(item.Items) > 0 {
		translation := ""
		for _, item := range item.Items {
			part, _ := getTranslation(&item.Content, lang)
			translation += part
		}
		if translation == "" {
			return "", errors.New("translation missing")
		}
		return translation, nil
	} else {
		return getTranslation(&item.Content, lang)
	}
}

func getTranslation(content *[]types.LocalisedObject, lang string) (string, error) {
	if len(*content) < 1 {
		return "", errors.New("translations missing")
	}

	for _, translation := range *content {
		if translation.Code == lang {
			mergedText := ""
			for _, p := range translation.Parts {
				mergedText += p.ToAPI().GetStr()
			}
			return mergedText, nil
		}
	}
	return "", errors.New("translation missing")
}

func extractResponses(rg *types.ItemComponent, lang string) ([]ResponseDef, string) {
	if rg == nil {
		return []ResponseDef{}, QUESTION_TYPE_EMPTY
	}

	responses := []ResponseDef{}
	for _, item := range rg.Items {
		r := mapToResponseDef(&item, rg.Key, lang)
		responses = append(responses, r...)

	}

	qType := getQuestionType(responses)
	return responses, qType

}

func mapToResponseDef(rItem *types.ItemComponent, parentKey string, lang string) []ResponseDef {
	if rItem == nil {
		logger.Info.Println("mapToResponseDef: unexpected nil input")
		return []ResponseDef{}
	}

	key := rItem.Key
	responseDef := ResponseDef{
		ID: key,
	}

	var itemRole string
	roleSeparatorIndex := strings.Index(rItem.Role, ":")

	if roleSeparatorIndex == -1 {
		itemRole = rItem.Role
	} else if roleSeparatorIndex == 0 {
		responseDef.ResponseType = QUESTION_TYPE_UNKNOWN
		return []ResponseDef{responseDef}
	} else {
		itemRole = rItem.Role[0:roleSeparatorIndex]
	}

	switch itemRole {
	case "singleChoiceGroup":
		for _, o := range rItem.Items {
			label, err := getPreviewText(&o, lang)
			if err != nil {
				logger.Debug.Printf("mapToResponseDef: label not found for: %v", o)
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
			case "timeInput":
				option.OptionType = OPTION_TYPE_NUMBER_INPUT
			case "numberInput":
				option.OptionType = OPTION_TYPE_NUMBER_INPUT
			case "cloze":
				option.OptionType = OPTION_TYPE_CLOZE
			}
			responseDef.Options = append(responseDef.Options, option)
			if option.OptionType == OPTION_TYPE_CLOZE {
				clozeOptions := extractClozeInputOptions(o, option.ID, lang)
				responseDef.Options = append(responseDef.Options, clozeOptions...)
			}
		}
		responseDef.ResponseType = QUESTION_TYPE_SINGLE_CHOICE
		return []ResponseDef{responseDef}
	case "multipleChoiceGroup":
		for _, o := range rItem.Items {
			label, err := getPreviewText(&o, lang)
			if err != nil {
				logger.Debug.Printf("mapToResponseDef: label not found for: %v", o)
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
			case "timeInput":
				option.OptionType = OPTION_TYPE_NUMBER_INPUT
			case "numberInput":
				option.OptionType = OPTION_TYPE_NUMBER_INPUT
			case "cloze":
				option.OptionType = OPTION_TYPE_CLOZE
			}
			responseDef.Options = append(responseDef.Options, option)
			if option.OptionType == OPTION_TYPE_CLOZE {
				clozeOptions := extractClozeInputOptions(o, option.ID, lang)
				responseDef.Options = append(responseDef.Options, clozeOptions...)
			}
		}
		responseDef.ResponseType = QUESTION_TYPE_MULTIPLE_CHOICE
		return []ResponseDef{responseDef}
	case "dropDownGroup":
		for _, o := range rItem.Items {
			label, err := getPreviewText(&o, lang)
			if err != nil {
				logger.Debug.Printf("mapToResponseDef: label not found for: %v", o)
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
		label, err := getPreviewText(rItem, lang)
		if err != nil {
			logger.Debug.Printf("mapToResponseDef: label not found for: %v", rItem)
		}
		responseDef.Label = label
		responseDef.ResponseType = QUESTION_TYPE_TEXT_INPUT
		return []ResponseDef{responseDef}
	case "consent":
		label, err := getPreviewText(rItem, lang)
		if err != nil {
			logger.Debug.Printf("mapToResponseDef: label not found for: %v", rItem)
		}
		responseDef.Label = label
		responseDef.ResponseType = QUESTION_TYPE_CONSENT
		return []ResponseDef{responseDef}
	case "multilineTextInput":
		label, err := getPreviewText(rItem, lang)
		if err != nil {
			logger.Debug.Printf("mapToResponseDef: label not found for: %v", rItem)
		}
		responseDef.Label = label
		responseDef.ResponseType = QUESTION_TYPE_TEXT_INPUT
		return []ResponseDef{responseDef}
	case "numberInput":
		label, err := getPreviewText(rItem, lang)
		if err != nil {
			logger.Debug.Printf("mapToResponseDef: label not found for: %v", rItem)
		}
		responseDef.Label = label
		responseDef.ResponseType = QUESTION_TYPE_NUMBER_INPUT
		return []ResponseDef{responseDef}
	case "dateInput":
		label, err := getPreviewText(rItem, lang)
		if err != nil {
			logger.Debug.Printf("mapToResponseDef: label not found for: %v", rItem)
		}
		responseDef.Label = label
		responseDef.ResponseType = QUESTION_TYPE_DATE_INPUT
		return []ResponseDef{responseDef}
	case "timeInput":
		label, err := getPreviewText(rItem, lang)
		if err != nil {
			logger.Debug.Printf("mapToResponseDef: label not found for: %v", rItem)
		}
		responseDef.Label = label
		responseDef.ResponseType = QUESTION_TYPE_NUMBER_INPUT
		return []ResponseDef{responseDef}
	case "eq5d-health-indicator":
		responseDef.Label = ""
		responseDef.ResponseType = QUESTION_TYPE_EQ5D_SLIDER
		return []ResponseDef{responseDef}
	case "sliderNumeric":
		label, err := getPreviewText(rItem, lang)
		if err != nil {
			logger.Debug.Printf("mapToResponseDef: label not found for: %v", rItem)
		}
		responseDef.Label = label
		responseDef.ResponseType = QUESTION_TYPE_NUMERIC_SLIDER
		return []ResponseDef{responseDef}
	case "likert":
		for _, o := range rItem.Items {
			label, err := getPreviewText(&o, lang)
			if err != nil {
				logger.Debug.Printf("mapToResponseDef: label not found for: %v", o)
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
				ResponseType: QUESTION_TYPE_LIKERT_GROUP,
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
	case "responsiveSingleChoiceArray":
		responses := []ResponseDef{}

		var options *types.ItemComponent
		for _, item := range rItem.Items {
			if item.Role == "options" {
				options = &item
				break
			}
			continue
		}
		if options == nil {
			logger.Debug.Printf("mapToResponseDef: responsiveSingleChoiceArray - options not found in %v", rItem)
			return responses
		}

		for _, slot := range rItem.Items {
			if slot.Role != "row" {
				continue
			}
			subKey := slot.Key

			label, err := getPreviewText(&slot, lang)
			if err != nil {
				logger.Debug.Printf("mapToResponseDef: label not found for: %v", slot)
			}

			currentResponseDef := ResponseDef{
				ID:           subKey,
				ResponseType: QUESTION_TYPE_RESPONSIVE_SINGLE_CHOICE_ARRAY,
				Label:        label,
			}
			for _, o := range options.Items {
				label, err := getPreviewText(&o, lang)
				if err != nil {
					logger.Debug.Printf("mapToResponseDef: label not found for: %v", o)
				}

				option := ResponseOption{
					ID:    o.Key,
					Label: label,
				}
				option.OptionType = OPTION_TYPE_RADIO
				currentResponseDef.Options = append(currentResponseDef.Options, option)
			}
			responses = append(responses, currentResponseDef)
		}
		return responses
	case "responsiveBipolarLikertScaleArray":
		responses := []ResponseDef{}

		var options *types.ItemComponent
		for _, item := range rItem.Items {
			if item.Role == "options" {
				options = &item
				break
			}
			continue
		}
		if options == nil {
			logger.Debug.Printf("mapToResponseDef: responsiveBipolarLikertScaleArray - options not found in %v", rItem)
			return responses
		}

		for _, slot := range rItem.Items {
			if slot.Role != "row" {
				continue
			}
			subKey := slot.Key

			var start *types.ItemComponent
			var end *types.ItemComponent
			for _, item := range slot.Items {
				if start != nil && end != nil {
					break
				}
				if item.Role == "start" {
					start = &item
					continue
				} else if item.Role == "end" {
					end = &item
					continue
				}
			}

			startLabel, err := getPreviewText(start, lang)
			if err != nil {
				logger.Debug.Printf("mapToResponseDef: start label not found for: %v", slot)
			}
			endLabel, err := getPreviewText(end, lang)
			if err != nil {
				logger.Debug.Printf("mapToResponseDef: end label not found for: %v", slot)
			}

			currentResponseDef := ResponseDef{
				ID:           subKey,
				ResponseType: QUESTION_TYPE_RESPONSIVE_BIPOLAR_LIKERT_ARRAY,
				Label:        startLabel + " vs. " + endLabel,
			}
			for _, o := range options.Items {
				option := ResponseOption{
					ID:    o.Key,
					Label: o.Key,
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
							dL, err := getPreviewText(&o, lang)
							if err != nil {
								logger.Debug.Printf("mapToResponseDef: label not found for: %v", o)
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
						label, err := getPreviewText(&col, lang)
						if err != nil {
							logger.Debug.Printf("mapToResponseDef: label not found for: %v", col)
						}
						currentResponseDef.ResponseType = QUESTION_TYPE_MATRIX_INPUT
						currentResponseDef.Label = label
					} else if col.Role == "check" {
						currentResponseDef.ResponseType = QUESTION_TYPE_MATRIX_CHECKBOX
					} else if col.Role == "numberInput" {
						label, err := getPreviewText(&col, lang)
						if err != nil {
							logger.Debug.Printf("mapToResponseDef: label not found for: %v", col)
						}
						currentResponseDef.ResponseType = QUESTION_TYPE_MATRIX_NUMBER_INPUT
						currentResponseDef.Label = label
					} else {
						logger.Debug.Printf("mapToResponseDef: matrix cell role %s ignored.", col.Role)
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
						label, err := getPreviewText(&o, lang)
						if err != nil {
							logger.Debug.Printf("mapToResponseDef: label not found for: %v", o)
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
	case "responsiveMatrix":
		responses := []ResponseDef{}

		var columns *types.ItemComponent
		for _, item := range rItem.Items {
			if item.Role == "columns" {
				columns = &item
				break
			}
			continue
		}
		if columns == nil {
			logger.Debug.Printf("mapToResponseDef: responsiveMatrix - columns not found in %v", rItem)
			return responses
		}

		var rows *types.ItemComponent
		for _, item := range rItem.Items {
			if item.Role == "rows" {
				rows = &item
				break
			}
			continue
		}
		if rows == nil {
			logger.Debug.Printf("mapToResponseDef: responsiveMatrix - rows not found in %v", rItem)
			return responses
		}

		for _, row := range rows.Items {
			if row.Role == "category" {
				// ignore category rows
				continue
			}
			rowLabel, err := getPreviewText(&row, lang)
			if err != nil {
				logger.Debug.Printf("mapToResponseDef: row label not found for: %v, %v", row, err)
			}

			for _, col := range columns.Items {
				slotKey := row.Key + "-" + col.Key

				colLabel, err := getPreviewText(&col, lang)
				if err != nil {
					logger.Debug.Printf("mapToResponseDef: column label not found for: %v, %v", col, err)
				}
				currentResponseDef := ResponseDef{
					ID:           slotKey,
					ResponseType: QUESTION_TYPE_RESPONSIVE_TABLE,
					Label:        rowLabel + " || " + colLabel,
				}
				responses = append(responses, currentResponseDef)
			}
		}
		return responses
	case "cloze":
		for _, o := range rItem.Items {
			label, err := getPreviewText(&o, lang)
			if err != nil {
				logger.Debug.Printf("mapToResponseDef: label not found for: %v", o)
			}
			option := ResponseOption{
				ID:    o.Key,
				Label: label,
			}
			switch o.Role {
			case "input":
				option.OptionType = OPTION_TYPE_TEXT_INPUT
			case "dateInput":
				option.OptionType = OPTION_TYPE_DATE_INPUT
			case "timeInput":
				option.OptionType = OPTION_TYPE_NUMBER_INPUT
			case "numberInput":
				option.OptionType = OPTION_TYPE_NUMBER_INPUT
			case "dropDownGroup":
				option.OptionType = OPTION_TYPE_DROPDOWN
			}
			if option.OptionType != "" {
				responseDef.Options = append(responseDef.Options, option)
			}
		}
		responseDef.ResponseType = QUESTION_TYPE_CLOZE
		return []ResponseDef{responseDef}
	default:
		if roleSeparatorIndex > 0 {
			responseDef.ResponseType = QUESTION_TYPE_UNKNOWN
			return []ResponseDef{responseDef}
		}
		logger.Debug.Printf("mapToResponseDef: component with role is ignored: %s [%s]", rItem.Role, key)
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

func extractClozeInputOptions(option types.ItemComponent, clozeKey string, lang string) (clozeoptions []ResponseOption) {
	clozeInputs := []ResponseOption{}
	for _, o := range option.Items {
		label, err := getPreviewText(&o, lang)
		if err != nil {
			logger.Debug.Printf("mapToResponseDef: label not found for: %v", o)
		}
		option := ResponseOption{
			ID:    clozeKey + "." + o.Key,
			Label: label,
		}
		switch o.Role {
		case "input":
			option.OptionType = OPTION_TYPE_TEXT_INPUT
		case "dateInput":
			option.OptionType = OPTION_TYPE_DATE_INPUT
		case "timeInput":
			option.OptionType = OPTION_TYPE_NUMBER_INPUT
		case "numberInput":
			option.OptionType = OPTION_TYPE_NUMBER_INPUT
		case "dropDownGroup":
			option.OptionType = OPTION_TYPE_DROPDOWN
		}
		if option.OptionType != "" {
			clozeInputs = append(clozeInputs, option)
		}
	}
	return clozeInputs
}
