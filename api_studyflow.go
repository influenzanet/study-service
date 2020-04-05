package main

import (
	"context"
	"log"

	"github.com/influenzanet/study-service/api"
	"github.com/influenzanet/study-service/models"
	"github.com/influenzanet/study-service/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *studyServiceServer) EnterStudy(ctx context.Context, req *api.EnterStudyRequest) (*api.AssignedSurveys, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	// ParticipantID
	participantID, err := userIDToParticipantID(req.Token.InstanceId, req.StudyKey, req.Token.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Exists already?
	exists := checkIfParticipantExists(req.Token.InstanceId, req.StudyKey, participantID)
	if exists {
		log.Printf("error: participant (%s) already exists for this study", participantID)
		return nil, status.Error(codes.Internal, "participant already exists for this study")
	}

	// Init state and perform rules
	pState := models.ParticipantState{
		ParticipantID: participantID,
	}

	// perform study rules/actions
	currentEvent := models.StudyEvent{
		Type: "ENTER",
	}
	pState, err = getAndPerformStudyRules(req.Token.InstanceId, req.StudyKey, pState, currentEvent)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// save state back to DB
	pState, err = saveParticipantStateDB(req.Token.InstanceId, req.StudyKey, pState)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Prepare response
	resp := api.AssignedSurveys{
		Surveys: []*api.AssignedSurvey{},
	}
	for _, as := range pState.AssignedSurveys {
		cs := as.ToAPI()
		cs.StudyKey = req.StudyKey
		resp.Surveys = append(resp.Surveys, cs)
	}
	return &resp, nil
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
