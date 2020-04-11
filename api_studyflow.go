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
	participantID, err := userIDToParticipantID(req.Token.InstanceId, req.StudyKey, req.Token.ProfilId)
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
		StudyStatus:   "active",
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

	studies, err := getStudiesByStatus(req.InstanceId, "active", true)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := api.AssignedSurveys{
		Surveys: []*api.AssignedSurvey{},
	}
	for _, study := range studies {
		participantID, err := utils.UserIDtoParticipantID(req.ProfilId, conf.Study.GlobalSecret, study.SecretKey)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		pState, err := findParticipantStateDB(req.InstanceId, study.Key, participantID)
		if err != nil || pState.StudyStatus != "active" {
			continue
		}

		for _, as := range pState.AssignedSurveys {
			cs := as.ToAPI()
			cs.StudyKey = study.Key
			resp.Surveys = append(resp.Surveys, cs)
		}
	}

	return &resp, nil
}

func (s *studyServiceServer) GetAssignedSurvey(ctx context.Context, req *api.SurveyReferenceRequest) (*api.SurveyAndContext, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.StudyKey == "" || req.SurveyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	// ParticipantID
	participantID, err := userIDToParticipantID(req.Token.InstanceId, req.StudyKey, req.Token.ProfilId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Get survey definition
	surveyDef, err := findSurveyDefDB(req.Token.InstanceId, req.StudyKey, req.SurveyKey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	surveyContext, err := resolveContextRules(req.Token.InstanceId, req.StudyKey, participantID, *surveyDef.ContextRules)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	prefill, err := resolvePrefillRules(req.Token.InstanceId, req.StudyKey, participantID, surveyDef.PrefillRules)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// empty irrelevant fields for this purpose
	surveyDef.ContextRules = &models.SurveyContextDef{}
	surveyDef.PrefillRules = []models.Expression{}
	surveyDef.History = []models.SurveyVersion{}

	resp := api.SurveyAndContext{
		Survey:  surveyDef.ToAPI(),
		Context: surveyContext.ToAPI(),
		Prefill: prefill.ToAPI(),
	}
	return &resp, nil
}

func (s *studyServiceServer) SubmitStatusReport(ctx context.Context, req *api.StatusReportRequest) (*api.AssignedSurveys, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.StatusSurvey == nil {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	studies, err := getStudiesByStatus(req.Token.InstanceId, "active", true)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	resp := api.AssignedSurveys{
		Surveys: []*api.AssignedSurvey{},
	}
	for _, study := range studies {
		participantID, err := utils.UserIDtoParticipantID(req.Token.ProfilId, conf.Study.GlobalSecret, study.SecretKey)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		pState, err := findParticipantStateDB(req.Token.InstanceId, study.Key, participantID)
		if err != nil {
			log.Println(err)
			continue
		}

		if pState.StudyStatus != "active" {
			continue
		}

		// Save responses
		response := models.SurveyResponseFromAPI(req.StatusSurvey)
		response.ParticipantID = participantID
		err = addSurveyResponseToDB(req.Token.InstanceId, study.Key, response)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		// perform study rules/actions
		currentEvent := models.StudyEvent{
			Type:     "SUBMIT",
			Response: response,
		}
		pState, err = getAndPerformStudyRules(req.Token.InstanceId, study.Key, pState, currentEvent)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		// save state back to DB
		pState, err = saveParticipantStateDB(req.Token.InstanceId, study.Key, pState)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		for _, as := range pState.AssignedSurveys {
			cs := as.ToAPI()
			cs.StudyKey = study.Key
			resp.Surveys = append(resp.Surveys, cs)
		}
	}
	return &resp, nil
}

func (s *studyServiceServer) SubmitResponse(ctx context.Context, req *api.SubmitResponseReq) (*api.AssignedSurveys, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.StudyKey == "" || req.Response == nil {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	// ParticipantID
	participantID, err := userIDToParticipantID(req.Token.InstanceId, req.StudyKey, req.Token.ProfilId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pState, err := findParticipantStateDB(req.Token.InstanceId, req.StudyKey, participantID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if pState.StudyStatus != "active" {
		return nil, status.Error(codes.Internal, "user is not active in the current study")
	}

	// Save responses
	response := models.SurveyResponseFromAPI(req.Response)
	response.ParticipantID = participantID
	err = addSurveyResponseToDB(req.Token.InstanceId, req.StudyKey, response)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// perform study rules/actions
	currentEvent := models.StudyEvent{
		Type:     "SUBMIT",
		Response: response,
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
