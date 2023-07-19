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

	// create required indexes for the new study:
	err = s.studyDBservice.CreateSurveyDefintionIndexForStudy(req.Token.InstanceId, study.Key)
	if err != nil {
		logger.Error.Printf("unexpected error when creating survey definition indexes: %v", err)
	}

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
	if newSurvey.VersionID == "" {
		surveyHistory, err := s.studyDBservice.FindSurveyDefHistory(req.Token.InstanceId, req.StudyKey, req.Survey.SurveyDefinition.Key, true)
		if err != nil {
			errMsg := fmt.Sprintf("fetching survey history returned: %v", err)
			logger.Error.Println(errMsg)
			return nil, status.Error(codes.Internal, errMsg)
		}
		newSurvey.VersionID = utils.GenerateSurveyVersionID(surveyHistory)
	}

	newSurvey.Published = time.Now().Unix()
	createdSurvey, err := s.studyDBservice.SaveSurvey(req.Token.InstanceId, req.StudyKey, newSurvey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_LOG, constants.LOG_EVENT_SAVE_SURVEY, req.StudyKey)
	return createdSurvey.ToAPI(), nil
}

func (s *studyServiceServer) GetSurveyDefForStudy(ctx context.Context, req *api.SurveyVersionReferenceRequest) (*api.Survey, error) {
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

	var survey *types.Survey
	var err error
	if req.VersionId == "" {
		survey, err = s.studyDBservice.FindCurrentSurveyDef(req.Token.InstanceId, req.StudyKey, req.SurveyKey, false)
	} else {
		survey, err = s.studyDBservice.FindSurveyDefByVersionID(req.Token.InstanceId, req.StudyKey, req.SurveyKey, req.VersionId)
	}
	if err != nil {
		s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_LOG, constants.LOG_EVENT_GET_SURVEY_DEF, "not found"+req.StudyKey+"-"+req.SurveyKey)
		return nil, status.Error(codes.Internal, err.Error())
	}
	s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_LOG, constants.LOG_EVENT_GET_SURVEY_DEF, req.StudyKey+"-"+req.SurveyKey)
	return survey.ToAPI(), nil
}

func (s *studyServiceServer) RemoveSurveyVersion(ctx context.Context, req *api.SurveyVersionReferenceRequest) (*api.ServiceStatus, error) {
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

	err := s.studyDBservice.DeleteSurveyVersion(req.Token.InstanceId, req.StudyKey, req.SurveyKey, req.VersionId)
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

func (s *studyServiceServer) UnpublishSurvey(ctx context.Context, req *api.SurveyReferenceRequest) (*api.ServiceStatus, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" || req.SurveyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if !token_checks.CheckRoleInToken(req.Token, constants.USER_ROLE_ADMIN) {

		err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id, []string{
			types.STUDY_ROLE_MAINTAINER,
			types.STUDY_ROLE_OWNER,
		})
		if err != nil {
			s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_UNPUBLISH_SURVEY, fmt.Sprintf("permission denied for unpublishing %s from study %s  ", req.SurveyKey, req.StudyKey))
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	err := s.studyDBservice.UnpublishSurvey(req.Token.InstanceId, req.StudyKey, req.SurveyKey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_LOG, constants.LOG_EVENT_UNPUBLISH_SURVEY, fmt.Sprintf("unpublished %s from study %s  ", req.SurveyKey, req.StudyKey))
	return &api.ServiceStatus{
		Status:  api.ServiceStatus_NORMAL,
		Msg:     "survey unpublished",
		Version: apiVersion,
	}, nil
}

func (s *studyServiceServer) GetStudySurveyInfos(ctx context.Context, req *api.StudyReferenceReq) (*api.SurveyInfoResp, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	surveys, err := s.studyDBservice.FindAllCurrentSurveyDefsForStudy(req.Token.InstanceId, req.StudyKey, true)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	infos := make([]*api.SurveyInfo, len(surveys))
	for i, s := range surveys {
		apiS := s.ToAPI()
		infos[i] = &api.SurveyInfo{
			StudyKey:        req.StudyKey,
			SurveyKey:       s.SurveyDefinition.Key,
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

func (s *studyServiceServer) GetSurveyVersionInfos(ctx context.Context, req *api.SurveyReferenceRequest) (*api.SurveyVersions, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" || req.SurveyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	if !token_checks.CheckIfAnyRolesInToken(req.Token, []string{constants.USER_ROLE_ADMIN}) {
		err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id,
			[]string{types.STUDY_ROLE_MAINTAINER, types.STUDY_ROLE_OWNER},
		)
		if err != nil {
			logger.Error.Println(err)
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	surveys, err := s.studyDBservice.FindSurveyDefHistory(req.Token.InstanceId, req.StudyKey, req.SurveyKey, true)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	versions := make([]*api.Survey, len(surveys))
	for i, s := range surveys {
		apiS := s.ToAPI()
		versions[i] = apiS
	}

	resp := api.SurveyVersions{
		SurveyVersions: versions,
	}
	return &resp, nil
}

func (s *studyServiceServer) GetSurveyKeys(ctx context.Context, req *api.GetSurveyKeysRequest) (*api.SurveyKeys, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	if !token_checks.CheckIfAnyRolesInToken(req.Token, []string{constants.USER_ROLE_ADMIN}) {
		err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id,
			[]string{types.STUDY_ROLE_MAINTAINER, types.STUDY_ROLE_OWNER},
		)
		if err != nil {
			logger.Error.Println(err)
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	surveyKeys, err := s.studyDBservice.GetSurveyKeysInStudy(req.Token.InstanceId, req.StudyKey, req.IncludeUnpublished)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := api.SurveyKeys{
		Keys: surveyKeys,
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

func (s *studyServiceServer) GetCurrentStudyRules(ctx context.Context, req *api.StudyReferenceReq) (*api.StudyRules, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if !token_checks.CheckRoleInToken(req.Token, constants.USER_ROLE_ADMIN) {
		err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id,
			[]string{types.STUDY_ROLE_MAINTAINER, types.STUDY_ROLE_OWNER},
		)
		if err != nil {
			s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_STUDY_RULES, "permission denied for: "+req.StudyKey)
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	studyRules, err := s.studyDBservice.GetCurrentStudyRules(req.Token.InstanceId, req.StudyKey)
	if err != nil {
		logger.Warning.Printf("study rules for study %s not found", req.StudyKey)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return studyRules.ToAPI(), nil
}

func (s *studyServiceServer) GetStudyRulesHistory(ctx context.Context, req *api.StudyRulesHistoryReq) (*api.StudyRulesHistory, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if !token_checks.CheckRoleInToken(req.Token, constants.USER_ROLE_ADMIN) {
		err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id,
			[]string{types.STUDY_ROLE_MAINTAINER, types.STUDY_ROLE_OWNER},
		)
		if err != nil {
			s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_STUDY_RULES, "permission denied for: "+req.StudyKey)
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	studyRules, itemCount, err := s.studyDBservice.GetStudyRulesHistory(req.Token.InstanceId, req.StudyKey, req.PageSize, req.Page, req.Descending, req.Since, req.Until)
	if err != nil {
		logger.Warning.Printf("no study rules found for study %s", req.StudyKey)
		return nil, status.Error(codes.Internal, err.Error())
	}
	versions := make([]*api.StudyRules, len(studyRules))
	for i, s := range studyRules {
		versions[i] = s.ToAPI()
	}

	pageCount := int32(1)
	pageSize := itemCount
	page := req.Page
	if utils.CheckForValidPaginationParameter(req.PageSize, req.Page) {
		pageCount = utils.ComputePageCount(req.PageSize, itemCount)
		pageSize = req.PageSize
		if page > pageCount {
			if pageCount > 0 {
				page = pageCount
			} else {
				page = 1
			}
		}
	} else {
		page = 1
	}

	resp := &api.StudyRulesHistory{
		Rules:     versions,
		PageCount: pageCount,
		ItemCount: itemCount,
		Page:      page,
		PageSize:  pageSize,
	}
	if utils.CheckForValidPaginationParameter(req.PageSize, req.Page) {
		logger.Debug.Printf("received %d study rules objects for query, page %d out of %d pages is displayed with %d items per page", resp.ItemCount, resp.Page, resp.PageCount, resp.PageSize)
	} else {
		logger.Debug.Printf("received %d study rules objects for query", resp.ItemCount)
	}
	return resp, nil
}

func (s *studyServiceServer) RemoveStudyRulesVersion(ctx context.Context, req *api.StudyRulesVersionReferenceReq) (*api.ServiceStatus, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	studyKey, err := s.studyDBservice.GetStudyKeyByStudyRulesID(req.Token.InstanceId, req.Id)
	if err != nil {
		logger.Error.Printf(err.Error())
		return nil, status.Error(codes.Internal, "deletion failed")
	}

	if !token_checks.CheckRoleInToken(req.Token, constants.USER_ROLE_ADMIN) {
		err := s.HasRoleInStudy(req.Token.InstanceId, studyKey, req.Token.Id,
			[]string{types.STUDY_ROLE_MAINTAINER, types.STUDY_ROLE_OWNER},
		)
		if err != nil {
			s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_STUDY_RULES, "permission denied for: "+studyKey)
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	//delete all study rules for study
	err = s.studyDBservice.DeleteStudyRulesVersion(req.Token.InstanceId, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_LOG, constants.LOG_EVENT_STUDY_DELETION, studyKey)
	return &api.ServiceStatus{
		Status: api.ServiceStatus_NORMAL,
		Msg:    "study rules deleted",
	}, nil
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

	studyRules := types.StudyRules{
		StudyKey:   req.StudyKey,
		UploadedAt: time.Now().Unix(),
		UploadedBy: req.Token.ProfilId,
		Rules:      rules,
	}
	_, err = s.studyDBservice.AddStudyRules(req.Token.InstanceId, studyRules)
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

func (s *studyServiceServer) GetResearcherNotificationSubscriptions(ctx context.Context, req *api.GetResearcherNotificationSubscriptionsReq) (*api.NotificationSubscriptions, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if !token_checks.CheckRoleInToken(req.Token, constants.USER_ROLE_ADMIN) {
		err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id,
			[]string{types.STUDY_ROLE_MAINTAINER, types.STUDY_ROLE_OWNER},
		)
		if err != nil {
			s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_STUDY_UPDATE, "permission denied for fetching notification subscriptions for "+req.StudyKey)
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	study, err := s.studyDBservice.GetStudyByStudyKey(req.Token.InstanceId, req.StudyKey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	subscriptions := &api.NotificationSubscriptions{
		Subscriptions: []*api.Subscription{},
	}

	for _, sub := range study.NotificationSubscriptions {
		subscriptions.Subscriptions = append(subscriptions.Subscriptions, sub.ToAPI())
	}
	logger.Info.Printf("(instance: %s): %s fetched notification subscriptions for study %s", req.Token.InstanceId, req.Token.Id, req.StudyKey)
	return subscriptions, nil
}

func (s *studyServiceServer) UpdateResearcherNotificationSubscriptions(ctx context.Context, req *api.UpdateResearcherNotificationSubscriptionsReq) (*api.NotificationSubscriptions, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if !token_checks.CheckRoleInToken(req.Token, constants.USER_ROLE_ADMIN) {
		err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id,
			[]string{types.STUDY_ROLE_MAINTAINER, types.STUDY_ROLE_OWNER},
		)
		if err != nil {
			s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_STUDY_UPDATE, "permission denied for updating notification subscriptions for "+req.StudyKey)
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	study, err := s.studyDBservice.GetStudyByStudyKey(req.Token.InstanceId, req.StudyKey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	subs := []types.NotificationSubscription{}
	for _, sub := range req.Subscriptions {
		subs = append(subs, types.NotificationSubscriptionFromAPI(sub))
	}

	study.NotificationSubscriptions = subs

	uStudy, err := s.studyDBservice.UpdateStudyInfo(req.Token.InstanceId, study)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	subscriptions := &api.NotificationSubscriptions{
		Subscriptions: []*api.Subscription{},
	}

	for _, sub := range uStudy.NotificationSubscriptions {
		subscriptions.Subscriptions = append(subscriptions.Subscriptions, sub.ToAPI())
	}
	logger.Info.Printf("(instance: %s): %s updated notification subscriptions for study %s", req.Token.InstanceId, req.Token.Id, req.StudyKey)
	s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_LOG, constants.LOG_EVENT_STUDY_UPDATE, req.StudyKey)
	return subscriptions, nil
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

	// Convert rules from API type:
	rules := make([]*types.Expression, len(req.Rules))
	for index, rule := range req.Rules {
		rules[index] = types.ExpressionFromAPI(rule)
	}

	err := s.studyDBservice.FindAndExecuteOnParticipantsStates(
		ctx,
		req.Token.InstanceId,
		req.StudyKey,
		"",
		func(dbService *studydb.StudyDBService, p types.ParticipantState, instanceID, studyKey string, args ...interface{}) error {
			if p.StudyStatus == types.PARTICIPANT_STUDY_STATUS_TEMPORARY {
				// ignore temporary participants
				return nil
			}

			counters.Participants += 1

			participantID2, _, err := s.profileIDToParticipantID(instanceID, studyKey, p.ParticipantID, true)
			if err != nil {
				logger.Error.Printf("RunRules: %v", err)
				return status.Error(codes.Internal, err.Error())
			}

			actionData := studyengine.ActionData{
				PState:          p,
				ReportsToCreate: map[string]types.Report{},
			}
			anyChange := false
			for index, rule := range rules {
				if rule == nil {
					continue
				}

				event := types.StudyEvent{
					InstanceID:                            instanceID,
					StudyKey:                              studyKey,
					ParticipantIDForConfidentialResponses: participantID2,
				}
				newState, err := studyengine.ActionEval(*rule, actionData, event, studyengine.ActionConfigs{
					DBService:              s.studyDBservice,
					ExternalServiceConfigs: s.studyEngineExternalServices,
				})
				if err != nil {
					return err
				}

				if !reflect.DeepEqual(newState.PState, actionData.PState) {
					counters.ParticipantStateChangePerRule[index] += 1
					anyChange = true
				}
				actionData = newState
			}

			if anyChange {
				// save state back to DB
				_, err := s.studyDBservice.SaveParticipantState(instanceID, studyKey, actionData.PState)
				if err != nil {
					logger.Error.Printf("RunRules: %v", err)
					return status.Error(codes.Internal, err.Error())
				}

			}
			s.saveReports(instanceID, req.StudyKey, actionData.ReportsToCreate, "")
			return nil
		},
	)
	if err != nil {
		logger.Error.Println(err)
	}

	s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_LOG, constants.LOG_EVENT_RUN_CUSTOM_RULES, fmt.Sprintf("rules run for study %s: %v", req.StudyKey, req.Rules))
	resp := api.RuleRunSummary{
		ParticipantCount:              counters.Participants,
		ParticipantStateChangePerRule: counters.ParticipantStateChangePerRule,
		Duration:                      time.Now().Unix() - start,
	}
	return &resp, nil
}

func (s *studyServiceServer) RunRulesForSingleParticipant(ctx context.Context, req *api.RunRulesForSingleParticipantReq) (*api.RuleRunSummary, error) {
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

	// Convert rules from API type:
	rules := make([]*types.Expression, len(req.Rules))
	for index, rule := range req.Rules {
		rules[index] = types.ExpressionFromAPI(rule)
	}

	p, err := s.studyDBservice.FindParticipantState(req.Token.InstanceId, req.StudyKey, req.ParticipantId)
	if err != nil {
		logger.Debug.Printf("participant not found: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	if p.StudyStatus == types.PARTICIPANT_STUDY_STATUS_TEMPORARY {
		// ignore temporary participants
		resp := api.RuleRunSummary{
			ParticipantCount:              counters.Participants,
			ParticipantStateChangePerRule: counters.ParticipantStateChangePerRule,
			Duration:                      time.Now().Unix() - start,
		}
		return &resp, nil
	}

	participantID2, _, err := s.profileIDToParticipantID(req.Token.InstanceId, req.StudyKey, p.ParticipantID, true)
	if err != nil {
		logger.Debug.Printf("unexpected error: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	counters.Participants += 1
	actionData := studyengine.ActionData{
		PState:          p,
		ReportsToCreate: map[string]types.Report{},
	}
	anyChange := false
	for index, rule := range rules {
		if rule == nil {
			continue
		}

		event := types.StudyEvent{
			InstanceID:                            req.Token.InstanceId,
			StudyKey:                              req.StudyKey,
			ParticipantIDForConfidentialResponses: participantID2,
		}
		newState, err := studyengine.ActionEval(*rule, actionData, event, studyengine.ActionConfigs{
			DBService:              s.studyDBservice,
			ExternalServiceConfigs: s.studyEngineExternalServices,
		})
		if err != nil {
			logger.Debug.Printf("unexpected error: %v", err)
			return nil, status.Error(codes.Internal, err.Error())
		}

		if !reflect.DeepEqual(newState.PState, actionData.PState) {
			counters.ParticipantStateChangePerRule[index] += 1
			anyChange = true
		}
		actionData = newState
	}

	if anyChange {
		// save state back to DB
		_, err := s.studyDBservice.SaveParticipantState(req.Token.InstanceId, req.StudyKey, actionData.PState)
		if err != nil {
			logger.Debug.Printf("unexpected error: %v", err)
			return nil, status.Error(codes.Internal, err.Error())
		}

	}
	s.saveReports(req.Token.InstanceId, req.StudyKey, actionData.ReportsToCreate, "")

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

	//delete all study rules for study
	err := s.studyDBservice.DeleteStudyRulesByStudyKey(req.Token.InstanceId, req.StudyKey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	err = s.studyDBservice.DeleteStudy(req.Token.InstanceId, req.StudyKey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_LOG, constants.LOG_EVENT_STUDY_DELETION, req.StudyKey)
	return &api.ServiceStatus{
		Status: api.ServiceStatus_NORMAL,
		Msg:    "study deleted",
	}, nil
}

func (s *studyServiceServer) GetParticipantStateByID(ctx context.Context, req *api.ParticipantStateByIDQuery) (*api.ParticipantState, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if !token_checks.CheckRoleInToken(req.Token, constants.USER_ROLE_ADMIN) {
		err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id,
			[]string{types.STUDY_ROLE_MAINTAINER, types.STUDY_ROLE_OWNER},
		)
		if err != nil {
			s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_GET_PARTICIPANT_STATES, "permission denied for: "+req.StudyKey)
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	participantState, err := s.studyDBservice.FindParticipantState(req.Token.InstanceId, req.StudyKey, req.ParticipantId)
	if err != nil {
		logger.Warning.Printf("participant with ID %s not found", req.ParticipantId)
		return nil, status.Error(codes.Internal, err.Error())
	}

	logger.Debug.Printf("found participant with ID %s", req.ParticipantId)
	return participantState.ToAPI(), nil
}

func (s *studyServiceServer) GetParticipantStatesWithPagination(ctx context.Context, req *api.GetPStatesWithPaginationQuery) (*api.ParticipantStatesWithPagination, error) {
	if req == nil || token_checks.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if !token_checks.CheckRoleInToken(req.Token, constants.USER_ROLE_ADMIN) {
		err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id,
			[]string{types.STUDY_ROLE_MAINTAINER, types.STUDY_ROLE_OWNER},
		)
		if err != nil {
			s.SaveLogEvent(req.Token.InstanceId, req.Token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_GET_PARTICIPANT_STATES, "permission denied for: "+req.StudyKey)
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	participantStates, itemCount, err := s.studyDBservice.FindParticipantsByQuery(req.Token.InstanceId, req.StudyKey, req.Query, req.SortBy, req.PageSize, req.Page)
	if err != nil {
		logger.Error.Printf("error while fetching participant states")
		return nil, status.Error(codes.Internal, err.Error())
	}

	ps := []*api.ParticipantState{}
	for _, participantState := range participantStates {
		state := participantState.ToAPI()
		ps = append(ps, state)
	}

	pageCount := int32(1)
	pageSize := itemCount
	page := req.Page
	if utils.CheckForValidPaginationParameter(req.PageSize, req.Page) {
		pageCount = utils.ComputePageCount(req.PageSize, itemCount)
		pageSize = req.PageSize
		if page > pageCount {
			if pageCount > 0 {
				page = pageCount
			} else {
				page = 1
			}
		}
	} else {
		page = 1
	}
	if itemCount == 0 {
		pageCount = 0
	}
	resp := &api.ParticipantStatesWithPagination{
		ItemCount: itemCount,
		PageCount: pageCount,
		Page:      page,
		Items:     ps,
		PageSize:  pageSize,
	}
	if utils.CheckForValidPaginationParameter(req.PageSize, req.Page) {
		logger.Debug.Printf("received %d participant states for query, page %d out of %d pages is displayed with %d items per page", resp.ItemCount, resp.Page, resp.PageCount, resp.PageSize)
	} else {
		logger.Debug.Printf("received %d participant states for query", resp.ItemCount)
	}
	return resp, nil
}
