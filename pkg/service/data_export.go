package service

import (
	"context"
	"log"

	"github.com/influenzanet/study-service/pkg/api"
	"github.com/influenzanet/study-service/pkg/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *studyServiceServer) GetStudyResponseStatistics(ctx context.Context, req *api.SurveyResponseQuery) (*api.StudyResponseStatistics, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id, []string{"analyst", "maintainer", "owner"})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
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

	return resp, nil
}

func (s *studyServiceServer) StreamStudyResponses(req *api.SurveyResponseQuery, stream api.StudyServiceApi_StreamStudyResponsesServer) error {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return status.Error(codes.InvalidArgument, "missing argument")
	}

	err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id, []string{"analyst", "maintainer", "owner"})
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return status.Error(codes.Unimplemented, "unimplemented")
}
