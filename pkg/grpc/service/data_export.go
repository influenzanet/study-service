package service

import (
	"context"
	"errors"
	"log"

	"github.com/influenzanet/go-utils/pkg/constants"
	"github.com/influenzanet/go-utils/pkg/token_checks"
	"github.com/influenzanet/study-service/pkg/api"
	"github.com/influenzanet/study-service/pkg/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	loggingAPI "github.com/influenzanet/logging-service/pkg/api"
)

func (s *studyServiceServer) GetStudyResponseStatistics(ctx context.Context, req *api.SurveyResponseQuery) (*api.StudyResponseStatistics, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
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
			log.Printf("GetStudyResponseStatistics: unexpected error: %v", err)
			continue
		}
		resp.SurveyResponseCounts[k] = count
	}
	s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_LOG, constants.LOG_EVENT_DOWNLOAD_RESPONSES, "statistics: "+req.StudyKey)
	return resp, nil
}

func (s *studyServiceServer) StreamStudyResponses(req *api.SurveyResponseQuery, stream api.StudyServiceApi_StreamStudyResponsesServer) error {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return status.Error(codes.InvalidArgument, "missing argument")
	}

	if !(token_checks.CheckRoleInToken(req.Token, constants.USER_ROLE_ADMIN) &&
		token_checks.CheckRoleInToken(req.Token, constants.USER_ROLE_RESEARCHER)) {
		err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id, []string{
			types.STUDY_ROLE_OWNER,
			types.STUDY_ROLE_MAINTAINER,
		})
		if err != nil {
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

	err := s.studyDBservice.PerfomActionForSurveyResponses(req.Token.InstanceId, req.StudyKey, req.SurveyKey, req.From, req.Until,
		sendResponseOverGrpc, stream)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_LOG, constants.LOG_EVENT_DOWNLOAD_RESPONSES, req.StudyKey)
	return nil
}
