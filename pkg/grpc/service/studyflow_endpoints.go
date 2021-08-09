package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/influenzanet/go-utils/pkg/api_types"
	"github.com/influenzanet/go-utils/pkg/constants"
	"github.com/influenzanet/go-utils/pkg/token_checks"
	loggingAPI "github.com/influenzanet/logging-service/pkg/api"
	"github.com/influenzanet/study-service/pkg/api"
	"github.com/influenzanet/study-service/pkg/dbs/studydb"
	"github.com/influenzanet/study-service/pkg/studyengine"
	"github.com/influenzanet/study-service/pkg/types"
	"github.com/influenzanet/study-service/pkg/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *studyServiceServer) EnterStudy(ctx context.Context, req *api.EnterStudyRequest) (*api.AssignedSurveys, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if err := utils.CheckIfProfileIDinToken(req.Token, req.ProfileId); err != nil {
		s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_WRONG_PROFILE_ID, "enter study:"+req.ProfileId)
		return nil, status.Error(codes.Internal, "permission denied")
	}

	// ParticipantID
	participantID, err := s.profileIDToParticipantID(req.Token.InstanceId, req.StudyKey, req.ProfileId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Exists already?
	exists := s.checkIfParticipantExists(req.Token.InstanceId, req.StudyKey, participantID, "active")
	if exists {
		log.Printf("error: participant (%s) already exists for this study", participantID)
		return nil, status.Error(codes.Internal, "participant already exists for this study")
	}

	// Init state and perform rules
	pState := types.ParticipantState{
		ParticipantID: participantID,
		EnteredAt:     time.Now().Unix(),
		StudyStatus:   "active",
	}

	// perform study rules/actions
	currentEvent := types.StudyEvent{
		Type:       "ENTER",
		InstanceID: req.Token.InstanceId,
		StudyKey:   req.StudyKey,
	}
	pState, err = s.getAndPerformStudyRules(req.Token.InstanceId, req.StudyKey, pState, currentEvent)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// save state back to DB
	pState, err = s.studyDBservice.SaveParticipantState(req.Token.InstanceId, req.StudyKey, pState)
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

	resp := api.AssignedSurveys{
		Surveys:     []*api.AssignedSurvey{},
		SurveyInfos: []*api.SurveyInfo{},
	}
	for _, study := range studies {
		studySurveys, err := s.studyDBservice.FindAllSurveyDefsForStudy(req.InstanceId, study.Key, false)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		for _, profileID := range profileIDs {

			participantID, err := utils.ProfileIDtoParticipantID(profileID, s.StudyGlobalSecret, study.SecretKey)
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

				sDef := types.Survey{}
				for _, def := range studySurveys {
					if def.Current.SurveyDefinition.Key == cs.SurveyKey {
						sDef = def
						break
					}
				}

				found := false
				for _, info := range resp.SurveyInfos {
					if info.SurveyKey == sDef.Current.SurveyDefinition.Key && info.StudyKey == cs.StudyKey {
						found = true
						break
					}
				}
				if !found {
					apiS := sDef.ToAPI()
					resp.SurveyInfos = append(resp.SurveyInfos, &api.SurveyInfo{
						StudyKey:        cs.StudyKey,
						SurveyKey:       apiS.Current.SurveyDefinition.Key,
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

func (s *studyServiceServer) GetAssignedSurvey(ctx context.Context, req *api.SurveyReferenceRequest) (*api.SurveyAndContext, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" || req.SurveyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if err := utils.CheckIfProfileIDinToken(req.Token, req.ProfileId); err != nil {
		s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_WRONG_PROFILE_ID, "get assigned survey:"+req.ProfileId)
		return nil, status.Error(codes.Internal, "permission denied")
	}

	// ParticipantID
	participantID, err := s.profileIDToParticipantID(req.Token.InstanceId, req.StudyKey, req.ProfileId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Get survey definition
	surveyDef, err := s.studyDBservice.FindSurveyDef(req.Token.InstanceId, req.StudyKey, req.SurveyKey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	surveyContext, err := s.resolveContextRules(req.Token.InstanceId, req.StudyKey, participantID, surveyDef.ContextRules)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	prefill, err := s.resolvePrefillRules(req.Token.InstanceId, req.StudyKey, participantID, surveyDef.PrefillRules)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// empty irrelevant fields for this purpose
	surveyDef.ContextRules = nil
	surveyDef.PrefillRules = []types.Expression{}
	surveyDef.History = []types.SurveyVersion{}

	resp := api.SurveyAndContext{
		Survey:  surveyDef.ToAPI(),
		Context: surveyContext.ToAPI(),
	}
	if len(prefill.Responses) > 0 {
		resp.Prefill = prefill.ToAPI()
	}
	return &resp, nil
}

func (s *studyServiceServer) PostponeSurvey(ctx context.Context, req *api.PostponeSurveyRequest) (*api.AssignedSurveys, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" || req.SurveyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if err := utils.CheckIfProfileIDinToken(req.Token, req.ProfileId); err != nil {
		s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_WRONG_PROFILE_ID, "postpone survey:"+req.ProfileId)
		return nil, status.Error(codes.Internal, "permission denied")
	}

	// ParticipantID
	participantID, err := s.profileIDToParticipantID(req.Token.InstanceId, req.StudyKey, req.ProfileId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pState, err := s.studyDBservice.FindParticipantState(req.Token.InstanceId, req.StudyKey, participantID)
	if err != nil {
		log.Println("PostponeSurvey: participant state not found")
		return nil, status.Error(codes.Internal, err.Error())
	}

	for i, as := range pState.AssignedSurveys {
		if as.SurveyKey == req.SurveyKey {
			newValidFrom := time.Now().Unix() + req.Delay

			if as.ValidUntil > 0 {
				if newValidFrom > as.ValidUntil-1800 {
					// submit survey as empty
					emptyResponse := types.SurveyResponse{
						Key:           req.SurveyKey,
						ParticipantID: participantID,
						SubmittedAt:   time.Now().Unix(),
						ArrivedAt:     time.Now().Unix(),
					}
					// perform study rules/actions
					currentEvent := types.StudyEvent{
						Type:       "SUBMIT",
						Response:   emptyResponse,
						InstanceID: req.Token.InstanceId,
						StudyKey:   req.StudyKey,
					}
					pState, err = s.getAndPerformStudyRules(req.Token.InstanceId, req.StudyKey, pState, currentEvent)
					if err != nil {
						return nil, status.Error(codes.Internal, err.Error())
					}
					break
				}
			}
			pState.AssignedSurveys[i].ValidFrom = newValidFrom
		}
	}

	// save state back to DB
	pState, err = s.studyDBservice.SaveParticipantState(req.Token.InstanceId, req.StudyKey, pState)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

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

func (s *studyServiceServer) SubmitStatusReport(ctx context.Context, req *api.StatusReportRequest) (*api.AssignedSurveys, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StatusSurvey == nil {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if err := utils.CheckIfProfileIDinToken(req.Token, req.ProfileId); err != nil {
		s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_WRONG_PROFILE_ID, "submit status report:"+req.ProfileId)
		return nil, status.Error(codes.Internal, "permission denied")
	}

	studies, err := s.studyDBservice.GetStudiesByStatus(req.Token.InstanceId, "active", true)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	resp := api.AssignedSurveys{
		Surveys: []*api.AssignedSurvey{},
	}
	for _, study := range studies {
		participantID, err := utils.ProfileIDtoParticipantID(req.ProfileId, s.StudyGlobalSecret, study.SecretKey)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		pState, err := s.studyDBservice.FindParticipantState(req.Token.InstanceId, study.Key, participantID)
		if err != nil {
			// user not in the study - log.Println(err)
			continue
		}

		if pState.StudyStatus != types.PARTICIPANT_STUDY_STATUS_ACTIVE {
			continue
		}

		// Save responses
		response := types.SurveyResponseFromAPI(req.StatusSurvey)
		response.ParticipantID = participantID
		err = s.studyDBservice.AddSurveyResponse(req.Token.InstanceId, study.Key, response)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		// perform study rules/actions
		currentEvent := types.StudyEvent{
			Type:       "SUBMIT",
			Response:   response,
			InstanceID: req.Token.InstanceId,
			StudyKey:   study.Key,
		}
		pState, err = s.getAndPerformStudyRules(req.Token.InstanceId, study.Key, pState, currentEvent)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		// save state back to DB
		pState, err = s.studyDBservice.SaveParticipantState(req.Token.InstanceId, study.Key, pState)
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
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" || req.Response == nil || len(req.Response.Responses) < 1 {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if err := utils.CheckIfProfileIDinToken(req.Token, req.ProfileId); err != nil {
		s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_WRONG_PROFILE_ID, "submit responses study:"+req.ProfileId)
		return nil, status.Error(codes.Internal, "permission denied")
	}

	// ParticipantID
	participantID, err := s.profileIDToParticipantID(req.Token.InstanceId, req.StudyKey, req.ProfileId)
	if err != nil {
		return nil, status.Error(codes.Internal, "could not compute participant id")
	}

	pState, err := s.studyDBservice.FindParticipantState(req.Token.InstanceId, req.StudyKey, participantID)
	if err != nil {
		return nil, status.Error(codes.Internal, "participant state not found")
	}
	if pState.StudyStatus != types.PARTICIPANT_STUDY_STATUS_ACTIVE {
		return nil, status.Error(codes.Internal, "user is not active in the current study")
	}

	// Save responses
	response := types.SurveyResponseFromAPI(req.Response)
	response.ParticipantID = participantID
	err = s.studyDBservice.AddSurveyResponse(req.Token.InstanceId, req.StudyKey, response)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// perform study rules/actions
	currentEvent := types.StudyEvent{
		Type:       "SUBMIT",
		Response:   response,
		InstanceID: req.Token.InstanceId,
		StudyKey:   req.StudyKey,
	}
	pState, err = s.getAndPerformStudyRules(req.Token.InstanceId, req.StudyKey, pState, currentEvent)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// save state back to DB
	pState, err = s.studyDBservice.SaveParticipantState(req.Token.InstanceId, req.StudyKey, pState)
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

func (s *studyServiceServer) LeaveStudy(ctx context.Context, req *api.LeaveStudyMsg) (*api.AssignedSurveys, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if err := utils.CheckIfProfileIDinToken(req.Token, req.ProfileId); err != nil {
		s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_WRONG_PROFILE_ID, "leave study:"+req.ProfileId)
		return nil, status.Error(codes.Internal, "permission denied")
	}

	// ParticipantID
	participantID, err := s.profileIDToParticipantID(req.Token.InstanceId, req.StudyKey, req.ProfileId)
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
		Type:       "LEAVE",
		InstanceID: req.Token.InstanceId,
		StudyKey:   req.StudyKey,
	}
	pState, err = s.getAndPerformStudyRules(req.Token.InstanceId, req.StudyKey, pState, currentEvent)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, err = s.studyDBservice.SaveParticipantState(req.Token.InstanceId, req.StudyKey, pState)
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
			participantID, err := s.profileIDToParticipantID(req.InstanceId, study.Key, profileID)
			if err != nil {
				log.Printf("DeleteParticipantData: %v", err)
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
		}

	}
	return &api.ServiceStatus{
		Status: api.ServiceStatus_NORMAL,
		Msg:    "all responses deleted",
	}, nil
}

const maxParticipantFileSize = 1 << 25

func (s *studyServiceServer) UploadParticipantFile(stream api.StudyServiceApi_UploadParticipantFileServer) error {
	req, err := stream.Recv()
	if err != nil {
		log.Println("Error: UploadParticipantFile missing file info")
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
		participantID, err = s.profileIDToParticipantID(instanceID, info.StudyKey, x.ProfileId)
		if err != nil {
			return status.Error(codes.Internal, "could not compute participant id")
		}
	default:
		errMsg := fmt.Sprintf("Participant has unexpected type %T", x)
		log.Printf("Error UploadParticipantFile: %s", errMsg)
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
		log.Printf("Error UploadParticipantFile: err at get study %v", err.Error())
		return status.Error(codes.Internal, "could not retrieve study")
	}
	if studyDef.ParticipantFileUploadRule == nil {
		s.SaveLogEvent(info.Token.InstanceId, info.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_SAVE_SURVEY, " upload participant file not permitted")
		return status.Error(codes.PermissionDenied, "no permission to upload files")
	} else {
		// TODO: check upload condition for participant
		val, err := studyengine.ExpressionEval(*studyDef.ParticipantFileUploadRule, studyengine.EvalContext{
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
			DbService:        s.studyDBservice,
		})
		if err != nil {
			log.Printf("Error UploadParticipantFile: err at eval rule %v", err.Error())
			return status.Error(codes.PermissionDenied, "no permission to upload files")
		}
		if !val.(bool) {
			return status.Error(codes.PermissionDenied, "no permission to upload files")
		}

	}

	rootPath := "todo"
	tempPath := filepath.Join(rootPath, "temp")
	err = os.MkdirAll(tempPath, os.ModePerm)
	if err != nil {
		log.Printf("Error UploadParticipantFile: err at mkdir %v", err.Error())
	}

	// TODO create file reference entry in DB
	fileInfo, err := s.studyDBservice.SaveFileInfo(instanceID, info.StudyKey, types.FileInfo{
		ParticipantID: participantID,
		Status:        types.FILE_STATUS_UPLOADING,
		FileType:      info.FileType.Value,
	})
	if err != nil {
		log.Printf("Error UploadParticipantFile: %v", err.Error())
		return status.Error(codes.Internal, "unexpected error when creating file info object in DB.")
	}

	filename := fileInfo.ID.Hex()
	if info.FileType != nil && len(info.FileType.Subtype) > 0 {
		filename += "." + info.FileType.Subtype
	}

	fileSize := 0
	var newFile *os.File
	newFile, err = os.Create(filepath.Join(tempPath, filename))
	if err != nil {
		log.Printf("error at creating file: %s", err.Error())
		return status.Error(codes.Internal, "todo")
	}

	for {
		log.Print("waiting to receive more data")

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

		log.Printf("received a chunk with size: %d", size)

		fileSize += size
		if fileSize > maxParticipantFileSize {
			// TODO: remove temp file and DB reference
			return status.Errorf(codes.InvalidArgument, "file is too large: %d > %d", fileSize, maxParticipantFileSize)
		}

		if newFile == nil {
			return status.Error(codes.Internal, "todo")
		}
		_, err = newFile.Write(chunk)
		if err != nil {
			return status.Error(codes.Internal, "todo")
		}

	}
	if newFile == nil {
		return status.Error(codes.Internal, "todo")
	}
	newFile.Close()

	// TODO: move file to where it should be
	// TODO: update file reference entry with path and finished upload
	fileInfo.Size = int32(fileSize)
	fileInfo.Status = types.FILE_STATUS_READY
	fileInfo, err = s.studyDBservice.SaveFileInfo(instanceID, info.StudyKey, fileInfo)
	if err != nil {
		log.Printf("Error UploadParticipantFile: %v", err.Error())
	}
	// TODO: if necessary, start go process to generate preview

	// Remove infos not necessary for client:
	fileInfo.Path = ""
	fileInfo.PreviewPath = ""
	stream.SendAndClose(fileInfo.ToAPI())
	return nil
}
