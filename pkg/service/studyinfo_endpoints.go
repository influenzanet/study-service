package service

import (
	"context"

	"github.com/influenzanet/study-service/pkg/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *studyServiceServer) GetStudiesForUser(ctx context.Context, req *api.TokenInfos) (*api.Studies, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (s *studyServiceServer) GetActiveStudies(ctx context.Context, req *api.TokenInfos) (*api.Studies, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (s *studyServiceServer) HasParticipantStateWithCondition(ctx context.Context, req *api.ProfilesWithConditionReq) (*api.AssignedSurveys, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}
