package main

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/influenzanet/study-service/api"
	"github.com/influenzanet/study-service/models"
	"github.com/influenzanet/study-service/utils"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Status endpoint should return internal status of the system if running correctly
func (s *studyServiceServer) Status(ctx context.Context, _ *empty.Empty) (*api.Status, error) {
	return &api.Status{
		Status: api.Status_NORMAL,
		Msg:    "service running",
	}, nil
}

func (s *studyServiceServer) AddSurveyToStudy(ctx context.Context, req *api.AddSurveyReq) (*api.SurveyVersion, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.StudyKey == "" || req.Survey == nil {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	members, err := getStudyMembers(req.Token.InstanceId, req.StudyKey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !utils.CheckIfMember(req.Token.Id, members, "") {
		return nil, status.Error(codes.Unauthenticated, "not authorized to access this study")
	}

	newSurvey := models.SurveyFromAPI(req.Survey)
	createdSurvey, err := addSurveyToDB(req.Token.InstanceId, req.StudyKey, newSurvey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.SurveyVersion{
		SurveyDefinition: createdSurvey.Current.SurveyDefinition.ToAPI(),
		Published:        createdSurvey.Current.Published,
	}, nil
}
