package exporter

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/coneno/logger"
	"github.com/influenzanet/study-service/pkg/types"
)

type ResponseExporter struct {
	surveyKey            string
	surveyVersions       []SurveyVersionPreview
	responses            []ParsedResponse
	contextColNames      []string
	contextColSeen       map[string]struct{}
	responseColNames     []string
	responseColSeen      map[string]struct{}
	metaColNames         []string
	metaColSeen          map[string]struct{}
	shortQuestionKeys    bool
	questionOptionKeySep string

	// metaColCache holds the four meta column name strings for every
	// question of every survey version keyed by VersionID
	metaColCache map[string]versionMetaCols

	// versionResponseCols holds the response column set for each
	// survey version keyed by VersionID
	versionResponseCols map[string]map[string]struct{}

	frozen             bool
	sortedContextCols  []string
	sortedResponseCols []string
	sortedMetaCols     []string
}

type versionMetaCols struct {
	init []string
	disp []string
	resp []string
	pos  []string
}

// Also update getFixedColumns when updating this
var fixedColumnKeys = []string{
	"ID",
	"participantID",
	"version",
	"opened",
	"submitted",
}

func (rp *ResponseExporter) getFixedColumns(resp ParsedResponse) map[string]interface{} {
	// Must always assign every entry of fixedColumnKeys
	return map[string]interface{}{
		fixedColumnKeys[0]: resp.ID,
		fixedColumnKeys[1]: resp.ParticipantID,
		fixedColumnKeys[2]: resp.Version,
		fixedColumnKeys[3]: resp.OpenedAt,
		fixedColumnKeys[4]: resp.SubmittedAt,
	}
}

func (rp *ResponseExporter) getFixedColumnValueStrings(resp ParsedResponse) []string {
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
		contextColSeen:       map[string]struct{}{},
		responseColSeen:      map[string]struct{}{},
		metaColSeen:          map[string]struct{}{},
		metaColCache:         map[string]versionMetaCols{},
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

	for _, sv := range rp.surveyVersions {
		cols := versionMetaCols{
			init: make([]string, len(sv.Questions)),
			disp: make([]string, len(sv.Questions)),
			resp: make([]string, len(sv.Questions)),
			pos:  make([]string, len(sv.Questions)),
		}
		for i, q := range sv.Questions {
			cols.init[i] = q.ID + rp.questionOptionKeySep + "metaInit"
			cols.disp[i] = q.ID + rp.questionOptionKeySep + "metaDisplayed"
			cols.resp[i] = q.ID + rp.questionOptionKeySep + "metaResponse"
			cols.pos[i] = q.ID + rp.questionOptionKeySep + "metaPosition"
		}
		rp.metaColCache[sv.VersionID] = cols
	}

	return &rp, nil
}

// InitSchemaColumns registers every response and meta column name derivable
// from the survey schema across all known versions. Must be called before the
// first response is processed on streaming export paths.
func (rp *ResponseExporter) InitSchemaColumns() {
	rp.InitSchemaColumnsForWindow(0, 0)
}

// InitSchemaColumnsForWindow is like InitSchemaColumns but restricts columns
// to survey versions that were active during [from, until].
// A version is included if it was published before until AND was not unpublished
// before from (unpublished==0 means still active). When from==0 and until==0
// all versions are included
func (rp *ResponseExporter) InitSchemaColumnsForWindow(from, until int64) {
	if rp.versionResponseCols == nil {
		rp.versionResponseCols = make(map[string]map[string]struct{})
	}
	for _, sv := range rp.surveyVersions {
		if from > 0 || until > 0 {
			// Exclude versions published after the window ends.
			if until > 0 && sv.Published > until {
				continue
			}
			// Exclude versions that were fully unpublished before the window starts.
			if from > 0 && sv.Unpublished > 0 && sv.Unpublished < from {
				continue
			}
		}
		vCols := make(map[string]struct{})
		for _, question := range sv.Questions {
			for k := range getResponseColumns(question, nil, rp.questionOptionKeySep) {
				rp.AddResponseColName(k)
				vCols[k] = struct{}{}
			}
		}
		rp.versionResponseCols[sv.VersionID] = vCols
		cols := rp.metaColCache[sv.VersionID]
		for i := range cols.init {
			for _, n := range []string{cols.init[i], cols.disp[i], cols.resp[i], cols.pos[i]} {
				if _, ok := rp.metaColSeen[n]; !ok {
					rp.metaColSeen[n] = struct{}{}
					rp.metaColNames = append(rp.metaColNames, n)
				}
			}
		}
	}
}

// AddResponse parses rawResp and retains the result in rp.responses.
// Used for callers that materialize the whole result set in memory (GetResponsesFlatJSONWithPagination).
func (rp *ResponseExporter) AddResponse(rawResp *types.SurveyResponse) error {
	parsed, err := rp.parseResponseInternal(rawResp)
	if err != nil {
		return err
	}
	rp.responses = append(rp.responses, *parsed)
	return nil
}

// ParseResponse parses rawResp and returns the result without retaining it.
// Used by the streaming export after Freeze().
func (rp *ResponseExporter) ParseResponse(rawResp *types.SurveyResponse) (*ParsedResponse, error) {
	return rp.parseResponseInternal(rawResp)
}

func (rp *ResponseExporter) parseResponseInternal(rawResp *types.SurveyResponse) (*ParsedResponse, error) {
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
			Position:    map[string]int32{},
		},
	}

	currentVersion, err := findSurveyVersion(rawResp.VersionID, rawResp.SubmittedAt, rp.surveyVersions)
	if err != nil {
		return nil, err
	}
	parsedResponse.resolvedVersionID = currentVersion.VersionID
	if currentVersion.VersionID != rawResp.VersionID && currentVersion.VersionID != "" {
		parsedResponse.Version = rawResp.VersionID + " (" + currentVersion.VersionID + ")"
		if rawResp.VersionID == "" {
			parsedResponse.Version = currentVersion.VersionID
			logger.Warning.Printf("VersionID of used survey is empty, only mapped versionID is displayed.")
		}
	}

	if rp.shortQuestionKeys {
		for i, r := range rawResp.Responses {
			rawResp.Responses[i].Key = strings.TrimPrefix(r.Key, rp.surveyKey+".")
		}
	}

	cache := rp.metaColCache[currentVersion.VersionID]
	for qIdx, question := range currentVersion.Questions {
		resp := findResponse(rawResp.Responses, question.ID)

		responseColumns := getResponseColumns(question, resp, rp.questionOptionKeySep)
		for k, v := range responseColumns {
			parsedResponse.Responses[k] = v
		}

		initColName := cache.init[qIdx]
		rp.AddMetaColName(initColName)
		parsedResponse.Meta.Initialised[initColName] = []int64{}

		dispColName := cache.disp[qIdx]
		rp.AddMetaColName(dispColName)
		parsedResponse.Meta.Displayed[dispColName] = []int64{}

		respColName := cache.resp[qIdx]
		rp.AddMetaColName(respColName)
		parsedResponse.Meta.Responded[respColName] = []int64{}

		positionColName := cache.pos[qIdx]
		rp.AddMetaColName(positionColName)
		parsedResponse.Meta.Position[positionColName] = 0

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
			parsedResponse.Meta.Position[positionColName] = resp.Meta.Position
		}
	}

	for k := range parsedResponse.Responses {
		rp.AddResponseColName(k)
	}
	for k := range parsedResponse.Context {
		rp.AddContextColName(k)
	}

	return &parsedResponse, nil
}

func (rp *ResponseExporter) AddResponseColName(name string) {
	if rp.frozen {
		return
	}
	if rp.responseColSeen == nil {
		rp.responseColSeen = map[string]struct{}{}
		for _, n := range rp.responseColNames {
			rp.responseColSeen[n] = struct{}{}
		}
	}
	if _, ok := rp.responseColSeen[name]; ok {
		return
	}
	rp.responseColSeen[name] = struct{}{}
	rp.responseColNames = append(rp.responseColNames, name)
}

func (rp *ResponseExporter) AddContextColName(name string) {
	if rp.frozen {
		return
	}
	if rp.contextColSeen == nil {
		rp.contextColSeen = map[string]struct{}{}
		for _, n := range rp.contextColNames {
			rp.contextColSeen[n] = struct{}{}
		}
	}
	if _, ok := rp.contextColSeen[name]; ok {
		return
	}
	rp.contextColSeen[name] = struct{}{}
	rp.contextColNames = append(rp.contextColNames, name)
}

func (rp *ResponseExporter) AddMetaColName(name string) {
	if rp.frozen {
		return
	}
	if rp.metaColSeen == nil {
		rp.metaColSeen = map[string]struct{}{}
		for _, n := range rp.metaColNames {
			rp.metaColSeen[n] = struct{}{}
		}
	}
	if _, ok := rp.metaColSeen[name]; ok {
		return
	}
	rp.metaColSeen[name] = struct{}{}
	rp.metaColNames = append(rp.metaColNames, name)
}

// Freeze locks the column set
func (rp *ResponseExporter) Freeze() {
	if rp.frozen {
		return
	}
	rp.sortedContextCols = append([]string(nil), rp.contextColNames...)
	sort.Strings(rp.sortedContextCols)
	rp.sortedResponseCols = append([]string(nil), rp.responseColNames...)
	sort.Strings(rp.sortedResponseCols)
	rp.sortedMetaCols = append([]string(nil), rp.metaColNames...)
	sort.Strings(rp.sortedMetaCols)
	rp.frozen = true
}

func (rp *ResponseExporter) GetSurveyVersionDefs() []SurveyVersionPreview {
	return rp.surveyVersions
}

func (rp *ResponseExporter) GetResponses() []ParsedResponse {
	return rp.responses
}

// RowSink writes serialized responses one at a time to its underlying writer.
// The order is: WriteHeader, WriteRow per ParsedResponse and then Flush.
// For the JSON sink, WriteHeader emits the opening bracket and Flush emits
// the closing bracket. For the CSV sinks WriteHeader writes the header row
// and Flush flushes the csv.Writer.
type RowSink interface {
	WriteHeader() error
	WriteRow(parsed *ParsedResponse) error
	Flush() error
}

// NewWideCSVSink returns a streaming wide-format CSV sink writing to w.
func (rp *ResponseExporter) NewWideCSVSink(w io.Writer, includeMeta *IncludeMeta) RowSink {
	return &wideCSVSink{rp: rp, csv: csv.NewWriter(w), includeMeta: includeMeta}
}

// NewLongCSVSink returns a streaming long-format CSV sink writing to w.
func (rp *ResponseExporter) NewLongCSVSink(w io.Writer, includeMeta *IncludeMeta) RowSink {
	return &longCSVSink{rp: rp, csv: csv.NewWriter(w), includeMeta: includeMeta}
}

// NewJSONSink returns a streaming JSON sink writing to w.      
func (rp *ResponseExporter) NewJSONSink(w io.Writer, includeMeta *IncludeMeta) RowSink {
	return &jsonSink{rp: rp, w: w, includeMeta: includeMeta}
}

type wideCSVSink struct {
	rp          *ResponseExporter
	csv         *csv.Writer
	includeMeta *IncludeMeta
}

func (s *wideCSVSink) WriteHeader() error {
	s.rp.Freeze()
	header := append([]string(nil), fixedColumnKeys...)
	header = append(header, s.rp.sortedContextCols...)
	header = append(header, s.rp.sortedResponseCols...)
	if s.includeMeta != nil {
		for _, c := range s.rp.sortedMetaCols {
			if !s.includeMeta.Postion && strings.Contains(c, "metaPosition") {
				continue
			}
			if !s.includeMeta.InitTimes && strings.Contains(c, "metaInit") {
				continue
			}
			if !s.includeMeta.DisplayedTimes && strings.Contains(c, "metaDisplayed") {
				continue
			}
			if !s.includeMeta.ResponsedTimes && strings.Contains(c, "metaResponse") {
				continue
			}
			header = append(header, c)
		}
	}
	return s.csv.Write(header)
}

func (s *wideCSVSink) WriteRow(parsed *ParsedResponse) error {
	line := s.rp.getFixedColumnValueStrings(*parsed)

	for _, colName := range s.rp.sortedContextCols {
		v, ok := parsed.Context[colName]
		if !ok {
			line = append(line, "")
			continue
		}
		line = append(line, v)
	}

	for _, colName := range s.rp.sortedResponseCols {
		v, ok := parsed.Responses[colName]
		if !ok {
			line = append(line, "")
			continue
		}
		line = append(line, responseColToString(v))
	}

	if s.includeMeta != nil {
		for _, colName := range s.rp.sortedMetaCols {
			if strings.Contains(colName, "metaInit") {
				if !s.includeMeta.InitTimes {
					continue
				}
				v, ok := parsed.Meta.Initialised[colName]
				if !ok {
					line = append(line, "")
					continue
				}
				line = append(line, timestampsToStr(v))
			} else if strings.Contains(colName, "metaDisplayed") {
				if !s.includeMeta.DisplayedTimes {
					continue
				}
				v, ok := parsed.Meta.Displayed[colName]
				if !ok {
					line = append(line, "")
					continue
				}
				line = append(line, timestampsToStr(v))
			} else if strings.Contains(colName, "metaResponse") {
				if !s.includeMeta.ResponsedTimes {
					continue
				}
				v, ok := parsed.Meta.Responded[colName]
				if !ok {
					line = append(line, "")
					continue
				}
				line = append(line, timestampsToStr(v))
			} else if strings.Contains(colName, "metaPosition") {
				if !s.includeMeta.Postion {
					continue
				}
				v, ok := parsed.Meta.Position[colName]
				if !ok {
					line = append(line, "")
					continue
				}
				line = append(line, fmt.Sprintf("%d", v))
			}
		}
	}

	return s.csv.Write(line)
}

func (s *wideCSVSink) Flush() error {
	s.csv.Flush()
	return s.csv.Error()
}

type longCSVSink struct {
	rp          *ResponseExporter
	csv         *csv.Writer
	includeMeta *IncludeMeta
}

func (s *longCSVSink) WriteHeader() error {
	s.rp.Freeze()
	header := append([]string(nil), fixedColumnKeys...)
	header = append(header, s.rp.sortedContextCols...)
	header = append(header, "responseSlot", "value")
	return s.csv.Write(header)
}

func (s *longCSVSink) WriteRow(parsed *ParsedResponse) error {
	line := s.rp.getFixedColumnValueStrings(*parsed)

	for _, colName := range s.rp.sortedContextCols {
		v, ok := parsed.Context[colName]
		if !ok {
			line = append(line, "")
			continue
		}
		line = append(line, v)
	}

	for _, colName := range s.rp.sortedResponseCols {
		v, ok := parsed.Responses[colName]
		if !ok {
			// Column has no data for this response. Skip if it belongs to a
			// different survey version; emit an empty row if it belongs to the
			// same version (matching stable behaviour for questions with no
			// selection, e.g. a multiple-choice group where nothing was picked).
			if vCols := s.rp.versionResponseCols[parsed.resolvedVersionID]; vCols != nil {
				if _, inVersion := vCols[colName]; !inVersion {
					continue
				}
			}
		}
		value := ""
		if ok {
			value = responseColToString(v)
		}
		currentRespLine := append(append([]string(nil), line...), colName, value)
		if err := s.csv.Write(currentRespLine); err != nil {
			return err
		}
	}

	if s.includeMeta != nil {
		for _, colName := range s.rp.sortedMetaCols {
			value := ""
			found := false
			if strings.Contains(colName, "metaInit") {
				if !s.includeMeta.InitTimes {
					continue
				}
				v, ok := parsed.Meta.Initialised[colName]
				if ok {
					value = timestampsToStr(v)
					found = true
				}
			} else if strings.Contains(colName, "metaDisplayed") {
				if !s.includeMeta.DisplayedTimes {
					continue
				}
				v, ok := parsed.Meta.Displayed[colName]
				if ok {
					value = timestampsToStr(v)
					found = true
				}
			} else if strings.Contains(colName, "metaResponse") {
				if !s.includeMeta.ResponsedTimes {
					continue
				}
				v, ok := parsed.Meta.Responded[colName]
				if ok {
					value = timestampsToStr(v)
					found = true
				}
			} else if strings.Contains(colName, "metaPosition") {
				if !s.includeMeta.Postion {
					continue
				}
				v, ok := parsed.Meta.Position[colName]
				if ok {
					value = fmt.Sprintf("%d", v)
					found = true
				}
			}
			if !found {
				continue
			}

			currentRespLine := []string{}
			currentRespLine = append(currentRespLine, line...)
			currentRespLine = append(currentRespLine, colName)
			currentRespLine = append(currentRespLine, value)
			if err := s.csv.Write(currentRespLine); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *longCSVSink) Flush() error {
	s.csv.Flush()
	return s.csv.Error()
}

type jsonSink struct {
	rp          *ResponseExporter
	w           io.Writer
	includeMeta *IncludeMeta
	opened      bool
	closed      bool
	rowEmitted  bool
}

func (s *jsonSink) WriteHeader() error {
	if s.opened {
		return nil
	}
	s.rp.Freeze()
	if _, err := io.WriteString(s.w, "["); err != nil {
		return err
	}
	s.opened = true
	return nil
}

func (s *jsonSink) WriteRow(parsed *ParsedResponse) error {
	if !s.opened {
		if err := s.WriteHeader(); err != nil {
			return err
		}
	}
	if s.rowEmitted {
		if _, err := io.WriteString(s.w, ","); err != nil {
			return err
		}
	}

	currentResp := s.rp.getFixedColumns(*parsed)

	for _, colName := range s.rp.sortedContextCols {
		v, ok := parsed.Context[colName]
		if !ok {
			currentResp[colName] = ""
		} else {
			currentResp[colName] = v
		}
	}

	for _, colName := range s.rp.sortedResponseCols {
		r, ok := parsed.Responses[colName]
		if !ok {
			currentResp[colName] = ""
		} else {
			currentResp[colName] = r
		}
	}

	if s.includeMeta != nil {
		for _, colName := range s.rp.sortedMetaCols {
			if strings.Contains(colName, "metaInit") {
				if !s.includeMeta.InitTimes {
					continue
				}
				v, ok := parsed.Meta.Initialised[colName]
				if !ok {
					currentResp[colName] = ""
				} else {
					currentResp[colName] = v
				}
			} else if strings.Contains(colName, "metaDisplayed") {
				if !s.includeMeta.DisplayedTimes {
					continue
				}
				v, ok := parsed.Meta.Displayed[colName]
				if !ok {
					currentResp[colName] = ""
				} else {
					currentResp[colName] = v
				}
			} else if strings.Contains(colName, "metaResponse") {
				if !s.includeMeta.ResponsedTimes {
					continue
				}
				v, ok := parsed.Meta.Responded[colName]
				if !ok {
					currentResp[colName] = ""
				} else {
					currentResp[colName] = v
				}
			} else if strings.Contains(colName, "metaPosition") {
				if !s.includeMeta.Postion {
					continue
				}
				v, ok := parsed.Meta.Position[colName]
				if !ok {
					currentResp[colName] = ""
				} else {
					currentResp[colName] = v
				}
			}
		}
	}

	b, err := json.Marshal(currentResp)
	if err != nil {
		return err
	}
	if _, err := s.w.Write(b); err != nil {
		return err
	}
	s.rowEmitted = true
	return nil
}

func (s *jsonSink) Flush() error {
	if s.closed {
		return nil
	}
	s.closed = true
	if !s.opened {
		_, err := io.WriteString(s.w, "[]")
		return err
	}
	_, err := io.WriteString(s.w, "]")
	return err
}

// Legacy buffered serializers.
// Retained for callers that load the whole result set in memory (e.g. GetResponsesFlatJSONWithPagination
// and those behind the streaming flag in data_export.go)
func (rp *ResponseExporter) GetResponsesJSON(writer io.Writer, includeMeta *IncludeMeta) error {
	sink := rp.NewJSONSink(writer, includeMeta)
	if err := sink.WriteHeader(); err != nil {
		return err
	}
	for i := range rp.responses {
		if err := sink.WriteRow(&rp.responses[i]); err != nil {
			return err
		}
	}
	return sink.Flush()
}

func (rp *ResponseExporter) GetResponsesCSV(writer io.Writer, includeMeta *IncludeMeta) error {
	if len(rp.responses) < 1 {
		return errors.New("no responses, nothing is generated")
	}
	sink := rp.NewWideCSVSink(writer, includeMeta)
	if err := sink.WriteHeader(); err != nil {
		return err
	}
	for i := range rp.responses {
		if err := sink.WriteRow(&rp.responses[i]); err != nil {
			return err
		}
	}
	return sink.Flush()
}

func (rp *ResponseExporter) GetResponsesLongFormatCSV(writer io.Writer, metaInfos *IncludeMeta) error {
	if len(rp.responses) < 1 {
		return errors.New("no responses, nothing is generated")
	}
	sink := rp.NewLongCSVSink(writer, metaInfos)
	if err := sink.WriteHeader(); err != nil {
		return err
	}
	for i := range rp.responses {
		if err := sink.WriteRow(&rp.responses[i]); err != nil {
			return err
		}
	}
	return sink.Flush()
}

func (rp *ResponseExporter) GetSurveyInfoCSV(writer io.Writer) error {
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
