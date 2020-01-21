package main

import (
	"context"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/influenzanet/study-service/api"
	"github.com/influenzanet/study-service/models"
	"github.com/influenzanet/study-service/utils"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *studyServiceServer) Status(ctx context.Context, _ *empty.Empty) (*api.Status, error) {
	return &api.Status{
		Status: api.Status_NORMAL,
		Msg:    "service running",
	}, nil
}

func (s *studyServiceServer) CreateSurvey(ctx context.Context, req *api.CreateSurveyReq) (*api.SurveyVersion, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.SurveyDef == nil {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	newSurvey := models.Survey{
		Current: models.SurveyVersion{
			Published:        time.Now().Unix(),
			SurveyDefinition: models.SurveyItemFromAPI(req.SurveyDef),
		},
	}

	createdSurvey, err := addSurveyToDB(req.Token.InstanceId, newSurvey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.SurveyVersion{
		SurveyDefinition: createdSurvey.Current.SurveyDefinition.ToAPI(),
		Published:        createdSurvey.Current.Published,
	}, nil
}
