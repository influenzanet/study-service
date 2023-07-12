package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/coneno/logger"
	"github.com/influenzanet/go-utils/pkg/api_types"
	"github.com/influenzanet/go-utils/pkg/constants"
	"github.com/influenzanet/go-utils/pkg/token_checks"
	"github.com/influenzanet/study-service/pkg/api"
	"github.com/influenzanet/study-service/pkg/dbs/studydb"
	"github.com/influenzanet/study-service/pkg/exporter"
	"github.com/influenzanet/study-service/pkg/studyengine"
	"github.com/influenzanet/study-service/pkg/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	loggingAPI "github.com/influenzanet/logging-service/pkg/api"
)

const CHUNK_SIZE = 64 * 1024 // 64 KiB

type ResponseFormat int

const (
	FLAT_JSON       ResponseFormat = iota
	WIDE_FORMAT_CSV ResponseFormat = iota
	LONG_FORMAT_CSV ResponseFormat = iota
)

func (s *studyServiceServer) GetStudyResponseStatistics(ctx context.Context, req *api.SurveyResponseQuery) (*api.StudyResponseStatistics, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return nil, s.missingArgumentError()
	}

	if !(token_checks.CheckRoleInToken(req.Token, constants.USER_ROLE_ADMIN) &&
		token_checks.CheckRoleInToken(req.Token, constants.USER_ROLE_RESEARCHER)) {
		err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id, []string{
			types.STUDY_ROLE_OWNER,
			types.STUDY_ROLE_MAINTAINER,
			"analyst"})
		if err != nil {
			s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_DOWNLOAD_RESPONSES, "Statistics: permission denied for "+req.StudyKey)
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	keys, err := s.studyDBservice.GetSurveyResponseKeys(req.Token.InstanceId, req.StudyKey, req.From, req.Until)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &api.StudyResponseStatistics{
		SurveyResponseCounts: map[string]int64{},
	}

	for _, k := range keys {
		count, err := s.studyDBservice.CountSurveyResponsesByKey(req.Token.InstanceId, req.StudyKey, k, req.From, req.Until)
		if err != nil {
			logger.Error.Printf("unexpected error: %v", err)
			continue
		}
		resp.SurveyResponseCounts[k] = count
	}
	s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_LOG, constants.LOG_EVENT_DOWNLOAD_RESPONSES, "statistics: "+req.StudyKey)
	return resp, nil
}

func (s *studyServiceServer) GetConfidentialResponses(ctx context.Context, req *api.ConfidentialResponsesQuery) (*api.ConfidentialResponses, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return nil, s.missingArgumentError()
	}

	if !(token_checks.CheckRoleInToken(req.Token, constants.USER_ROLE_ADMIN) &&
		token_checks.CheckRoleInToken(req.Token, constants.USER_ROLE_RESEARCHER)) {
		err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id, []string{
			types.STUDY_ROLE_OWNER,
			types.STUDY_ROLE_MAINTAINER,
		})
		if err != nil {
			s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_DOWNLOAD_RESPONSES, "Confidential data: permission denied for "+req.StudyKey)
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	resp := &api.ConfidentialResponses{
		Responses: []*api.SurveyResponse{},
	}
	for _, participantID := range req.ParticipantIds {
		confPID, _, err := s.profileIDToParticipantID(req.Token.InstanceId, req.StudyKey, participantID, true)
		if err != nil {
			logger.Error.Printf("unexpected error: %v", err)
			continue
		}

		pState, err := s.studyDBservice.FindParticipantState(req.Token.InstanceId, req.StudyKey, participantID)
		if err != nil {
			logger.Error.Printf("participant state for %s not found: %v", participantID, err)
			continue
		}

		val, err := studyengine.ExpressionEval(*types.ExpressionFromAPI(req.Condition), studyengine.EvalContext{
			Event: types.StudyEvent{
				InstanceID:                            req.Token.InstanceId,
				StudyKey:                              req.StudyKey,
				ParticipantIDForConfidentialResponses: confPID,
			},
			ParticipantState: pState,
			Configs: studyengine.ActionConfigs{
				DBService:              s.studyDBservice,
				ExternalServiceConfigs: s.studyEngineExternalServices,
			},
		})
		conditionTrue, ok := val.(bool)
		if err != nil || !ok || !conditionTrue {
			logger.Error.Printf("participant '%s' has not fulfilled condition. (%v)", participantID, err)
			continue
		}

		pResps, err := s.studyDBservice.FindConfidentialResponses(req.Token.InstanceId, req.StudyKey, confPID, req.KeyFilter)
		if err != nil {
			logger.Error.Printf("unexpected error: %v", err)
			continue
		}

		for _, r := range pResps {
			r.ParticipantID = participantID // override participant ID so that data can be used
			resp.Responses = append(resp.Responses, r.ToAPI())
		}
	}

	s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_LOG, constants.LOG_EVENT_DOWNLOAD_RESPONSES, "Confidential data: "+req.StudyKey)
	return resp, nil
}

func (s *studyServiceServer) StreamParticipantStates(req *api.ParticipantStateQuery, stream api.StudyServiceApi_StreamParticipantStatesServer) error {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return s.missingArgumentError()
	}

	if !(token_checks.CheckRoleInToken(req.Token, constants.USER_ROLE_ADMIN) &&
		token_checks.CheckRoleInToken(req.Token, constants.USER_ROLE_RESEARCHER)) {
		err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id, []string{
			types.STUDY_ROLE_OWNER,
			types.STUDY_ROLE_MAINTAINER,
		})
		if err != nil {
			logger.Warning.Printf("unauthorizd access attempt to participant states: (%s-%s): %v", req.Token.TempToken.InstanceId, req.StudyKey, err)
			s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_DOWNLOAD_RESPONSES, "permission denied for "+req.StudyKey)
			return status.Error(codes.Internal, err.Error())
		}
	}

	sendResponseOverGrpc := func(db *studydb.StudyDBService, pState types.ParticipantState, instanceID string, studyKey string, args ...interface{}) error {
		if len(args) != 1 {
			return errors.New("StudyServiceApi_StreamParticipantStatesServer callback: unexpected number of args")
		}
		stream, ok := args[0].(api.StudyServiceApi_StreamParticipantStatesServer)
		if !ok {
			return errors.New(("StudyServiceApi_StreamParticipantStatesServer callback: can't parse stream"))
		}

		if err := stream.Send(pState.ToAPI()); err != nil {
			return err
		}
		return nil
	}

	ctx := context.Background()
	err := s.studyDBservice.FindAndExecuteOnParticipantsStates(
		ctx,
		req.Token.InstanceId,
		req.StudyKey,
		req.Status,
		sendResponseOverGrpc, stream)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_LOG, constants.LOG_EVENT_DOWNLOAD_RESPONSES, req.StudyKey)
	return nil
}

func (s *studyServiceServer) StreamReportHistory(req *api.ReportHistoryQuery, stream api.StudyServiceApi_StreamReportHistoryServer) error {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return s.missingArgumentError()
	}

	if !(token_checks.CheckRoleInToken(req.Token, constants.USER_ROLE_ADMIN) &&
		token_checks.CheckRoleInToken(req.Token, constants.USER_ROLE_RESEARCHER)) {
		err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id, []string{
			types.STUDY_ROLE_OWNER,
			types.STUDY_ROLE_MAINTAINER,
		})
		if err != nil {
			logger.Warning.Printf("unauthorizd access attempt to report history in (%s-%s): %v", req.Token.TempToken.InstanceId, req.StudyKey, err)
			s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_DOWNLOAD_RESPONSES, "permission denied for "+req.StudyKey)
			return status.Error(codes.Internal, err.Error())
		}
	}

	sendResponseOverGrpc := func(instanceID string, studyKey string, report types.Report, args ...interface{}) error {
		if len(args) != 1 {
			return errors.New("stream callback: unexpected number of args")
		}
		stream, ok := args[0].(api.StudyServiceApi_StreamReportHistoryServer)
		if !ok {
			return errors.New(("stream callback: can't parse stream"))
		}

		if err := stream.Send(report.ToAPI()); err != nil {
			return err
		}
		return nil
	}

	query := studydb.ReportQuery{
		ParticipantID: req.ParticipantId,
		Key:           req.ReportKey,
		Since:         req.From,
		Until:         req.Until,
	}
	ctx := context.Background()
	err := s.studyDBservice.PerformActionForReport(
		ctx,
		req.Token.InstanceId,
		req.StudyKey,
		query,
		sendResponseOverGrpc, stream)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_LOG, constants.LOG_EVENT_DOWNLOAD_RESPONSES, req.StudyKey)
	return nil
}

func (s *studyServiceServer) StreamParticipantFileInfos(req *api.FileInfoQuery, stream api.StudyServiceApi_StreamParticipantFileInfosServer) error {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return s.missingArgumentError()
	}

	if !(token_checks.CheckRoleInToken(req.Token, constants.USER_ROLE_ADMIN) &&
		token_checks.CheckRoleInToken(req.Token, constants.USER_ROLE_RESEARCHER)) {
		err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id, []string{
			types.STUDY_ROLE_OWNER,
			types.STUDY_ROLE_MAINTAINER,
		})
		if err != nil {
			logger.Warning.Printf("unauthorizd access attempt to participant file infos responses in (%s-%s): %v", req.Token.InstanceId, req.StudyKey, err)
			s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_DOWNLOAD_RESPONSES, "permission denied for "+req.StudyKey)
			return status.Error(codes.Internal, err.Error())
		}
	}

	sendResponseOverGrpc := func(instanceID string, studyKey string, fileInfo types.FileInfo, args ...interface{}) error {
		if len(args) != 1 {
			return errors.New("StreamParticipantFileInfos callback: unexpected number of args")
		}
		stream, ok := args[0].(api.StudyServiceApi_StreamParticipantFileInfosServer)
		if !ok {
			return errors.New(("StreamParticipantFileInfos callback: can't parse stream"))
		}

		if err := stream.Send(fileInfo.ToAPI()); err != nil {
			return err
		}
		return nil
	}

	ctx := context.Background()
	err := s.studyDBservice.PerformActionForFileInfos(
		ctx,
		req.Token.InstanceId,
		req.StudyKey,
		studydb.FileInfoQuery{
			ParticipantID: req.ParticipantId,
			FileType:      req.FileType,
			Since:         req.From,
			Until:         req.Until,
		},
		sendResponseOverGrpc, stream)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_LOG, constants.LOG_EVENT_DOWNLOAD_RESPONSES, req.StudyKey)
	return nil
}

func (s *studyServiceServer) StreamStudyResponses(req *api.SurveyResponseQuery, stream api.StudyServiceApi_StreamStudyResponsesServer) error {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return s.missingArgumentError()
	}

	if !(token_checks.CheckRoleInToken(req.Token, constants.USER_ROLE_ADMIN) &&
		token_checks.CheckRoleInToken(req.Token, constants.USER_ROLE_RESEARCHER)) {
		err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id, []string{
			types.STUDY_ROLE_OWNER,
			types.STUDY_ROLE_MAINTAINER,
		})
		if err != nil {
			logger.Warning.Printf("unauthorizd access attempt to survey responses in (%s-%s): %v", req.Token.InstanceId, req.StudyKey, err)
			s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_DOWNLOAD_RESPONSES, "permission denied for "+req.StudyKey)
			return status.Error(codes.Internal, err.Error())
		}
	}

	sendResponseOverGrpc := func(instanceID string, studyKey string, response types.SurveyResponse, args ...interface{}) error {
		if len(args) != 1 {
			return errors.New("StreamStudyResponses callback: unexpected number of args")
		}
		stream, ok := args[0].(api.StudyServiceApi_StreamStudyResponsesServer)
		if !ok {
			return errors.New(("StreamStudyResponses callback: can't parse stream"))
		}

		if err := stream.Send(response.ToAPI()); err != nil {
			return err
		}
		return nil
	}

	ctx := context.Background()
	err := s.studyDBservice.PerformActionForSurveyResponses(
		ctx,
		req.Token.InstanceId,
		req.StudyKey,
		req.SurveyKey,
		req.From,
		req.Until,
		sendResponseOverGrpc, stream)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_LOG, constants.LOG_EVENT_DOWNLOAD_RESPONSES, req.StudyKey)
	return nil
}

func (s *studyServiceServer) HasAccessToDownload(t *api_types.TokenInfos, studyKey string) error {
	// Admin and service account has access to studies:
	if token_checks.CheckIfAnyRolesInToken(t, []string{
		constants.USER_ROLE_ADMIN,
		constants.USER_ROLE_SERVICE_ACCOUNT,
	}) {
		return nil
	}

	// Should have at least researcher access:
	if !token_checks.CheckRoleInToken(t, constants.USER_ROLE_RESEARCHER) {
		return fmt.Errorf("unexpected roles %v", t.Payload["roles"])
	}

	// Should have access to the specific study as well
	err := s.HasRoleInStudy(t.InstanceId, studyKey, t.Id, []string{
		types.STUDY_ROLE_OWNER,
		types.STUDY_ROLE_MAINTAINER,
	})
	return err
}

// TODO: Test GetResponsesFlatJSON
func (s *studyServiceServer) GetResponsesFlatJSON(req *api.ResponseExportQuery, stream api.StudyServiceApi_GetResponsesFlatJSONServer) error {
	buf, err := s.getResponseExportBuffer(req, FLAT_JSON)

	if err != nil {
		return err
	}

	return StreamFile(stream, buf)
}

// TODO: Test GetResponsesWideFormatCSV
func (s *studyServiceServer) GetResponsesWideFormatCSV(req *api.ResponseExportQuery, stream api.StudyServiceApi_GetResponsesWideFormatCSVServer) error {
	buf, err := s.getResponseExportBuffer(req, WIDE_FORMAT_CSV)

	if err != nil {
		return err
	}

	return StreamFile(stream, buf)
}

// TODO: Test GetResponsesLongFormatCSV
func (s *studyServiceServer) GetResponsesLongFormatCSV(req *api.ResponseExportQuery, stream api.StudyServiceApi_GetResponsesLongFormatCSVServer) error {
	buf, err := s.getResponseExportBuffer(req, LONG_FORMAT_CSV)

	if err != nil {
		return err
	}

	return StreamFile(stream, buf)
}

// TODO: Test GetSurveyInfoPreviewCSV
func (s *studyServiceServer) GetSurveyInfoPreviewCSV(req *api.SurveyInfoExportQuery, stream api.StudyServiceApi_GetSurveyInfoPreviewCSVServer) error {
	responseExporter, err := s.getResponseExporterSurveyInfo(req)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	err = responseExporter.GetSurveyInfoCSV(buf)
	if err != nil {
		logger.Info.Println(err)
		return status.Error(codes.Internal, err.Error())
	}

	return StreamFile(stream, buf)
}

// TODO: Test GetSurveyInfoPreview
func (s *studyServiceServer) GetSurveyInfoPreview(ctx context.Context, req *api.SurveyInfoExportQuery) (*api.SurveyInfoExport, error) {
	responseExporter, err := s.getResponseExporterSurveyInfo(req)
	if err != nil {
		return nil, err
	}

	resp := responseExporter.GetSurveyVersionDefs()
	versions := make([]*api.SurveyVersionPreview, len(resp))
	for i, v := range resp {
		versions[i] = v.ToAPI()
	}

	return &api.SurveyInfoExport{
		Key:      req.SurveyKey,
		Versions: versions,
	}, nil
}

type StreamObj interface {
	Send(*api.Chunk) error
}

func StreamFile(stream StreamObj, buf *bytes.Buffer) error {
	chnk := &api.Chunk{}

	for currentByte := 0; currentByte < len(buf.Bytes()); currentByte += CHUNK_SIZE {
		if currentByte+CHUNK_SIZE > len(buf.Bytes()) {
			chnk.Chunk = buf.Bytes()[currentByte:len(buf.Bytes())]
		} else {
			chnk.Chunk = buf.Bytes()[currentByte : currentByte+CHUNK_SIZE]
		}

		if err := stream.Send(chnk); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}
	return nil
}

func (s *studyServiceServer) getResponseExportBuffer(req *api.ResponseExportQuery, fmt ResponseFormat) (*bytes.Buffer, error) {
	responseExporter, err := s.getResponseExporterResponseExport(req)
	if err != nil {
		return nil, err
	}
	logger.Info.Println("pageSize: ", req.PageSize, "pageNumber: ", req.Page)

	// Download responses
	ctx := context.Background()
	err = s.studyDBservice.PerformActionForSurveyResponses(
		ctx,
		req.Token.InstanceId, req.StudyKey, req.SurveyKey,
		req.From, req.Until, func(instanceID, studyKey string, response types.SurveyResponse, args ...interface{}) error {
			if len(args) != 3 {
				return errors.New("[getResponseExportBuffer]: wrong DB method argument")
			}
			rExp, ok := args[0].(*exporter.ResponseExporter)
			if !ok {
				return errors.New("[getResponseExportBuffer]: wrong DB method argument")
			}
			return rExp.AddResponse(&response)
		},
		responseExporter, req.Page, req.PageSize,
	)
	if err != nil {
		logger.Info.Print(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	buf := new(bytes.Buffer)
	includeMeta := &exporter.IncludeMeta{
		Postion:        req.IncludeMeta.Position,
		InitTimes:      req.IncludeMeta.InitTimes,
		ResponsedTimes: req.IncludeMeta.ResponsedTimes,
		DisplayedTimes: req.IncludeMeta.DisplayedTimes,
	}

	switch fmt {
	case FLAT_JSON:
		err = responseExporter.GetResponsesJSON(buf, includeMeta)
	case WIDE_FORMAT_CSV:
		err = responseExporter.GetResponsesCSV(buf, includeMeta)
	case LONG_FORMAT_CSV:
		err = responseExporter.GetResponsesLongFormatCSV(buf, includeMeta)
	default:
		return nil, status.Error(codes.Internal, errors.New("[getResponseExportBuffer]: wrong response format").Error())
	}

	if err != nil {
		logger.Info.Println(err)
		return nil, err
	}
	return buf, nil
}

func (s *studyServiceServer) getResponseExporterSurveyInfo(req *api.SurveyInfoExportQuery) (*exporter.ResponseExporter, error) {
	if req == nil {
		return nil, s.missingArgumentError()
	}
	return s.getResponseExporter(req.Token, req.StudyKey, req.SurveyKey, req.PreviewLanguage, req.ShortQuestionKeys, "ignored", nil)
}

func (s *studyServiceServer) getResponseExporterResponseExport(req *api.ResponseExportQuery) (*exporter.ResponseExporter, error) {
	if req == nil {
		return nil, s.missingArgumentError()
	}
	return s.getResponseExporter(req.Token, req.StudyKey, req.SurveyKey, "ignored", req.ShortQuestionKeys, req.Separator, req.ItemFilter)
}

func (s *studyServiceServer) getResponseExporter(
	token *api_types.TokenInfos,
	studyKey string,
	surveyKey string,
	previewLanguage string,
	shortQuestionKeys bool,
	separator string,
	itemFilter *api.ResponseExportQuery_ItemFilter,
) (*exporter.ResponseExporter, error) {
	if token_checks.IsTokenEmpty(token) || studyKey == "" {
		return nil, s.missingArgumentError()
	}

	if err := s.HasAccessToDownload(token, studyKey); err != nil {
		s.SaveLogEvent(token.InstanceId, token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_GET_SURVEY_DEF, "Permission denied for "+studyKey)
		return nil, status.Error(codes.Internal, err.Error())
	}

	surveyHistory, err := s.studyDBservice.FindSurveyDefHistory(token.InstanceId, studyKey, surveyKey, false)
	if err != nil {
		logger.Error.Println(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Init exporter:
	var responseExporter *exporter.ResponseExporter
	if itemFilter != nil {
		if itemFilter.Mode == api.ResponseExportQuery_ItemFilter_INCLUDE {
			responseExporter, err = exporter.NewResponseExporterWithIncludeFilter(
				surveyHistory,
				previewLanguage,
				shortQuestionKeys,
				separator,
				itemFilter.Keys,
			)
		} else {
			responseExporter, err = exporter.NewResponseExporterWithExcludeFilter(
				surveyHistory,
				previewLanguage,
				shortQuestionKeys,
				separator,
				itemFilter.Keys,
			)
		}
	} else {
		responseExporter, err = exporter.NewResponseExporter(
			surveyHistory,
			previewLanguage,
			shortQuestionKeys,
			separator,
		)
	}
	if err != nil {
		logger.Info.Println(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return responseExporter, nil
}

func (s *studyServiceServer) missingArgumentError() error {
	return status.Error(codes.InvalidArgument, "missing argument")
}
