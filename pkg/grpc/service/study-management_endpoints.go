package service

import (
	"context"
	"fmt"
	"reflect"
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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *studyServiceServer) CreateNewStudy(ctx context.Context, req *api.NewStudyRequest) (*api.Study, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.Study == nil {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if !token_checks.CheckIfAnyRolesInToken(req.Token, []string{constants.USER_ROLE_RESEARCHER, constants.USER_ROLE_ADMIN}) {
		s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_STUDY_CREATION, "permission denied for "+req.Study.Key)
		return nil, status.Error(codes.Unauthenticated, "not authorized to create a study")
	}

	study := types.StudyFromAPI(req.Study)
	study.Members = []types.StudyMember{
		{
			Role:     types.STUDY_ROLE_OWNER,
			UserID:   req.Token.Id,
			UserName: token_checks.GetUsernameFromToken(req.Token),
		},
	}

	cStudy, err := s.studyDBservice.CreateStudy(req.Token.InstanceId, study)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_LOG, constants.LOG_EVENT_STUDY_CREATION, req.Study.Key)
	return cStudy.ToAPI(), nil
}

func (s *studyServiceServer) GetAllStudies(ctx context.Context, req *api_types.TokenInfos) (*api.Studies, error) {
	if req == nil || token_checks.IsTokenEmpty(req) {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if !token_checks.CheckIfAnyRolesInToken(req, []string{
		constants.USER_ROLE_ADMIN,
		constants.USER_ROLE_RESEARCHER,
	}) {
		s.SaveLogEvent(req.InstanceId, req.Id, loggingAPI.LogEventType_LOG, constants.LOG_EVENT_GET_STUDY, "permission denied for all studies")
		return nil, status.Error(codes.Unauthenticated, "not authorized")
	}

	studies, err := s.studyDBservice.GetStudiesByStatus(req.InstanceId, "", false)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &api.Studies{
		Studies: []*api.Study{},
	}
	for _, study := range studies {
		if !utils.CheckIfMember(req.Id, study.Members, []string{
			types.STUDY_ROLE_MAINTAINER,
			types.STUDY_ROLE_OWNER,
		}) {
			// don't share secret key if not study admin
			study.SecretKey = ""
		}
		resp.Studies = append(resp.Studies, study.ToAPI())
	}
	s.SaveLogEvent(req.InstanceId, req.Id, loggingAPI.LogEventType_LOG, constants.LOG_EVENT_GET_STUDY, "all")
	return resp, nil
}

func (s *studyServiceServer) GetStudy(ctx context.Context, req *api.StudyReferenceReq) (*api.Study, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if !token_checks.CheckIfAnyRolesInToken(req.Token, []string{
		constants.USER_ROLE_RESEARCHER,
		constants.USER_ROLE_ADMIN,
	}) {
		s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_GET_STUDY, "permission denied for: "+req.StudyKey)
		return nil, status.Error(codes.Unauthenticated, "not authorized")
	}

	study, err := s.studyDBservice.GetStudyByStudyKey(req.Token.InstanceId, req.StudyKey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !token_checks.CheckRoleInToken(req.Token, constants.USER_ROLE_ADMIN) &&
		!utils.CheckIfMember(req.Token.Id, study.Members, []string{
			types.STUDY_ROLE_MAINTAINER,
			types.STUDY_ROLE_OWNER,
		}) {
		// don't share secret key if not study admin
		study.SecretKey = ""
	}
	s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_LOG, constants.LOG_EVENT_GET_STUDY, req.StudyKey)
	return study.ToAPI(), nil
}

func (s *studyServiceServer) SaveSurveyToStudy(ctx context.Context, req *api.AddSurveyReq) (*api.Survey, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" || req.Survey == nil {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if !token_checks.CheckRoleInToken(req.Token, constants.USER_ROLE_ADMIN) {
		err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id, []string{
			types.STUDY_ROLE_MAINTAINER,
			types.STUDY_ROLE_OWNER,
		})
		if err != nil {
			s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_SAVE_SURVEY, fmt.Sprintf("permission denied for %s in %s", req.StudyKey, req.StudyKey))
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	newSurvey := types.SurveyFromAPI(req.Survey)
	if newSurvey.Current.VersionID == "" {
		newSurvey.Current.VersionID = utils.GenerateSurveyVersionID(newSurvey.History)
	}
	createdSurvey, err := s.studyDBservice.SaveSurvey(req.Token.InstanceId, req.StudyKey, newSurvey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_LOG, constants.LOG_EVENT_SAVE_SURVEY, req.StudyKey)
	return createdSurvey.ToAPI(), nil
}

func (s *studyServiceServer) GetSurveyDefForStudy(ctx context.Context, req *api.SurveyReferenceRequest) (*api.Survey, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" || req.SurveyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if !token_checks.CheckRoleInToken(req.Token, constants.USER_ROLE_ADMIN) {
		err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id, []string{
			types.STUDY_ROLE_MAINTAINER,
			types.STUDY_ROLE_OWNER,
		})
		if err != nil {
			s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_GET_SURVEY_DEF, "permission denied for "+req.StudyKey)
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	survey, err := s.studyDBservice.FindSurveyDef(req.Token.InstanceId, req.StudyKey, req.SurveyKey)
	if err != nil {
		s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_LOG, constants.LOG_EVENT_GET_SURVEY_DEF, "not found"+req.StudyKey+"-"+req.SurveyKey)
		return nil, status.Error(codes.Internal, err.Error())
	}
	s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_LOG, constants.LOG_EVENT_GET_SURVEY_DEF, req.StudyKey+"-"+req.SurveyKey)
	return survey.ToAPI(), nil
}

func (s *studyServiceServer) RemoveSurveyFromStudy(ctx context.Context, req *api.SurveyReferenceRequest) (*api.ServiceStatus, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" || req.SurveyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if !token_checks.CheckRoleInToken(req.Token, constants.USER_ROLE_ADMIN) {

		err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id, []string{
			types.STUDY_ROLE_MAINTAINER,
			types.STUDY_ROLE_OWNER,
		})
		if err != nil {
			s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_REMOVE_SURVEY, fmt.Sprintf("permission denied for removing %s from study %s  ", req.SurveyKey, req.StudyKey))
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	err := s.studyDBservice.RemoveSurveyFromStudy(req.Token.InstanceId, req.StudyKey, req.SurveyKey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_LOG, constants.LOG_EVENT_REMOVE_SURVEY, fmt.Sprintf("removed %s from study %s  ", req.SurveyKey, req.StudyKey))
	return &api.ServiceStatus{
		Status:  api.ServiceStatus_NORMAL,
		Msg:     "survey removed",
		Version: apiVersion,
	}, nil
}

func (s *studyServiceServer) GetStudySurveyInfos(ctx context.Context, req *api.StudyReferenceReq) (*api.SurveyInfoResp, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	surveys, err := s.studyDBservice.FindAllSurveyDefsForStudy(req.Token.InstanceId, req.StudyKey, false)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	infos := make([]*api.SurveyInfo, len(surveys))
	for i, s := range surveys {
		apiS := s.ToAPI()
		infos[i] = &api.SurveyInfo{
			StudyKey:        req.StudyKey,
			SurveyKey:       s.Current.SurveyDefinition.Key,
			Name:            apiS.Props.Name,
			Description:     apiS.Props.Description,
			TypicalDuration: apiS.Props.TypicalDuration,
		}
	}

	resp := api.SurveyInfoResp{
		Infos: infos,
	}
	return &resp, nil
}

type StudyRole string

func (s *studyServiceServer) SaveStudyMember(ctx context.Context, req *api.StudyMemberReq) (*api.Study, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" || req.Member == nil || req.Member.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if !token_checks.CheckIfAnyRolesInToken(req.Token, []string{constants.USER_ROLE_ADMIN}) {
		err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id,
			[]string{types.STUDY_ROLE_MAINTAINER, types.STUDY_ROLE_OWNER},
		)
		if err != nil {
			s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_STUDY_MEMBER, fmt.Sprintf("permission denied for saving member %s from study %s  ", req.Member.UserId, req.StudyKey))
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	study, err := s.studyDBservice.GetStudyByStudyKey(req.Token.InstanceId, req.StudyKey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	existing := false
	for i, m := range study.Members {
		if m.UserID == req.Member.UserId {
			study.Members[i] = types.StudyMemberFromAPI(req.Member)
			existing = true
			break
		}
	}
	if !existing {
		study.Members = append(study.Members, types.StudyMemberFromAPI(req.Member))
	}

	uStudy, err := s.studyDBservice.UpdateStudyInfo(req.Token.InstanceId, study)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_LOG, constants.LOG_EVENT_STUDY_MEMBER, fmt.Sprintf("save membmer (%s) for %s with role %s", req.Member.UserId, req.StudyKey, req.Member.Role))
	return uStudy.ToAPI(), nil
}

func (s *studyServiceServer) RemoveStudyMember(ctx context.Context, req *api.StudyMemberReq) (*api.Study, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" || req.Member == nil || req.Member.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if !token_checks.CheckRoleInToken(req.Token, constants.USER_ROLE_ADMIN) {
		err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id,
			[]string{types.STUDY_ROLE_MAINTAINER, types.STUDY_ROLE_OWNER},
		)
		if err != nil {
			s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_STUDY_MEMBER, fmt.Sprintf("permission denied for removing %s from study %s  ", req.Member.UserId, req.StudyKey))
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	study, err := s.studyDBservice.GetStudyByStudyKey(req.Token.InstanceId, req.StudyKey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	for i, m := range study.Members {
		if m.UserID == req.Member.UserId {
			study.Members = append(study.Members[:i], study.Members[i+1:]...)
			break
		}
	}

	uStudy, err := s.studyDBservice.UpdateStudyInfo(req.Token.InstanceId, study)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_LOG, constants.LOG_EVENT_STUDY_MEMBER, fmt.Sprintf("removed %s from study %s  ", req.Member.UserId, req.StudyKey))
	return uStudy.ToAPI(), nil
}

func (s *studyServiceServer) SaveStudyRules(ctx context.Context, req *api.StudyRulesReq) (*api.Study, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if !token_checks.CheckRoleInToken(req.Token, constants.USER_ROLE_ADMIN) {
		err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id,
			[]string{types.STUDY_ROLE_MAINTAINER, types.STUDY_ROLE_OWNER},
		)
		if err != nil {
			s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_STUDY_UPDATE, fmt.Sprintf("permission denied for rule update in study %s  ", req.StudyKey))
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	study, err := s.studyDBservice.GetStudyByStudyKey(req.Token.InstanceId, req.StudyKey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	rules := []types.Expression{}
	for _, exp := range req.Rules {
		rules = append(rules, *types.ExpressionFromAPI(exp))
	}
	study.Rules = rules

	uStudy, err := s.studyDBservice.UpdateStudyInfo(req.Token.InstanceId, study)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_LOG, constants.LOG_EVENT_STUDY_UPDATE, fmt.Sprintf("rules updated for %s", req.StudyKey))
	return uStudy.ToAPI(), nil
}

func (s *studyServiceServer) SaveStudyStatus(ctx context.Context, req *api.StudyStatusReq) (*api.Study, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if !token_checks.CheckRoleInToken(req.Token, constants.USER_ROLE_ADMIN) {
		err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id,
			[]string{types.STUDY_ROLE_MAINTAINER, types.STUDY_ROLE_OWNER},
		)
		if err != nil {
			s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_STUDY_UPDATE, fmt.Sprintf("permission denied for status update of study %s  ", req.StudyKey))
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	err := s.studyDBservice.UpdateStudyStatus(req.Token.InstanceId, req.StudyKey, req.NewStatus)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	study, err := s.studyDBservice.GetStudyByStudyKey(req.Token.InstanceId, req.StudyKey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_LOG, constants.LOG_EVENT_STUDY_UPDATE, fmt.Sprintf("status updated for %s", req.StudyKey))
	return study.ToAPI(), nil
}

func (s *studyServiceServer) SaveStudyProps(ctx context.Context, req *api.StudyPropsReq) (*api.Study, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if !token_checks.CheckRoleInToken(req.Token, constants.USER_ROLE_ADMIN) {
		err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id,
			[]string{types.STUDY_ROLE_MAINTAINER, types.STUDY_ROLE_OWNER},
		)
		if err != nil {
			s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_STUDY_UPDATE, "permission denied for updating props for "+req.StudyKey)
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	study, err := s.studyDBservice.GetStudyByStudyKey(req.Token.InstanceId, req.StudyKey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	study.Props = types.StudyPropsFromAPI(req.Props)

	uStudy, err := s.studyDBservice.UpdateStudyInfo(req.Token.InstanceId, study)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_LOG, constants.LOG_EVENT_STUDY_UPDATE, req.StudyKey)
	return uStudy.ToAPI(), nil
}

func (s *studyServiceServer) RunRules(ctx context.Context, req *api.StudyRulesReq) (*api.RuleRunSummary, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if !token_checks.CheckRoleInToken(req.Token, constants.USER_ROLE_ADMIN) {
		err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id,
			[]string{types.STUDY_ROLE_MAINTAINER, types.STUDY_ROLE_OWNER},
		)
		if err != nil {
			s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_RUN_CUSTOM_RULES, fmt.Sprintf("permission denied for running custom rules in study %s  ", req.StudyKey))
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	start := time.Now().Unix()

	type Counters struct {
		Participants                  int32
		ParticipantStateChangePerRule []int32
	}
	counters := &Counters{
		Participants:                  0,
		ParticipantStateChangePerRule: make([]int32, len(req.Rules)),
	}

	s.studyDBservice.FindAndExecuteOnParticipantsStates(
		req.Token.InstanceId,
		req.StudyKey,
		"",
		func(dbService *studydb.StudyDBService, p types.ParticipantState, instanceID, studyKey string, args ...interface{}) error {
			counters.Participants += 1

			currentState := p
			anyChange := false
			for index, rule := range req.Rules {
				exp := types.ExpressionFromAPI(rule)
				if exp == nil {
					continue
				}

				event := types.StudyEvent{
					InstanceID: instanceID,
					StudyKey:   studyKey,
				}
				newState, err := studyengine.ActionEval(*exp, currentState, event, s.studyDBservice)
				if err != nil {
					return err
				}

				if !reflect.DeepEqual(newState, currentState) {
					counters.ParticipantStateChangePerRule[index] += 1
					anyChange = true
				}
				currentState = newState
			}

			if anyChange {
				// save state back to DB
				_, err := s.studyDBservice.SaveParticipantState(instanceID, studyKey, currentState)
				if err != nil {
					logger.Error.Printf("RunRules: %v", err)
					return status.Error(codes.Internal, err.Error())
				}

			}

			return nil
		},
	)

	s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_LOG, constants.LOG_EVENT_RUN_CUSTOM_RULES, fmt.Sprintf("rules run for study %s: %v", req.StudyKey, req.Rules))
	resp := api.RuleRunSummary{
		ParticipantCount:              counters.Participants,
		ParticipantStateChangePerRule: counters.ParticipantStateChangePerRule,
		Duration:                      time.Now().Unix() - start,
	}
	return &resp, nil
}

func (s *studyServiceServer) DeleteStudy(ctx context.Context, req *api.StudyReferenceReq) (*api.ServiceStatus, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if !token_checks.CheckRoleInToken(req.Token, constants.USER_ROLE_ADMIN) {
		err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id,
			[]string{types.STUDY_ROLE_MAINTAINER, types.STUDY_ROLE_OWNER},
		)
		if err != nil {
			s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_STUDY_DELETION, "permission denied for: "+req.StudyKey)
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	err := s.studyDBservice.DeleteStudy(req.Token.InstanceId, req.StudyKey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_LOG, constants.LOG_EVENT_STUDY_DELETION, req.StudyKey)
	return &api.ServiceStatus{
		Status: api.ServiceStatus_NORMAL,
		Msg:    "study deleted",
	}, nil
}
