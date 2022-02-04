package service

import (
	"context"
	"time"

	"github.com/coneno/logger"
	"github.com/influenzanet/go-utils/pkg/api_types"
	"github.com/influenzanet/go-utils/pkg/token_checks"
	"github.com/influenzanet/study-service/pkg/api"
	"github.com/influenzanet/study-service/pkg/dbs/studydb"
	"github.com/influenzanet/study-service/pkg/studyengine"
	"github.com/influenzanet/study-service/pkg/types"
	"github.com/influenzanet/study-service/pkg/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *studyServiceServer) GetStudiesForUser(ctx context.Context, req *api.GetStudiesForUserReq) (*api.StudiesForUser, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	// for every profile form the token
	profileIDs := []string{req.Token.ProfilId}
	profileIDs = append(profileIDs, req.Token.OtherProfileIds...)

	studies, err := s.studyDBservice.GetStudiesByStatus(req.Token.InstanceId, "", false)
	if err != nil {
		logger.Info.Println(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &api.StudiesForUser{Studies: []*api.StudyForUser{}}
	for _, study := range studies {
		studyInfos := api.StudyForUser{
			Key:        study.Key,
			Props:      study.Props.ToAPI(),
			Status:     study.Status,
			Stats:      study.Stats.ToAPI(),
			ProfileIds: []string{},
		}
		for _, profileID := range profileIDs {
			// ParticipantID
			participantID, err := utils.ProfileIDtoParticipantID(profileID, s.StudyGlobalSecret, study.SecretKey, study.Configs.IdMappingMethod)
			if err != nil {
				continue
			}

			pState, err := s.studyDBservice.FindParticipantState(req.Token.InstanceId, study.Key, participantID)
			if err != nil {
				// user not in the study
				continue
			}

			if pState.StudyStatus != types.PARTICIPANT_STUDY_STATUS_ACTIVE {
				continue
			}

			// at least one profile in the study:
			studyInfos.ProfileIds = append(studyInfos.ProfileIds, profileID)
		}
		if len(studyInfos.ProfileIds) > 0 {
			resp.Studies = append(resp.Studies, &studyInfos)
		}
	}

	return resp, nil
}

func (s *studyServiceServer) GetActiveStudies(ctx context.Context, req *api_types.TokenInfos) (*api.Studies, error) {
	if req == nil || token_checks.IsTokenEmpty(req) {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	studies, err := s.studyDBservice.GetStudiesByStatus(req.InstanceId, types.STUDY_STATUS_ACTIVE, false)
	if err != nil {
		logger.Info.Println(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &api.Studies{Studies: []*api.Study{}}
	for _, study := range studies {
		// at least one profile in the study:
		resp.Studies = append(resp.Studies, &api.Study{
			Key:    study.Key,
			Status: study.Status,
			Props:  study.Props.ToAPI(),
			Stats:  study.Stats.ToAPI(),
		})

	}
	return resp, nil
}

func (s *studyServiceServer) GetReportsForUser(ctx context.Context, req *api.GetReportsForUserReq) (*api.ReportHistory, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	profileIDs := []string{req.Token.ProfilId}
	profileIDs = append(profileIDs, req.Token.OtherProfileIds...)

	// filter profiles
	if len(req.OnlyForProfiles) > 0 {
		filtered_list := []string{}
		for _, pF := range req.OnlyForProfiles {
			for _, p := range profileIDs {
				if p == pF {
					filtered_list = append(filtered_list, p)
					break
				}
			}
		}
		profileIDs = filtered_list
	}

	resp := &api.ReportHistory{
		Reports: []*api.Report{},
	}
	studies, err := s.studyDBservice.GetStudiesByStatus(req.Token.InstanceId, types.STUDY_STATUS_ACTIVE, false)
	if err != nil {
		logger.Info.Println(err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	for _, study := range studies {
		// filter study keys
		if len(req.OnlyForStudies) > 0 {
			keyOk := false
			for _, fk := range req.OnlyForStudies {
				if fk == study.Key {
					keyOk = true
					break
				}
			}
			if !keyOk {
				continue
			}
		}

		for _, profileID := range profileIDs {
			// ParticipantID
			participantID, err := utils.ProfileIDtoParticipantID(profileID, s.StudyGlobalSecret, study.SecretKey, study.Configs.IdMappingMethod)
			if err != nil {
				logger.Error.Printf("couldn't compute participant ID: %v", err)
				continue
			}

			query := studydb.ReportQuery{
				ParticipantID: participantID,
				Key:           req.ReportKeyFilter,
				Since:         req.From,
				Until:         req.Until,
			}
			reports, err := s.studyDBservice.FindReports(req.Token.InstanceId, study.Key, query)
			if err != nil {
				logger.Debug.Printf("couldn't find reports: %v", err)
				continue
			}

			for _, r := range reports {
				ro := r.ToAPI()
				ro.StudyKey = study.Key
				ro.ProfileId = profileID
				ro.ParticipantId = "" // don't share internal participant id
				resp.Reports = append(resp.Reports, ro)
			}
		}
	}

	return resp, nil
}

func (s *studyServiceServer) HasParticipantStateWithCondition(ctx context.Context, req *api.ProfilesWithConditionReq) (*api.ServiceStatus, error) {
	if req == nil || req.StudyKey == "" || len(req.ProfileIds) < 1 {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	study, err := s.studyDBservice.GetStudyByStudyKey(req.InstanceId, req.StudyKey)
	if err != nil {
		return nil, status.Error(codes.Internal, "study not found")
	}

	for _, profileID := range req.ProfileIds {
		// ParticipantID
		participantID, err := utils.ProfileIDtoParticipantID(profileID, s.StudyGlobalSecret, study.SecretKey, study.Configs.IdMappingMethod)
		if err != nil {
			continue
		}

		pState, err := s.studyDBservice.FindParticipantState(req.InstanceId, study.Key, participantID)
		if err != nil {
			// profile not in the study
			continue
		}

		cond := types.ExpressionArgFromAPI(req.Condition)
		if cond == nil {
			return &api.ServiceStatus{
				Version: apiVersion,
				Status:  api.ServiceStatus_NORMAL,
				Msg:     "participant found in study",
			}, nil
		} else if cond.IsExpression() {
			evalCtx := studyengine.EvalContext{
				ParticipantState: pState,
				DbService:        s.studyDBservice,
				Event: types.StudyEvent{
					InstanceID: req.InstanceId,
					StudyKey:   study.Key,
				},
			}
			resp, err := studyengine.ExpressionEval(*cond.Exp, evalCtx)
			if err != nil {
				logger.Debug.Println(err)
				// profile not in the study
				continue
			}
			bVal, ok := resp.(bool)
			if ok && bVal {
				return &api.ServiceStatus{
					Version: apiVersion,
					Status:  api.ServiceStatus_NORMAL,
					Msg:     "participant found in study",
				}, nil
			}
		} else if cond.Num > 0 {
			// hardcoded true
			return &api.ServiceStatus{
				Version: apiVersion,
				Status:  api.ServiceStatus_NORMAL,
				Msg:     "participant found in study",
			}, nil
		}
	}
	return nil, status.Error(codes.NotFound, "no participant found")
}

func (s *studyServiceServer) GetParticipantMessages(ctx context.Context, req *api.GetParticipantMessagesReq) (*api.GetParticipantMessagesResp, error) {
	if req == nil || req.InstanceId == "" || req.StudyKey == "" || req.ProfileId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	participantID, _, err := s.profileIDToParticipantID(req.InstanceId, req.StudyKey, req.ProfileId, true)
	if err != nil {
		logger.Debug.Println(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	pState, err := s.studyDBservice.FindParticipantState(req.InstanceId, req.StudyKey, participantID)
	if err != nil {
		logger.Debug.Println(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &api.GetParticipantMessagesResp{
		Messages: []*api.GetParticipantMessagesResp_Messages{},
	}
	for _, message := range pState.Messages {
		if message.ScheduledFor > time.Now().Unix() {
			continue
		}
		resp.Messages = append(resp.Messages, &api.GetParticipantMessagesResp_Messages{
			Id:   message.ID,
			Type: message.Type,
		})
	}
	return resp, nil
}

func (s *studyServiceServer) DeleteMessagesFromParticipant(ctx context.Context, req *api.DeleteMessagesFromParticipantReq) (*api.ServiceStatus, error) {
	if req == nil || req.InstanceId == "" || req.StudyKey == "" || req.ProfileId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	participantID, _, err := s.profileIDToParticipantID(req.InstanceId, req.StudyKey, req.ProfileId, true)
	if err != nil {
		logger.Debug.Println(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, err = s.studyDBservice.FindParticipantState(req.InstanceId, req.StudyKey, participantID)
	if err != nil {
		logger.Debug.Println(err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	err = s.studyDBservice.DeleteMessagesFromParticipant(req.InstanceId, req.StudyKey, participantID, req.MessageIds)
	if err != nil {
		logger.Debug.Println(err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &api.ServiceStatus{
		Status: api.ServiceStatus_NORMAL,
		Msg:    "deleted",
	}, nil
}
