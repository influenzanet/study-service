package main

import (
	"context"

	"github.com/influenzanet/study-service/api"
	"github.com/influenzanet/study-service/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *studyServiceServer) EnterStudy(ctx context.Context, req *api.EnterStudyRequest) (*api.AssignedSurveys, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	return nil, status.Error(codes.Unimplemented, "unimplmented")
}

func (s *studyServiceServer) GetAssignedSurveys(ctx context.Context, req *api.TokenInfos) (*api.AssignedSurveys, error) {
	if utils.IsTokenEmpty(req) {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	return nil, status.Error(codes.Unimplemented, "unimplmented")
}

func (s *studyServiceServer) GetAssignedSurvey(ctx context.Context, req *api.GetSurveyRequest) (*api.SurveyAndContext, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.StudyKey == "" || req.SurveyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	return nil, status.Error(codes.Unimplemented, "unimplmented")
}

func (s *studyServiceServer) SubmitStatusReport(ctx context.Context, req *api.StatusReportRequest) (*api.AssignedSurveys, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.StatusSurvey == nil {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	return nil, status.Error(codes.Unimplemented, "unimplmented")
}

func (s *studyServiceServer) SubmitResponse(ctx context.Context, req *api.SubmitResponseReq) (*api.Status, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.StudyKey == "" || req.Response == nil {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	return nil, status.Error(codes.Unimplemented, "unimplmented")
	/*


		report := models.SurveyResponseReport{
			By:          req.Token.Id,
			For:         req.ProfileId,
			SubmittedAt: time.Now().Unix(),
			Responses:   models.SurveyItemResponseFromAPI(req.Responses),
		}

		if err := addSurveyResponseToDB(req.Token.InstanceId, req.StudyId, report); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		return &api.Status{
			Status: api.Status_NORMAL,
			Msg:    "report successfully submitted",
		}, nil
	*/
}
