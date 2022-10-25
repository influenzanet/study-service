package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/coneno/logger"
	"github.com/influenzanet/go-utils/pkg/api_types"
	"github.com/influenzanet/go-utils/pkg/constants"
	"github.com/influenzanet/go-utils/pkg/token_checks"
	loggingAPI "github.com/influenzanet/logging-service/pkg/api"
	"github.com/influenzanet/study-service/pkg/api"
	"github.com/influenzanet/study-service/pkg/dbs/studydb"
	"github.com/influenzanet/study-service/pkg/studyengine"
	"github.com/influenzanet/study-service/pkg/types"
	"github.com/influenzanet/study-service/pkg/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const TEMPORARY_PARTICIPANT_TAKEOVER_PERIOD = 60 * 60 // seconds

func (s *studyServiceServer) EnterStudy(ctx context.Context, req *api.EnterStudyRequest) (*api.AssignedSurveys, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if err := utils.CheckIfProfileIDinToken(req.Token, req.ProfileId); err != nil {
		s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_WRONG_PROFILE_ID, "enter study:"+req.ProfileId)
		return nil, status.Error(codes.Internal, "permission denied")
	}

	// ParticipantID
	participantID, participantID2, err := s.profileIDToParticipantID(req.Token.InstanceId, req.StudyKey, req.ProfileId, false)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Exists already?
	exists := s.checkIfParticipantExists(req.Token.InstanceId, req.StudyKey, participantID, types.PARTICIPANT_STUDY_STATUS_ACTIVE)
	if exists {
		logger.Debug.Printf("error: participant (%s) already exists for this study", participantID)
		return nil, status.Error(codes.Internal, "participant already exists for this study")
	}

	// Init state and perform rules
	pState := types.ParticipantState{
		ParticipantID: participantID,
		EnteredAt:     time.Now().Unix(),
		StudyStatus:   types.PARTICIPANT_STUDY_STATUS_ACTIVE,
	}

	// perform study rules/actions
	currentEvent := types.StudyEvent{
		Type:                                  "ENTER",
		InstanceID:                            req.Token.InstanceId,
		StudyKey:                              req.StudyKey,
		ParticipantIDForConfidentialResponses: participantID2,
	}
	actionResult, err := s.getAndPerformStudyRules(req.Token.InstanceId, req.StudyKey, pState, currentEvent)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// save state back to DB
	pState, err = s.studyDBservice.SaveParticipantState(req.Token.InstanceId, req.StudyKey, actionResult.PState)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	s.saveReports(req.Token.InstanceId, req.StudyKey, actionResult.ReportsToCreate, "ENTER")

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

func (s *studyServiceServer) RegisterTemporaryParticipant(ctx context.Context, req *api.RegisterTempParticipantReq) (*api.RegisterTempParticipantResponse, error) {
	if req == nil || req.StudyKey == "" || req.InstanceId == "" {
		logger.Debug.Println("missing argument in request")
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	// Generate new participantID
	participantID, _, err := s.profileIDToParticipantID(req.InstanceId, req.StudyKey, primitive.NewObjectID().Hex(), true)
	if err != nil {
		logger.Debug.Println(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Exists already?
	exists := s.checkIfParticipantExists(req.InstanceId, req.StudyKey, participantID, types.PARTICIPANT_STUDY_STATUS_ACTIVE)
	if exists {
		logger.Error.Printf("error: participant (%s) already exists for this study", participantID)
		return nil, status.Error(codes.Internal, "participant already exists for this study")
	}

	// Init state and perform rules
	pState := types.ParticipantState{
		ParticipantID: participantID,
		EnteredAt:     time.Now().Unix(),
		StudyStatus:   types.PARTICIPANT_STUDY_STATUS_TEMPORARY,
	}

	// perform study rules/actions
	currentEvent := types.StudyEvent{
		Type:       "ENTER",
		InstanceID: req.InstanceId,
		StudyKey:   req.StudyKey,
	}
	actionResult, err := s.getAndPerformStudyRules(req.InstanceId, req.StudyKey, pState, currentEvent)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// save state back to DB
	_, err = s.studyDBservice.SaveParticipantState(req.InstanceId, req.StudyKey, actionResult.PState)
	if err != nil {
		logger.Error.Println(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	s.saveReports(req.InstanceId, req.StudyKey, actionResult.ReportsToCreate, "ENTER")

	// Prepare response
	resp := api.RegisterTempParticipantResponse{
		TemporaryParticipantId: participantID,
		Timestamp:              pState.EnteredAt,
	}
	return &resp, nil
}

func (s *studyServiceServer) ConvertTemporaryToParticipant(ctx context.Context, req *api.ConvertTempParticipantReq) (*api.ServiceStatus, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" || req.TemporaryParticipantId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	// Find temporary participant:
	pState, err := s.studyDBservice.FindParticipantState(req.Token.InstanceId, req.StudyKey, req.TemporaryParticipantId)
	if err != nil ||
		pState.StudyStatus != types.PARTICIPANT_STUDY_STATUS_TEMPORARY ||
		pState.EnteredAt != req.Timestamp ||
		pState.EnteredAt+TEMPORARY_PARTICIPANT_TAKEOVER_PERIOD < time.Now().Unix() {
		// problem with temporary participant
		logger.Error.Printf("user (%s:%s) attempted to convert wrong temporary participant (ID: %s)", req.Token.InstanceId, req.Token.Id, req.TemporaryParticipantId)
		s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_PARTICIPANT_ACTION, fmt.Sprintf("user attempted to convert wrong temporary participant (ID: %s)", req.TemporaryParticipantId))
		time.Sleep(5 * time.Second)
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Calculate real participant ID:
	realParticipantID, realParticipantID2, err := s.profileIDToParticipantID(req.Token.InstanceId, req.StudyKey, req.ProfileId, false)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Exists already?
	exists := s.checkIfParticipantExists(req.Token.InstanceId, req.StudyKey, realParticipantID, types.PARTICIPANT_STUDY_STATUS_ACTIVE)
	if exists {
		existingPState, err := s.studyDBservice.FindParticipantState(req.Token.InstanceId, req.StudyKey, realParticipantID)
		if err != nil {
			logger.Debug.Println(err)
			return nil, status.Error(codes.Internal, err.Error())
		}

		// Merge participant states
		event := types.StudyEvent{
			InstanceID:                            req.Token.InstanceId,
			StudyKey:                              req.StudyKey,
			Type:                                  "MERGE",
			MergeWithParticipant:                  pState,
			ParticipantIDForConfidentialResponses: realParticipantID2,
		}
		mergeResult, err := s.getAndPerformStudyRules(req.Token.InstanceId, req.StudyKey, existingPState, event)
		if err != nil {
			logger.Error.Println(err)
			return nil, status.Error(codes.Internal, err.Error())
		}

		_, err = s.studyDBservice.SaveParticipantState(req.Token.InstanceId, req.StudyKey, mergeResult.PState)
		if err != nil {
			logger.Error.Println(err)
			return nil, status.Error(codes.Internal, err.Error())
		}
	} else {
		// Participant did not exist before:
		pState.ID = primitive.ObjectID{}
		pState.ParticipantID = realParticipantID
		pState.StudyStatus = types.PARTICIPANT_STUDY_STATUS_ACTIVE

		_, err = s.studyDBservice.SaveParticipantState(req.Token.InstanceId, req.StudyKey, pState)
		if err != nil {
			logger.Error.Println(err)
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	err = s.studyDBservice.DeleteParticipantState(req.Token.InstanceId, req.StudyKey, req.TemporaryParticipantId)
	if err != nil {
		logger.Error.Println(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	// update participant ID to all response object
	count, err := s.studyDBservice.UpdateParticipantIDonResponses(req.Token.InstanceId, req.StudyKey, req.TemporaryParticipantId, realParticipantID)
	if err != nil {
		logger.Error.Println(err)
	} else {
		logger.Debug.Printf("updated %d responses for participant %s", count, realParticipantID)
	}

	// update participant ID to all history object
	count, err = s.studyDBservice.UpdateParticipantIDonReports(req.Token.InstanceId, req.StudyKey, req.TemporaryParticipantId, realParticipantID)
	if err != nil {
		logger.Error.Println(err)
	} else {
		logger.Debug.Printf("updated %d reports for participant %s", count, realParticipantID)
	}

	// update participant ID to all confidential responses
	oldID, _, err := s.profileIDToParticipantID(req.Token.InstanceId, req.StudyKey, req.TemporaryParticipantId, true)
	if err != nil {
		logger.Error.Println(err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	count, err = s.studyDBservice.UpdateParticipantIDonConfidentialResponses(req.Token.InstanceId, req.StudyKey, oldID, realParticipantID2)
	if err != nil {
		logger.Error.Println(err)
	} else {
		logger.Debug.Printf("updated %d confidential responses for participant %s", count, realParticipantID)
	}

	return &api.ServiceStatus{
		Status:  api.ServiceStatus_NORMAL,
		Msg:     "conversion successful",
		Version: apiVersion,
	}, nil
}

func (s *studyServiceServer) GetAssignedSurveys(ctx context.Context, req *api_types.TokenInfos) (*api.AssignedSurveys, error) {
	if token_checks.IsTokenEmpty(req) {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	studies, err := s.studyDBservice.GetStudiesByStatus(req.InstanceId, types.STUDY_STATUS_ACTIVE, true)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// for every profile form the token
	profileIDs := []string{req.ProfilId}
	profileIDs = append(profileIDs, req.OtherProfileIds...)

	surveyCache := map[string]*types.Survey{}

	resp := api.AssignedSurveys{
		Surveys:     []*api.AssignedSurvey{},
		SurveyInfos: []*api.SurveyInfo{},
	}
	for _, study := range studies {
		for _, profileID := range profileIDs {
			participantID, err := utils.ProfileIDtoParticipantID(profileID, s.StudyGlobalSecret, study.SecretKey, study.Configs.IdMappingMethod)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			pState, err := s.studyDBservice.FindParticipantState(req.InstanceId, study.Key, participantID)
			if err != nil || pState.StudyStatus != types.PARTICIPANT_STUDY_STATUS_ACTIVE {
				continue
			}

			for _, as := range pState.AssignedSurveys {
				cs := as.ToAPI()
				cs.StudyKey = study.Key
				cs.ProfileId = profileID
				resp.Surveys = append(resp.Surveys, cs)

				cacheKey := req.InstanceId + study.Key + cs.SurveyKey
				sDef, ok := surveyCache[cacheKey]
				if !ok {
					sDef, err = s.studyDBservice.FindCurrentSurveyDef(req.InstanceId, study.Key, cs.SurveyKey, true)
					if err != nil {
						logger.Error.Printf("could not retrieve current survey defintions [%s:%s:%s]: %v", req.InstanceId, study.Key, cs.SurveyKey, err)
						continue
					}
					surveyCache[cacheKey] = sDef
				}

				found := false
				for _, info := range resp.SurveyInfos {
					if info.SurveyKey == sDef.SurveyDefinition.Key && info.StudyKey == cs.StudyKey {
						found = true
						break
					}
				}
				if !found {
					apiS := sDef.ToAPI()
					resp.SurveyInfos = append(resp.SurveyInfos, &api.SurveyInfo{
						StudyKey:        cs.StudyKey,
						SurveyKey:       apiS.SurveyDefinition.Key,
						Name:            apiS.Props.Name,
						Description:     apiS.Props.Description,
						TypicalDuration: apiS.Props.TypicalDuration,
					})
				}
			}
		}
	}

	return &resp, nil
}

func (s *studyServiceServer) GetAssignedSurveysForTemporaryParticipant(ctx context.Context, req *api.GetAssignedSurveysForTemporaryParticipantReq) (*api.AssignedSurveys, error) {
	if req == nil || req.StudyKey == "" || req.InstanceId == "" || req.TemporaryParticipantId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	pState, err := s.studyDBservice.FindParticipantState(req.InstanceId, req.StudyKey, req.TemporaryParticipantId)
	if err != nil || pState.StudyStatus != types.PARTICIPANT_STUDY_STATUS_TEMPORARY {
		logger.Warning.Printf("Attempt to access participant with wrong temporary ID (%s) - Err: %v", req.TemporaryParticipantId, err)
		time.Sleep(5 * time.Second)
		return nil, status.Error(codes.Internal, "wrong argument")
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

func (s *studyServiceServer) GetAssignedSurvey(ctx context.Context, req *api.SurveyReferenceRequest) (*api.SurveyAndContext, error) {
	if req == nil || req.StudyKey == "" || req.SurveyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if token_checks.IsTokenEmpty(req.Token) {
		if req.InstanceId == "" {
			return nil, status.Error(codes.InvalidArgument, "missing argument")
		}
		return s._getSurveyWithoutLogin(req.InstanceId, req.StudyKey, req.SurveyKey, req.TemporaryParticipantId)
	} else {
		return s._getSurveyWithLoggedInUser(req.Token, req.StudyKey, req.SurveyKey, req.ProfileId)
	}
}

func (s *studyServiceServer) SubmitResponse(ctx context.Context, req *api.SubmitResponseReq) (*api.AssignedSurveys, error) {
	if req == nil || req.StudyKey == "" || req.Response == nil || len(req.Response.Responses) < 1 {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	var participantID string
	var participantID2 string
	var instanceID string
	if req.Token != nil {
		if token_checks.IsTokenEmpty(req.Token) {
			return nil, status.Error(codes.InvalidArgument, "missing argument")
		}

		if err := utils.CheckIfProfileIDinToken(req.Token, req.ProfileId); err != nil {
			s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_WRONG_PROFILE_ID, "submit responses study:"+req.ProfileId)
			return nil, status.Error(codes.Internal, "permission denied")
		}

		var err error
		participantID, participantID2, err = s.profileIDToParticipantID(req.Token.InstanceId, req.StudyKey, req.ProfileId, false)
		if err != nil {
			return nil, status.Error(codes.Internal, "could not compute participant id")
		}
		instanceID = req.Token.InstanceId
	} else {
		if req.InstanceId == "" || req.TemporaryParticipantId == "" {
			return nil, status.Error(codes.InvalidArgument, "missing argument")
		}
		participantID = req.TemporaryParticipantId
		instanceID = req.InstanceId
		var err error
		participantID2, _, err = s.profileIDToParticipantID(instanceID, req.StudyKey, participantID, false)
		if err != nil {
			return nil, status.Error(codes.Internal, "could not compute participant id")
		}
	}

	// Participant state:
	pState, err := s.studyDBservice.FindParticipantState(instanceID, req.StudyKey, participantID)
	if err != nil {
		// If no participant yet, but logged in: ENTER study
		if !token_checks.IsTokenEmpty(req.Token) {
			_, err := s.EnterStudy(ctx, &api.EnterStudyRequest{
				Token:     req.Token,
				StudyKey:  req.StudyKey,
				ProfileId: req.ProfileId,
			})
			if err != nil {
				logger.Error.Printf("Unexpected error when submitting with non-participant user: %v", err)
				return nil, status.Error(codes.Internal, "couldn't enter study")
			}
			pState, err = s.studyDBservice.FindParticipantState(instanceID, req.StudyKey, participantID)
			if err != nil {
				logger.Error.Printf("Unexpected error when submitting with non-participant user: %v", err)
				return nil, status.Error(codes.Internal, "couldn't enter study")
			}
		} else {
			req.Response = nil
			logger.Error.Printf("Participant not found for request; %v", req)
			return nil, status.Error(codes.Internal, "participant state not found")
		}
	}

	if req.Token == nil {
		if pState.StudyStatus != types.PARTICIPANT_STUDY_STATUS_TEMPORARY {
			req.Response = nil
			logger.Error.Printf("Exptected temporary participant, but got: %v for request; %v", pState, req)
			return nil, status.Error(codes.Internal, "expected temporary participant")
		}
		if pState.EnteredAt != req.TemporaryParticipantTimestamp ||
			pState.EnteredAt+TEMPORARY_PARTICIPANT_TAKEOVER_PERIOD < time.Now().Unix() {
			// problem with temporary participant
			logger.Error.Printf("attempted to submit for wrong temporary participant (instance: %s, ID: %s)", instanceID, req.TemporaryParticipantId)
			time.Sleep(5 * time.Second)
			return nil, status.Error(codes.InvalidArgument, "wrong temporary participant")
		}
	} else {
		if pState.StudyStatus != types.PARTICIPANT_STUDY_STATUS_ACTIVE {
			req.Response = nil
			logger.Error.Printf("Exptected active participant, but got: %v for request; %v", pState, req)
			return nil, status.Error(codes.Internal, "user is not active in the current study")
		}
	}

	response := types.SurveyResponseFromAPI(req.Response)

	/**
	 * perform study rules/actions
	 */
	currentEvent := types.StudyEvent{
		Type:                                  "SUBMIT",
		Response:                              response,
		InstanceID:                            instanceID,
		StudyKey:                              req.StudyKey,
		ParticipantIDForConfidentialResponses: participantID2,
	}
	actionResult, err := s.getAndPerformStudyRules(instanceID, req.StudyKey, pState, currentEvent)
	if err != nil {
		logger.Error.Printf("unexpected error_ %v", err)
	}

	// save state back to DB
	pState, err = s.studyDBservice.SaveParticipantState(instanceID, req.StudyKey, actionResult.PState)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	/**
	 * Save response
	 */
	nonConfidentialResponses := []types.SurveyItemResponse{}
	confidentialResponses := []types.SurveyItemResponse{}

	// sort out data by confidentiality:
	for _, item := range response.Responses {
		if len(item.ConfidentialMode) > 0 {
			item.Meta = types.ResponseMeta{}
			confidentialResponses = append(confidentialResponses, item)
		} else {
			nonConfidentialResponses = append(nonConfidentialResponses, item)
		}
	}
	response.Responses = nonConfidentialResponses
	response.ParticipantID = participantID

	if response.Context == nil {
		response.Context = map[string]string{}
	}
	response.Context["session"] = pState.CurrentStudySession
	var rID string
	if len(nonConfidentialResponses) > 0 || len(confidentialResponses) < 1 {
		// Save responses only if non empty or there were no confidential responses
		rID, err = s.studyDBservice.AddSurveyResponse(instanceID, req.StudyKey, response)
		if err != nil {
			logger.Error.Printf("Unexpected error: %v", err)
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	// Save confidential data:
	if len(confidentialResponses) > 0 {
		for _, confItem := range confidentialResponses {
			rItem := types.SurveyResponse{
				Key:           confItem.Key,
				ParticipantID: participantID2,
				Responses:     []types.SurveyItemResponse{confItem},
			}
			if confItem.ConfidentialMode == "add" {
				_, err := s.studyDBservice.AddConfidentialResponse(instanceID, req.StudyKey, rItem)
				if err != nil {
					logger.Error.Printf("Unexpected error: %v", err)
				}
			} else {
				// Replace
				err := s.studyDBservice.ReplaceConfidentialResponse(instanceID, req.StudyKey, rItem)
				if err != nil {
					logger.Error.Printf("Unexpected error: %v", err)
				}
			}
		}
	}

	/**
	 * save reports
	 */
	s.saveReports(instanceID, req.StudyKey, actionResult.ReportsToCreate, rID)

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

func (s *studyServiceServer) LeaveStudy(ctx context.Context, req *api.LeaveStudyMsg) (*api.AssignedSurveys, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if err := utils.CheckIfProfileIDinToken(req.Token, req.ProfileId); err != nil {
		s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_WRONG_PROFILE_ID, "leave study:"+req.ProfileId)
		return nil, status.Error(codes.Internal, "permission denied")
	}

	// ParticipantID
	participantID, participantID2, err := s.profileIDToParticipantID(req.Token.InstanceId, req.StudyKey, req.ProfileId, false)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pState, err := s.studyDBservice.FindParticipantState(req.Token.InstanceId, req.StudyKey, participantID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if pState.StudyStatus != types.PARTICIPANT_STUDY_STATUS_ACTIVE {
		return nil, status.Error(codes.Internal, "not active in the study")
	}

	// Init state and perform rules
	pState = types.ParticipantState{
		ParticipantID: participantID,
		StudyStatus:   types.PARTICIPANT_STUDY_STATUS_EXITED,
	}
	// perform study rules/actions
	currentEvent := types.StudyEvent{
		Type:                                  "LEAVE",
		InstanceID:                            req.Token.InstanceId,
		StudyKey:                              req.StudyKey,
		ParticipantIDForConfidentialResponses: participantID2,
	}
	actionResult, err := s.getAndPerformStudyRules(req.Token.InstanceId, req.StudyKey, pState, currentEvent)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, err = s.studyDBservice.SaveParticipantState(req.Token.InstanceId, req.StudyKey, actionResult.PState)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	s.saveReports(req.Token.InstanceId, req.StudyKey, actionResult.ReportsToCreate, "LEAVE")

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

func (s *studyServiceServer) RemoveConfidentialResponsesForProfiles(ctx context.Context, req *api.RemoveConfidentialResponsesForProfilesReq) (*api.ServiceStatus, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	studies, err := s.studyDBservice.GetStudiesByStatus(req.Token.InstanceId, "", true)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	profileIDs := []string{}

	userProfileIDs := []string{req.Token.ProfilId}
	userProfileIDs = append(userProfileIDs, req.Token.OtherProfileIds...)

	if len(req.ForProfiles) > 0 {
		for _, p := range req.ForProfiles {
			for _, up := range userProfileIDs {
				if p == up {
					profileIDs = append(profileIDs, p)
					break
				}
			}
		}
		if len(profileIDs) < 1 {
			logger.Warning.Printf("User '%s' attempted to remove confidential data for profiles with insufficient rights in the token: %v", req.Token.Id, req.ForProfiles)
			return nil, status.Error(codes.PermissionDenied, "not permitted to manage requested profiles")
		}
	} else {
		profileIDs = userProfileIDs
	}

	for _, study := range studies {
		for _, profileID := range profileIDs {
			_, participantID2, err := s.profileIDToParticipantID(req.Token.InstanceId, study.Key, profileID, false)
			if err != nil {
				logger.Error.Printf("unexpected error: %v", err)
				continue
			}
			_, err = s.studyDBservice.DeleteConfidentialResponses(req.Token.InstanceId, study.Key, participantID2, "")
			if err != nil {
				logger.Error.Printf("unexpected error: %v", err)
			}
		}
	}
	return &api.ServiceStatus{
		Version: "v1",
		Msg:     "confidential data removal triggered",
	}, nil
}

func (s *studyServiceServer) DeleteParticipantData(ctx context.Context, req *api_types.TokenInfos) (*api.ServiceStatus, error) {
	if req == nil || token_checks.IsTokenEmpty(req) {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	studies, err := s.studyDBservice.GetStudiesByStatus(req.InstanceId, "", true)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	profileIDs := []string{req.ProfilId}
	profileIDs = append(profileIDs, req.OtherProfileIds...)

	for _, study := range studies {
		for _, profileID := range profileIDs {
			// ParticipantID
			participantID, participantID2, err := s.profileIDToParticipantID(req.InstanceId, study.Key, profileID, false)
			if err != nil {
				logger.Error.Printf("DeleteParticipantData: %v", err)
				continue
			}
			err = s.studyDBservice.DeleteParticipantState(req.InstanceId, study.Key, participantID)
			if err != nil {
				continue
			}
			_, err = s.studyDBservice.DeleteSurveyResponses(req.InstanceId, study.Key, studydb.ResponseQuery{ParticipantID: participantID})
			if err != nil {
				continue
			}
			_, err = s.studyDBservice.DeleteConfidentialResponses(req.InstanceId, study.Key, participantID2, "")
			if err != nil {
				continue
			}
		}

	}
	return &api.ServiceStatus{
		Status: api.ServiceStatus_NORMAL,
		Msg:    "all responses deleted",
	}, nil
}

func (s *studyServiceServer) CreateReport(ctx context.Context, req *api.CreateReportReq) (*api.ServiceStatus, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" || req.ProfileId == "" || req.Report == nil {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	var err error
	participantID, _, err := s.profileIDToParticipantID(req.Token.InstanceId, req.StudyKey, req.ProfileId, true)
	if err != nil {
		logger.Error.Printf("unexpected error: %v", err.Error())
		return nil, status.Error(codes.Internal, "could not compute participant id")
	}
	instanceID := req.Token.InstanceId

	if err := utils.CheckIfProfileIDinToken(req.Token, req.ProfileId); err != nil {
		s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_WRONG_PROFILE_ID, "create report:"+req.ProfileId)
		return nil, status.Error(codes.Internal, "permission denied")
	}

	rData := make([]types.ReportData, len(req.Report.Data))
	for i, d := range req.Report.Data {
		if d != nil {
			rData[i] = types.ReportData{
				Key:   d.Key,
				Value: d.Value,
				Dtype: d.Dtype,
			}
		}
	}
	report := types.Report{
		Key:           req.Report.Key,
		ParticipantID: participantID,
		ResponseID:    req.Report.ResponseId,
		Timestamp:     req.Report.Timestamp,
		Data:          rData,
	}

	err = s.studyDBservice.SaveReport(instanceID, req.StudyKey, report)
	if err != nil {
		logger.Error.Printf("unexpected error while save report: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	} else {
		logger.Debug.Printf("Report with key '%s' for participant %s saved.", report.Key, report.ParticipantID)
	}

	return &api.ServiceStatus{
		Status: api.ServiceStatus_NORMAL,
	}, nil
}

func (s *studyServiceServer) UploadParticipantFile(stream api.StudyServiceApi_UploadParticipantFileServer) error {
	maxParticipantFileSize := s.persistentStorageConfig.MaxParticipantFileSize
	req, err := stream.Recv()
	if err != nil {
		logger.Error.Println("Error: UploadParticipantFile missing file info")
		return status.Errorf(codes.Unknown, "file info missing")
	}

	info := req.GetInfo()
	if info == nil || token_checks.IsTokenEmpty(info.Token) || info.StudyKey == "" {
		return status.Error(codes.InvalidArgument, "missing argument")
	}

	instanceID := info.Token.InstanceId

	// Check file type
	if info.FileType == nil {
		return status.Error(codes.InvalidArgument, "file type missing")
	}

	// ParticipantID
	participantID := ""
	switch x := info.Participant.(type) {
	case *api.UploadParticipantFileReq_Info_ParticipantId:
		participantID = x.ParticipantId
		if !token_checks.CheckRoleInToken(info.Token, constants.USER_ROLE_ADMIN) {
			err := s.HasRoleInStudy(info.Token.InstanceId, info.StudyKey, info.Token.Id,
				[]string{types.STUDY_ROLE_MAINTAINER, types.STUDY_ROLE_OWNER},
			)
			if err != nil {
				s.SaveLogEvent(info.Token.InstanceId, info.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_RUN_CUSTOM_RULES, fmt.Sprintf("permission denied for uploading participant file in study %s  ", info.StudyKey))
				return status.Error(codes.Internal, err.Error())
			}
		}
	case *api.UploadParticipantFileReq_Info_ProfileId:
		if err := utils.CheckIfProfileIDinToken(info.Token, x.ProfileId); err != nil {
			s.SaveLogEvent(info.Token.InstanceId, info.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_WRONG_PROFILE_ID, " upload participant file:"+x.ProfileId)
			return status.Error(codes.Internal, "permission denied")
		}
		participantID, _, err = s.profileIDToParticipantID(instanceID, info.StudyKey, x.ProfileId, true)
		if err != nil {
			return status.Error(codes.Internal, "could not compute participant id")
		}
	default:
		errMsg := fmt.Sprintf("Participant has unexpected type %T", x)
		logger.Info.Printf("Error UploadParticipantFile: %s", errMsg)
		return status.Error(codes.InvalidArgument, errMsg)
	}

	pState, err := s.studyDBservice.FindParticipantState(instanceID, info.StudyKey, participantID)
	if err != nil {
		return status.Error(codes.Internal, "participant state not found")
	}
	if pState.StudyStatus != types.PARTICIPANT_STUDY_STATUS_ACTIVE {
		return status.Error(codes.Internal, "user is not active in the current study")
	}

	// get study upload condition rules
	studyDef, err := s.studyDBservice.GetStudyByStudyKey(instanceID, info.StudyKey)
	if err != nil {
		logger.Info.Printf("Error UploadParticipantFile: err at get study %v", err.Error())
		return status.Error(codes.Internal, "could not retrieve study")
	}
	if studyDef.Configs.ParticipantFileUploadRule == nil {
		s.SaveLogEvent(info.Token.InstanceId, info.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_SAVE_SURVEY, " upload participant file not permitted")
		return status.Error(codes.PermissionDenied, "no permission to upload files")
	} else {
		// check upload condition for participant
		val, err := studyengine.ExpressionEval(*studyDef.Configs.ParticipantFileUploadRule, studyengine.EvalContext{
			Event: types.StudyEvent{
				InstanceID: instanceID,
				StudyKey:   info.StudyKey,
				Type:       "FILE_UPLOAD",
				Response: types.SurveyResponse{
					Context: map[string]string{
						"fileType": info.FileType.Value,
					},
				},
			},
			ParticipantState: pState,
			Configs: studyengine.ActionConfigs{
				DBService:              s.studyDBservice,
				ExternalServiceConfigs: s.studyEngineExternalServices,
			},
		})
		if err != nil {
			logger.Info.Printf("Error UploadParticipantFile: err at eval rule %v", err.Error())
			return status.Error(codes.PermissionDenied, "no permission to upload files")
		}
		if !val.(bool) {
			return status.Error(codes.PermissionDenied, "no permission to upload files")
		}

	}

	tempPath := filepath.Join(s.persistentStorageConfig.RootPath, "temp")
	err = os.MkdirAll(tempPath, os.ModePerm)
	if err != nil {
		logger.Info.Printf("Error UploadParticipantFile: err at mkdir %v", err.Error())
	}

	// Create file reference entry in DB
	fileInfo, err := s.studyDBservice.SaveFileInfo(instanceID, info.StudyKey, types.FileInfo{
		ParticipantID: participantID,
		Status:        types.FILE_STATUS_UPLOADING,
		FileType:      info.FileType.Value,
	})
	if err != nil {
		logger.Error.Printf("Error UploadParticipantFile: %v", err.Error())
		return status.Error(codes.Internal, "unexpected error when creating file info object in DB.")
	}

	filename := fileInfo.ID.Hex()
	if info.FileType != nil && len(info.FileType.Subtype) > 0 {
		filename += "." + info.FileType.Subtype
	}

	fileSize := 0
	tempFileName := filepath.Join(tempPath, filename)
	var newFile *os.File
	newFile, err = os.Create(tempFileName)
	if err != nil {
		logger.Error.Printf("error at creating file: %s", err.Error())
		return status.Errorf(codes.Internal, "error at creating file: %s", err.Error())
	}

	for {
		logger.Debug.Print("waiting to receive more data")

		req, err := stream.Recv()
		if err == io.EOF {
			// no more data
			break
		}
		if err != nil {
			return status.Errorf(codes.Unknown, "cannot receive chunk data: %v", err)
		}

		chunk := req.GetChunk()
		size := len(chunk)

		logger.Debug.Printf("received a chunk with size: %d", size)

		fileSize += size
		if fileSize > maxParticipantFileSize {
			logger.Error.Printf("file is too large: %d > %d - for: %v", fileSize, maxParticipantFileSize, fileInfo)

			// remove DB reference
			_, err := s.studyDBservice.DeleteFileInfo(instanceID, info.StudyKey, fileInfo.ID.Hex())
			if err != nil {
				logger.Error.Printf("unexpected error: %v", err)
			}
			// remove temp file
			err = os.Remove(tempFileName)
			if err != nil {
				logger.Error.Printf("unexpected error: %v", err)
			}
			return status.Errorf(codes.InvalidArgument, "file is too large: %d > %d", fileSize, maxParticipantFileSize)
		}

		if newFile == nil {
			return status.Error(codes.Internal, "file handler object not found")
		}
		_, err = newFile.Write(chunk)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

	}
	if newFile == nil {
		return status.Error(codes.Internal, "file handler object not found")
	}
	newFile.Close()

	// move file to where it should be
	relativeTargetFolder := filepath.Join(instanceID, info.StudyKey)
	absoluteTargetFolder := filepath.Join(s.persistentStorageConfig.RootPath, relativeTargetFolder)
	targetFileRelativePath := filepath.Join(relativeTargetFolder, filename)
	targetFileAbsolutePath := filepath.Join(absoluteTargetFolder, filename)

	err = os.MkdirAll(absoluteTargetFolder, os.ModePerm)
	if err != nil {
		logger.Info.Printf("Error UploadParticipantFile: err at target path mkdir %v", err.Error())
	}
	err = os.Rename(tempFileName, targetFileAbsolutePath)
	if err != nil {
		logger.Info.Printf("Error UploadParticipantFile: err at moving target %v", err.Error())
	}

	// update file reference entry with path and finished upload
	fileInfo.Size = int32(fileSize)
	fileInfo.Status = types.FILE_STATUS_READY
	fileInfo.Path = targetFileRelativePath
	fileInfo, err = s.studyDBservice.SaveFileInfo(instanceID, info.StudyKey, fileInfo)
	if err != nil {
		logger.Debug.Printf("Error UploadParticipantFile: %v", err.Error())
	}
	// TODO: if necessary, start go process to generate preview

	// Remove infos not necessary for client:
	fileInfo.Path = ""
	fileInfo.PreviewPath = ""
	stream.SendAndClose(fileInfo.ToAPI())
	return nil
}

func (s *studyServiceServer) checkIfHasAccessToFile(token *api_types.TokenInfos, studyKey string, fileInfo types.FileInfo) bool {
	if token_checks.CheckRoleInToken(token, constants.USER_ROLE_ADMIN) {
		return true
	}

	// if not admin check if has right role:
	err := s.HasRoleInStudy(token.InstanceId, studyKey, token.Id,
		[]string{types.STUDY_ROLE_MAINTAINER, types.STUDY_ROLE_OWNER},
	)
	if err == nil {
		return true
	}

	// if not researcher
	userProfileIDs := []string{token.ProfilId}
	userProfileIDs = append(userProfileIDs, token.OtherProfileIds...)

	for _, profileID := range userProfileIDs {
		pID, _, err := s.profileIDToParticipantID(token.InstanceId, studyKey, profileID, true)
		if err != nil {
			logger.Error.Printf("unexpected error: %v", err)
			return false
		}
		if fileInfo.ParticipantID == pID {
			return true
		}
	}
	return false
}

func (s *studyServiceServer) GetParticipantFile(req *api.GetParticipantFileReq, stream api.StudyServiceApi_GetParticipantFileServer) error {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return s.missingArgumentError()
	}

	fileInfo, err := s.studyDBservice.FindFileInfo(req.Token.InstanceId, req.StudyKey, req.FileId)
	if err != nil {
		logger.Error.Printf("unexpected error: %v", err)
		return status.Error(codes.Internal, "file info not found")
	}

	if !s.checkIfHasAccessToFile(req.Token, req.StudyKey, fileInfo) {
		logger.Warning.Printf("unauthorizd access attempt to participant file in (%s-%s)", req.Token.TempToken.InstanceId, req.StudyKey)
		s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_DOWNLOAD_RESPONSES, "permission denied for "+req.StudyKey)
		return status.Error(codes.Internal, "persmission denied")
	}

	content, err := os.ReadFile(filepath.Join(s.persistentStorageConfig.RootPath, fileInfo.Path))
	if err != nil {
		logger.Error.Printf("unexpected error: %v", err)
		return status.Error(codes.Internal, "file not found")
	}

	return StreamFile(stream, bytes.NewBuffer(content))
}

func (s *studyServiceServer) DeleteParticipantFiles(ctx context.Context, req *api.DeleteParticipantFilesReq) (*api.ServiceStatus, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return nil, s.missingArgumentError()
	}

	for _, id := range req.FileIds {
		fileInfo, err := s.studyDBservice.FindFileInfo(req.Token.InstanceId, req.StudyKey, id)
		if err != nil {
			logger.Error.Printf("unexpected error: %v", err)
			return nil, status.Error(codes.Internal, "file info not found")
		}

		if !s.checkIfHasAccessToFile(req.Token, req.StudyKey, fileInfo) {
			logger.Warning.Printf("unauthorizd access attempt to participant file in (%s-%s)", req.Token.TempToken.InstanceId, req.StudyKey)
			s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_DOWNLOAD_RESPONSES, "permission denied for "+req.StudyKey)
			return nil, status.Error(codes.Internal, "persmission denied")
		}

		// delete file
		err = os.Remove(filepath.Join(s.persistentStorageConfig.RootPath, fileInfo.Path))
		if err != nil {
			logger.Error.Printf("unexpected error: %v", err)
			continue
		}
		os.Remove(filepath.Join(s.persistentStorageConfig.RootPath, fileInfo.PreviewPath))

		// remove file info
		c, err := s.studyDBservice.DeleteFileInfo(req.Token.InstanceId, req.StudyKey, id)
		if err != nil {
			logger.Error.Printf("unexpected error: %v", err)
			continue
		}
		logger.Debug.Printf("%d file info removed", c)
	}
	return &api.ServiceStatus{}, nil
}
