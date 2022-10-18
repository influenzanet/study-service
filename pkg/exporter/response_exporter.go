package exporter

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/influenzanet/study-service/pkg/types"
)

type ResponseExporter struct {
	surveyKey            string
	surveyVersions       []SurveyVersionPreview
	responses            []ParsedResponse
	contextColNames      []string
	responseColNames     []string
	metaColNames         []string
	shortQuestionKeys    bool
	questionOptionKeySep string
}

// Also update getFixedColumns when updating this
var fixedColumnKeys = []string{
	"ID",
	"participantID",
	"version",
	"opened",
	"submitted",
}

func (rp ResponseExporter) getFixedColumns(resp ParsedResponse) map[string]interface{} {
	// Must always assign every entry of fixedColumnKeys
	return map[string]interface{}{
		fixedColumnKeys[0]: resp.ID,
		fixedColumnKeys[1]: resp.ParticipantID,
		fixedColumnKeys[2]: resp.Version,
		fixedColumnKeys[3]: resp.OpenedAt,
		fixedColumnKeys[4]: resp.SubmittedAt,
	}
}

func (rp ResponseExporter) getFixedColumnValueStrings(resp ParsedResponse) []string {
	fixedColumns := rp.getFixedColumns(resp)
	valueStrings := make([]string, 0, len(fixedColumnKeys))

	for _, k := range fixedColumnKeys {
		var stringValue string
		c, ok := fixedColumns[k]
		if !ok {
			stringValue = ""
		} else {
			switch value := c.(type) {
			case string:
				stringValue = value
			default:
				stringValue = fmt.Sprint(value)
			}
		}

		valueStrings = append(valueStrings, stringValue)
	}

	return valueStrings
}

func NewResponseExporter(
	surveyHistory []*types.Survey,
	previewLang string,
	shortQuestionKeys bool,
	questionOptionSep string,
) (*ResponseExporter, error) {
	return newResponseExporterBase(surveyHistory, previewLang, shortQuestionKeys, questionOptionSep, []string{}, []string{})
}

func NewResponseExporterWithIncludeFilter(
	surveyHistory []*types.Survey,
	previewLang string,
	shortQuestionKeys bool,
	questionOptionSep string,
	includeItemNames []string,
) (*ResponseExporter, error) {
	return newResponseExporterBase(surveyHistory, previewLang, shortQuestionKeys, questionOptionSep, includeItemNames, []string{})
}

func NewResponseExporterWithExcludeFilter(
	surveyHistory []*types.Survey,
	previewLang string,
	shortQuestionKeys bool,
	questionOptionSep string,
	excludeItemNames []string,
) (*ResponseExporter, error) {
	return newResponseExporterBase(surveyHistory, previewLang, shortQuestionKeys, questionOptionSep, []string{}, excludeItemNames)
}

func newResponseExporterBase(
	surveyHistory []*types.Survey,
	previewLang string,
	shortQuestionKeys bool,
	questionOptionSep string,
	includeItemNames []string,
	excludeItemNames []string,
) (*ResponseExporter, error) {
	if len(surveyHistory) < 1 {
		return nil, errors.New("survey definition history not found")
	}

	rp := ResponseExporter{
		surveyKey:            surveyHistory[0].SurveyDefinition.Key,
		surveyVersions:       []SurveyVersionPreview{},
		responses:            []ParsedResponse{},
		shortQuestionKeys:    shortQuestionKeys,
		questionOptionKeySep: questionOptionSep,
	}

	for _, v := range surveyHistory {
		rp.surveyVersions = append(rp.surveyVersions, surveyDefToVersionPreview(v, previewLang, includeItemNames, excludeItemNames))
	}

	if shortQuestionKeys {
		for versionInd, sv := range rp.surveyVersions {
			for qInd, question := range sv.Questions {
				rp.surveyVersions[versionInd].Questions[qInd].ID = strings.TrimPrefix(question.ID, rp.surveyKey+".")
			}
		}
	}

	return &rp, nil
}

func (rp *ResponseExporter) AddResponse(rawResp *types.SurveyResponse) error {
	parsedResponse := ParsedResponse{
		ID:            rawResp.ID.Hex(),
		ParticipantID: rawResp.ParticipantID,
		Version:       rawResp.VersionID,
		OpenedAt:      rawResp.OpenedAt,
		SubmittedAt:   rawResp.SubmittedAt,
		Context:       rawResp.Context,
		Responses:     map[string]interface{}{},
		Meta: ResponseMeta{
			Initialised: map[string][]int64{},
			Displayed:   map[string][]int64{},
			Responded:   map[string][]int64{},
			ItemVersion: map[string]string{},
			Position:    map[string]int32{},
		},
	}

	currentVersion, err := findSurveyVersion(rawResp.VersionID, rawResp.SubmittedAt, rp.surveyVersions)
	if err != nil {
		return err
	}

	if rp.shortQuestionKeys {
		for i, r := range rawResp.Responses {
			rawResp.Responses[i].Key = strings.TrimPrefix(r.Key, rp.surveyKey+".")
		}
	}

	for _, question := range currentVersion.Questions {
		resp := findResponse(rawResp.Responses, question.ID)

		responseColumns := getResponseColumns(question, resp, rp.questionOptionKeySep)
		for k, v := range responseColumns {
			parsedResponse.Responses[k] = v
		}

		// Set meta infos
		initColName := question.ID + rp.questionOptionKeySep + "metaInit"
		rp.AddMetaColName(initColName)
		parsedResponse.Meta.Initialised[initColName] = []int64{}

		dispColName := question.ID + rp.questionOptionKeySep + "metaDisplayed"
		rp.AddMetaColName(dispColName)
		parsedResponse.Meta.Displayed[dispColName] = []int64{}

		respColName := question.ID + rp.questionOptionKeySep + "metaResponse"
		rp.AddMetaColName(respColName)
		parsedResponse.Meta.Responded[respColName] = []int64{}

		itemVColName := question.ID + rp.questionOptionKeySep + "metaItemVersion"
		rp.AddMetaColName(itemVColName)
		parsedResponse.Meta.ItemVersion[itemVColName] = ""

		positionColName := question.ID + rp.questionOptionKeySep + "metaPosition"
		rp.AddMetaColName(positionColName)
		parsedResponse.Meta.ItemVersion[positionColName] = ""

		if resp != nil {
			if resp.Meta.Rendered != nil {
				parsedResponse.Meta.Initialised[initColName] = resp.Meta.Rendered
			}
			if resp.Meta.Displayed != nil {
				parsedResponse.Meta.Displayed[dispColName] = resp.Meta.Displayed
			}
			if resp.Meta.Responded != nil {
				parsedResponse.Meta.Responded[respColName] = resp.Meta.Responded
			}
			parsedResponse.Meta.ItemVersion[itemVColName] = strconv.Itoa(int(resp.Meta.Version))
			parsedResponse.Meta.Position[positionColName] = resp.Meta.Position
		}
	}

	// Extend response col names:
	for k := range parsedResponse.Responses {
		rp.AddResponseColName(k)
	}
	for k := range parsedResponse.Context {
		rp.AddContextColName(k)
	}

	rp.responses = append(rp.responses, parsedResponse)
	return nil
}

func (rp *ResponseExporter) AddResponseColName(name string) {
	for _, n := range rp.responseColNames {
		if n == name {
			return
		}
	}
	rp.responseColNames = append(rp.responseColNames, name)
}

func (rp *ResponseExporter) AddContextColName(name string) {
	for _, n := range rp.contextColNames {
		if n == name {
			return
		}
	}
	rp.contextColNames = append(rp.contextColNames, name)
}

func (rp *ResponseExporter) AddMetaColName(name string) {
	for _, n := range rp.metaColNames {
		if n == name {
			return
		}
	}
	rp.metaColNames = append(rp.metaColNames, name)
}

func (rp ResponseExporter) GetSurveyVersionDefs() []SurveyVersionPreview {
	return rp.surveyVersions
}

func (rp ResponseExporter) GetResponses() []ParsedResponse {
	return rp.responses
}

func (rp ResponseExporter) GetResponsesJSON(writer io.Writer, includeMeta *IncludeMeta) error {
	responseArray := []map[string]interface{}{}
	for _, resp := range rp.responses {

		currentResp := rp.getFixedColumns(resp)

		responseCols := rp.responseColNames
		for _, colName := range responseCols {
			r, ok := resp.Responses[colName]
			if !ok {
				currentResp[colName] = ""
			} else {
				currentResp[colName] = r
			}
		}

		metaCols := rp.metaColNames
		sort.Strings(metaCols)

		if includeMeta != nil {
			for _, colName := range metaCols {
				if strings.Contains(colName, "metaInit") {
					if !includeMeta.InitTimes {
						continue
					}
					v, ok := resp.Meta.Initialised[colName]
					if !ok {
						currentResp[colName] = ""
					} else {
						currentResp[colName] = v
					}
				} else if strings.Contains(colName, "metaDisplayed") {
					if !includeMeta.DisplayedTimes {
						continue
					}
					v, ok := resp.Meta.Displayed[colName]
					if !ok {
						currentResp[colName] = ""
					} else {
						currentResp[colName] = v
					}
				} else if strings.Contains(colName, "metaResponse") {
					if !includeMeta.ResponsedTimes {
						continue
					}
					v, ok := resp.Meta.Responded[colName]
					if !ok {
						currentResp[colName] = ""
					} else {
						currentResp[colName] = v
					}
				} else if strings.Contains(colName, "metaItemVersion") {
					if !includeMeta.ItemVersion {
						continue
					}
					v, ok := resp.Meta.ItemVersion[colName]
					if !ok {
						currentResp[colName] = ""
					} else {
						currentResp[colName] = v
					}
				} else if strings.Contains(colName, "metaPosition") {
					if !includeMeta.Postion {
						continue
					}
					v, ok := resp.Meta.Position[colName]
					if !ok {
						currentResp[colName] = ""
					} else {
						currentResp[colName] = v
					}
				}
			}
		}

		responseArray = append(responseArray, currentResp)
	}
	b, err := json.Marshal(responseArray)
	if err != nil {
		return err
	}
	_, err = writer.Write(b)
	return err
}

func (rp ResponseExporter) GetResponsesCSV(writer io.Writer, includeMeta *IncludeMeta) error {
	if len(rp.responses) < 1 {
		return errors.New("no responses, nothing is generated")
	}

	// Sort column names
	contextCols := rp.contextColNames
	sort.Strings(contextCols)
	responseCols := rp.responseColNames
	sort.Strings(responseCols)
	metaCols := rp.metaColNames
	sort.Strings(metaCols)

	// Prepare csv header
	header := fixedColumnKeys
	header = append(header, contextCols...)
	header = append(header, responseCols...)
	if includeMeta != nil {
		for _, c := range metaCols {
			if !includeMeta.Postion && strings.Contains(c, "metaPosition") {
				continue
			}
			if !includeMeta.InitTimes && strings.Contains(c, "metaInit") {
				continue
			}
			if !includeMeta.DisplayedTimes && strings.Contains(c, "metaDisplayed") {
				continue
			}
			if !includeMeta.ResponsedTimes && strings.Contains(c, "metaResponse") {
				continue
			}
			if !includeMeta.ItemVersion && strings.Contains(c, "metaItemVersion") {
				continue
			}
			header = append(header, c)
		}
	}

	// Init writer
	w := csv.NewWriter(writer)

	// Write header
	err := w.Write(header)
	if err != nil {
		return err
	}

	// Write responses
	for _, resp := range rp.responses {
		line := rp.getFixedColumnValueStrings(resp)

		for _, colName := range contextCols {
			v, ok := resp.Context[colName]
			if !ok {
				line = append(line, "")
				continue
			}
			line = append(line, v)
		}

		for _, colName := range responseCols {
			v, ok := resp.Responses[colName]
			if !ok {
				line = append(line, "")
				continue
			}
			line = append(line, responseColToString(v))
		}

		if includeMeta != nil {
			for _, colName := range metaCols {
				if strings.Contains(colName, "metaInit") {
					if !includeMeta.InitTimes {
						continue
					}
					v, ok := resp.Meta.Initialised[colName]
					if !ok {
						line = append(line, "")
						continue
					}
					line = append(line, timestampsToStr(v))
				} else if strings.Contains(colName, "metaDisplayed") {
					if !includeMeta.DisplayedTimes {
						continue
					}
					v, ok := resp.Meta.Displayed[colName]
					if !ok {
						line = append(line, "")
						continue
					}
					line = append(line, timestampsToStr(v))
				} else if strings.Contains(colName, "metaResponse") {
					if !includeMeta.ResponsedTimes {
						continue
					}
					v, ok := resp.Meta.Responded[colName]
					if !ok {
						line = append(line, "")
						continue
					}
					line = append(line, timestampsToStr(v))
				} else if strings.Contains(colName, "metaItemVersion") {
					if !includeMeta.ItemVersion {
						continue
					}
					v, ok := resp.Meta.ItemVersion[colName]
					if !ok {
						line = append(line, "")
						continue
					}
					line = append(line, v)
				} else if strings.Contains(colName, "metaPosition") {
					if !includeMeta.Postion {
						continue
					}
					v, ok := resp.Meta.Position[colName]
					if !ok {
						line = append(line, "")
						continue
					}
					line = append(line, fmt.Sprintf("%d", v))
				}
			}
		}

		err := w.Write(line)
		if err != nil {
			return err
		}
	}
	w.Flush()
	return nil
}

func (rp ResponseExporter) GetResponsesLongFormatCSV(writer io.Writer, metaInfos *IncludeMeta) error {
	if len(rp.responses) < 1 {
		return errors.New("no responses, nothing is generated")
	}

	// Sort column names
	contextCols := rp.contextColNames
	sort.Strings(contextCols)
	responseCols := rp.responseColNames
	sort.Strings(responseCols)
	metaCols := rp.metaColNames
	sort.Strings(metaCols)

	// Prepare csv header
	header := fixedColumnKeys
	header = append(header, contextCols...)
	header = append(header, "responseSlot")
	header = append(header, "value")

	// Init writer
	w := csv.NewWriter(writer)

	// Write header
	err := w.Write(header)
	if err != nil {
		return err
	}

	// Write responses
	for _, resp := range rp.responses {
		line := rp.getFixedColumnValueStrings(resp)

		for _, colName := range contextCols {
			v, ok := resp.Context[colName]
			if !ok {
				line = append(line, "")
				continue
			}
			line = append(line, v)
		}

		for _, colName := range responseCols {
			currentRespLine := []string{}
			currentRespLine = append(currentRespLine, line...)
			currentRespLine = append(currentRespLine, colName)
			v, ok := resp.Responses[colName]
			if !ok {
				currentRespLine = append(currentRespLine, "")
			} else {
				currentRespLine = append(currentRespLine, responseColToString(v))
			}

			err := w.Write(currentRespLine)
			if err != nil {
				return err
			}
		}

		if metaInfos != nil {
			for _, colName := range metaCols {
				value := ""
				if strings.Contains(colName, "metaInit") {
					if !metaInfos.InitTimes {
						continue
					}
					v, ok := resp.Meta.Initialised[colName]
					if ok {
						value = timestampsToStr(v)
					}
				} else if strings.Contains(colName, "metaDisplayed") {
					if !metaInfos.DisplayedTimes {
						continue
					}
					v, ok := resp.Meta.Displayed[colName]
					if ok {
						value = timestampsToStr(v)
					}
				} else if strings.Contains(colName, "metaResponse") {
					if !metaInfos.ResponsedTimes {
						continue
					}
					v, ok := resp.Meta.Responded[colName]
					if ok {
						value = timestampsToStr(v)
					}
				} else if strings.Contains(colName, "metaItemVersion") {
					if !metaInfos.ItemVersion {
						continue
					}
					v, ok := resp.Meta.ItemVersion[colName]
					if ok {
						value = v
					}
				} else if strings.Contains(colName, "metaPosition") {
					if !metaInfos.Postion {
						continue
					}
					v, ok := resp.Meta.Position[colName]
					if ok {
						value = fmt.Sprintf("%d", v)
					}
				}

				currentRespLine := []string{}
				currentRespLine = append(currentRespLine, line...)
				currentRespLine = append(currentRespLine, colName)
				currentRespLine = append(currentRespLine, value)

				err := w.Write(currentRespLine)
				if err != nil {
					return err
				}
			}
		}

	}
	w.Flush()
	return nil
}

func (rp ResponseExporter) GetSurveyInfoCSV(writer io.Writer) error {
	header := []string{
		"surveyKey", "versionID", "questionKey", "title",
		"responseKey", "type", "optionKey", "optionType", "optionLabel",
	}

	// Init writer
	w := csv.NewWriter(writer)

	// Write header
	err := w.Write(header)
	if err != nil {
		return err
	}

	for i, currentVersion := range rp.surveyVersions {
		version := currentVersion.VersionID
		if version == "" {
			version = fmt.Sprintf("%d", i)
		}

		for _, question := range currentVersion.Questions {
			questionCols := []string{
				rp.surveyKey,
				version,
				question.ID,
				question.Title,
			}
			for _, slot := range question.Responses {
				slotCols := []string{
					slot.ID,
					slot.ResponseType,
				}

				if len(slot.Options) > 0 {
					for _, option := range slot.Options {
						line := []string{}
						line = append(line, questionCols...)
						line = append(line, slotCols...)
						line = append(line, []string{
							option.ID,
							option.OptionType,
							option.Label,
						}...)

						err := w.Write(line)
						if err != nil {
							return err
						}
					}
				} else {
					line := []string{}
					line = append(line, questionCols...)
					line = append(line, slotCols...)
					line = append(line, []string{
						"",
						"",
						"",
					}...)
					err := w.Write(line)
					if err != nil {
						return err
					}
				}

			}
		}
	}

	w.Flush()
	return nil
}
