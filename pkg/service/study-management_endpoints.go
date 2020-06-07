package service

import (
	"context"

	"github.com/influenzanet/study-service/pkg/api"
	"github.com/influenzanet/study-service/pkg/types"
	"github.com/influenzanet/study-service/pkg/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *studyServiceServer) CreateNewStudy(ctx context.Context, req *api.NewStudyRequest) (*api.Study, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.Study == nil {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if !utils.CheckIfAnyRolesInToken(req.Token, []string{"RESEARCHER", "ADMIN"}) {
		return nil, status.Error(codes.Unauthenticated, "not authorized to create a study")
	}

	study := types.StudyFromAPI(req.Study)
	study.Members = []types.StudyMember{
		{
			Role:     "owner",
			UserID:   req.Token.Id,
			UserName: utils.GetUsernameFromToken(req.Token),
		},
	}

	cStudy, err := s.studyDBservice.CreateStudy(req.Token.InstanceId, study)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return cStudy.ToAPI(), nil
}

func (s *studyServiceServer) GetAllStudies(ctx context.Context, req *api.TokenInfos) (*api.Studies, error) {
	if req == nil || utils.IsTokenEmpty(req) {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if !utils.CheckIfAnyRolesInToken(req, []string{"RESEARCHER", "ADMIN"}) {
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
		if !utils.CheckIfMember(req.Id, study.Members, []string{"maintainer", "owner"}) {
			// don't share secret key if not study admin
			study.SecretKey = ""
		}
		resp.Studies = append(resp.Studies, study.ToAPI())
	}
	return resp, nil
}

func (s *studyServiceServer) GetStudy(ctx context.Context, req *api.StudyReferenceReq) (*api.Study, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if !utils.CheckIfAnyRolesInToken(req.Token, []string{"RESEARCHER", "ADMIN"}) {
		return nil, status.Error(codes.Unauthenticated, "not authorized")
	}

	study, err := s.studyDBservice.GetStudyByStudyKey(req.Token.InstanceId, req.StudyKey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !utils.CheckIfMember(req.Token.Id, study.Members, []string{"maintainer", "owner"}) {
		// don't share secret key if not study admin
		study.SecretKey = ""
	}

	return study.ToAPI(), nil
}

func (s *studyServiceServer) SaveSurveyToStudy(ctx context.Context, req *api.AddSurveyReq) (*api.Survey, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.StudyKey == "" || req.Survey == nil {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id, []string{"maintainer", "owner"})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	newSurvey := types.SurveyFromAPI(req.Survey)
	createdSurvey, err := s.studyDBservice.SaveSurvey(req.Token.InstanceId, req.StudyKey, newSurvey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return createdSurvey.ToAPI(), nil
}

func (s *studyServiceServer) GetSurveyDefForStudy(ctx context.Context, req *api.SurveyReferenceRequest) (*api.Survey, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.StudyKey == "" || req.SurveyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id, []string{"maintainer", "owner"})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	survey, err := s.studyDBservice.FindSurveyDef(req.Token.InstanceId, req.StudyKey, req.SurveyKey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return survey.ToAPI(), nil
}

func (s *studyServiceServer) RemoveSurveyFromStudy(ctx context.Context, req *api.SurveyReferenceRequest) (*api.ServiceStatus, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.StudyKey == "" || req.SurveyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id, []string{"maintainer", "owner"})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = s.studyDBservice.RemoveSurveyFromStudy(req.Token.InstanceId, req.StudyKey, req.SurveyKey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &api.ServiceStatus{
		Status:  api.ServiceStatus_NORMAL,
		Msg:     "survey removed",
		Version: apiVersion,
	}, nil
}

func (s *studyServiceServer) GetStudySurveyInfos(ctx context.Context, req *api.StudyReferenceReq) (*api.SurveyInfoResp, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	surveys, err := s.studyDBservice.FindAllSurveyDefsForStudy(req.Token.InstanceId, req.StudyKey, false)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	infos := make([]*api.SurveyInfoResp_SurveyInfo, len(surveys))
	for i, s := range surveys {
		apiS := s.ToAPI()
		infos[i] = &api.SurveyInfoResp_SurveyInfo{
			Key:         s.Current.SurveyDefinition.Key,
			Name:        apiS.Name,
			Description: apiS.Description,
		}
	}

	resp := api.SurveyInfoResp{
		Infos: infos,
	}
	return &resp, nil
}

func (s *studyServiceServer) SaveStudyMember(ctx context.Context, req *api.StudyMemberReq) (*api.Study, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.StudyKey == "" || req.Member == nil || req.Member.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	if !utils.CheckIfAnyRolesInToken(req.Token, []string{"ADMIN"}) {
		err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id, []string{"maintainer", "owner"})
		if err != nil {
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
	return uStudy.ToAPI(), nil
}

func (s *studyServiceServer) RemoveStudyMember(ctx context.Context, req *api.StudyMemberReq) (*api.Study, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.StudyKey == "" || req.Member == nil || req.Member.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id, []string{"maintainer", "owner"})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
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
	return uStudy.ToAPI(), nil
}

func (s *studyServiceServer) SaveStudyRules(ctx context.Context, req *api.StudyRulesReq) (*api.Study, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id, []string{"maintainer", "owner"})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (s *studyServiceServer) SaveStudyStatus(ctx context.Context, req *api.StudyStatusReq) (*api.Study, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id, []string{"maintainer", "owner"})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (s *studyServiceServer) SaveStudyProps(ctx context.Context, req *api.StudyPropsReq) (*api.Study, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id, []string{"maintainer", "owner"})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (s *studyServiceServer) DeleteStudy(ctx context.Context, req *api.StudyReferenceReq) (*api.ServiceStatus, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.StudyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	err := s.HasRoleInStudy(req.Token.InstanceId, req.StudyKey, req.Token.Id, []string{"maintainer", "owner"})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}
