package service

import (
	"context"
	"log"

	"github.com/influenzanet/go-utils/pkg/api_types"
	"github.com/influenzanet/go-utils/pkg/token_checks"
	"github.com/influenzanet/study-service/pkg/api"
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
		log.Printf("GetStudiesForUser.GetStudiesByStatus: %v", err)
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
			participantID, err := utils.ProfileIDtoParticipantID(profileID, s.StudyGlobalSecret, study.SecretKey)
			if err != nil {
				continue
			}

			pState, err := s.studyDBservice.FindParticipantState(req.Token.InstanceId, study.Key, participantID)
			if err != nil {
				// user not in the study
				continue
			}

			if pState.StudyStatus != "active" {
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

	studies, err := s.studyDBservice.GetStudiesByStatus(req.InstanceId, "active", false)
	if err != nil {
		log.Printf("GetActiveStudies.GetStudiesByStatus: %v", err)
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
		participantID, err := utils.ProfileIDtoParticipantID(profileID, s.StudyGlobalSecret, study.SecretKey)
		if err != nil {
			continue
		}

		pState, err := s.studyDBservice.FindParticipantState(req.InstanceId, study.Key, participantID)
		if err != nil {
			// profile not in the study
			continue
		}

		cond := types.ExpressionArgFromAPI(req.Condition)
		if cond.IsExpression() {
			evalCtx := studyengine.EvalContext{
				ParticipantState: pState,
			}
			resp, err := studyengine.ExpressionEval(*cond.Exp, evalCtx)
			if err != nil {
				log.Printf("HasParticipantStateWithCondition.ExpressionEval: %v", err)
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
			//log.Println(pState)
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
