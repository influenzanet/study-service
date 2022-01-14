package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/coneno/logger"
	"github.com/influenzanet/go-utils/pkg/api_types"
	"github.com/influenzanet/go-utils/pkg/constants"
	loggingAPI "github.com/influenzanet/logging-service/pkg/api"
	"github.com/influenzanet/study-service/pkg/api"
	"github.com/influenzanet/study-service/pkg/dbs/studydb"
	"github.com/influenzanet/study-service/pkg/studyengine"
	"github.com/influenzanet/study-service/pkg/types"
	"github.com/influenzanet/study-service/pkg/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *studyServiceServer) profileIDToParticipantID(instanceID string, studyKey string, userID string) (string, error) {
	idMappingMethod, studySecret, err := s.studyDBservice.GetStudySecretKey(instanceID, studyKey)
	if err != nil {
		return "", err
	}
	return utils.ProfileIDtoParticipantID(userID, s.StudyGlobalSecret, studySecret, idMappingMethod)
}

func (s *studyServiceServer) checkIfParticipantExists(instanceID string, studyKey string, participantID string, withStatus string) bool {
	pState, err := s.studyDBservice.FindParticipantState(instanceID, studyKey, participantID)
	if err != nil || (withStatus != "" && pState.StudyStatus != withStatus) {
		return false
	}
	return err == nil
}

func (s *studyServiceServer) getAndPerformStudyRules(instanceID string, studyKey string, pState types.ParticipantState, event types.StudyEvent) (newState types.ParticipantState, err error) {
	newState = pState
	rules, err := s.studyDBservice.GetStudyRules(instanceID, studyKey)
	if err != nil {
		return
	}
	for _, rule := range rules {
		newState, err = studyengine.ActionEval(rule, newState, event, s.studyDBservice)
		if err != nil {
			return
		}
	}

	return newState, nil
}

func (s *studyServiceServer) resolveContextRules(instanceID string, studyKey string, participantID string, rules *types.SurveyContextDef) (sCtx types.SurveyContext, err error) {
	pState, err := s.studyDBservice.FindParticipantState(instanceID, studyKey, participantID)
	if err != nil {
		return sCtx, errors.New("no participant with this id in this study")
	}
	// participant flags:
	sCtx.ParticipantFlags = pState.Flags

	if rules == nil {
		return sCtx, nil
	}

	// mode:
	if rules.Mode != nil {
		modeRule := rules.Mode
		switch modeRule.DType {
		case "exp":
			return sCtx, errors.New("expression arg type not supported yet")
		case "str":
			sCtx.Mode = modeRule.Str
		default:
			sCtx.Mode = modeRule.Str
		}
	}

	// previous responses:
	prevRespRules := rules.PreviousResponses
	for _, rule := range prevRespRules {
		switch rule.Name {
		case "LAST_RESPONSES_BY_KEY":
			if len(rule.Data) != 2 {
				return sCtx, errors.New("LAST_RESPONSES_BY_KEY must have two arguments")
			}
			arg1 := rule.Data[0].Str
			arg2 := int64(rule.Data[1].Num)
			if arg1 == "" || arg2 == 0 {
				return sCtx, errors.New("LAST_RESPONSES_BY_KEY arguments have to be defined")
			}
			cResps, _ := s.studyDBservice.FindSurveyResponses(instanceID, studyKey, studydb.ResponseQuery{
				ParticipantID: participantID,
				SurveyKey:     arg1,
				Limit:         arg2,
			})
			sCtx.PreviousResponses = append(sCtx.PreviousResponses, cResps...)
		case "ALL_RESPONSES_SINCE":
			if len(rule.Data) != 1 {
				return sCtx, errors.New("ALL_RESPONSES_SINCE must have one argument")
			}
			arg1 := int64(rule.Data[0].Num)
			cResps, _ := s.studyDBservice.FindSurveyResponses(instanceID, studyKey, studydb.ResponseQuery{
				ParticipantID: participantID,
				Since:         arg1,
			})
			sCtx.PreviousResponses = append(sCtx.PreviousResponses, cResps...)
		case "RESPONSES_SINCE_BY_KEY":
			if len(rule.Data) != 2 {
				return sCtx, errors.New("RESPONSES_SINCE_BY_KEY must have two arguments")
			}
			since := int64(rule.Data[0].Num)
			surveyKey := rule.Data[1].Str
			if surveyKey == "" || since == 0 {
				return sCtx, errors.New("RESPONSES_SINCE_BY_KEY arguments have to be defined")
			}
			cResps, _ := s.studyDBservice.FindSurveyResponses(instanceID, studyKey, studydb.ResponseQuery{
				ParticipantID: participantID,
				SurveyKey:     surveyKey,
				Since:         since,
			})
			sCtx.PreviousResponses = append(sCtx.PreviousResponses, cResps...)
		default:
			return sCtx, errors.New("expression is not supported yet")
		}
	}
	return sCtx, nil
}

func (s *studyServiceServer) resolvePrefillRules(instanceID string, studyKey string, participantID string, rules []types.Expression) (prefills types.SurveyResponse, err error) {
	lastSurveyCache := map[string]types.SurveyResponse{}
	for _, rule := range rules {
		switch rule.Name {
		case "GET_LAST_SURVEY_ITEM":
			if len(rule.Data) < 2 {
				return prefills, errors.New("GET_LAST_SURVEY_ITEM must have at least two arguments")
			}
			surveyKey := rule.Data[0].Str
			itemKey := rule.Data[1].Str
			since := int64(0)
			if len(rule.Data) == 3 {
				// look up responses that are not older than:
				since = time.Now().Unix() - int64(rule.Data[2].Num)
			}

			previousResp, ok := lastSurveyCache[surveyKey]
			if !ok {
				resps, err := s.studyDBservice.FindSurveyResponses(instanceID, studyKey, studydb.ResponseQuery{
					ParticipantID: participantID,
					SurveyKey:     surveyKey,
					Limit:         1,
					Since:         since,
				})

				if err != nil || len(resps) < 1 {
					continue
				}
				lastSurveyCache[surveyKey] = resps[0]
				previousResp = resps[0]
			}

			for _, item := range previousResp.Responses {
				if item.Key == itemKey {
					prefills.Responses = append(prefills.Responses, item)
					break
				}
			}
		default:
			return prefills, fmt.Errorf("expression is not supported yet: %s", rule.Name)
		}
	}
	return prefills, nil
}

func (s *studyServiceServer) saveReports(instanceID string, studyKey string, reports map[string]types.Report, withResponseID string) {
	// save reports
	for _, report := range reports {
		report.ResponseID = withResponseID
		err := s.studyDBservice.SaveReport(instanceID, studyKey, report)
		if err != nil {
			logger.Error.Printf("unexpected error while save report: %v", err)
		} else {
			logger.Debug.Printf("Report with key '%s' for participant %s saved.", report.Key, report.ParticipantID)
		}
	}
}

func (s *studyServiceServer) prepareSurveyWithoutParticipant(instanceID string, studyKey string, surveyDef types.Survey) (*api.SurveyAndContext, error) {
	// empty irrelevant fields for this purpose
	surveyDef.ContextRules = nil
	surveyDef.PrefillRules = []types.Expression{}
	surveyDef.History = []types.SurveyVersion{}

	resp := &api.SurveyAndContext{
		Survey: surveyDef.ToAPI(),
	}
	return resp, nil
}

func (s *studyServiceServer) prepareSurveyForParticipant(instanceID string, studyKey string, participantID string, surveyDef types.Survey) (*api.SurveyAndContext, error) {
	// Prepare context
	surveyContext, err := s.resolveContextRules(instanceID, studyKey, participantID, surveyDef.ContextRules)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// Prepare prefill
	prefill, err := s.resolvePrefillRules(instanceID, studyKey, participantID, surveyDef.PrefillRules)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// empty irrelevant fields for this purpose
	surveyDef.ContextRules = nil
	surveyDef.PrefillRules = []types.Expression{}
	surveyDef.History = []types.SurveyVersion{}

	resp := &api.SurveyAndContext{
		Survey:  surveyDef.ToAPI(),
		Context: surveyContext.ToAPI(),
	}
	if len(prefill.Responses) > 0 {
		resp.Prefill = prefill.ToAPI()
	}
	return resp, nil
}

func (s *studyServiceServer) _getSurveyWithoutLogin(instanceID string, studyKey string, surveyKey string, tempParticipantID string) (*api.SurveyAndContext, error) {
	// Get survey definition:
	surveyDef, err := s.studyDBservice.FindSurveyDef(instanceID, studyKey, surveyKey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if surveyDef.AvailableFor == types.SURVEY_AVAILABLE_FOR_ACTIVE_PARTICIPANTS ||
		surveyDef.AvailableFor == "" {
		// ONLY FOR ACTIVE PARTICIPANTS: token must be present
		logger.Error.Printf("Trying to access survey that requires login. %s > %s > %s", instanceID, studyKey, surveyKey)
		return nil, status.Error(codes.InvalidArgument, "must login first")
	}

	if tempParticipantID == "" {
		// Without temporary participant
		if surveyDef.AvailableFor == types.SURVEY_AVAILABLE_FOR_TEMPORARY_PARTICIPANTS {
			// FOR ACTIVE OR TEMPORARY PARTICIPANTS: temporary participant id must be present
			logger.Error.Printf("Trying to access survey that requires at least temporary participant. %s > %s > %s", instanceID, studyKey, surveyKey)
			return nil, status.Error(codes.InvalidArgument, "must send a temporary participant id or login first")
		}

		resp, err := s.prepareSurveyWithoutParticipant(instanceID, studyKey, surveyDef)
		if err != nil {
			logger.Debug.Println(err)
		}
		return resp, err
	} else {
		// For temporary participant
		if !s.checkIfParticipantExists(instanceID, studyKey, tempParticipantID, types.PARTICIPANT_STUDY_STATUS_TEMPORARY) {
			logger.Error.Printf("Trying to access not existing temporary participant. %s > %s > %s : %s", instanceID, studyKey, surveyKey, tempParticipantID)
			return nil, status.Error(codes.PermissionDenied, "wrong participant id")
		}

		resp, err := s.prepareSurveyForParticipant(instanceID, studyKey, tempParticipantID, surveyDef)
		if err != nil {
			logger.Debug.Println(err)
		}
		return resp, err
	}
}

func (s *studyServiceServer) _getSurveyWithLoggedInUser(token *api_types.TokenInfos, studyKey string, surveyKey string, profileID string) (*api.SurveyAndContext, error) {
	// Get survey definition:
	surveyDef, err := s.studyDBservice.FindSurveyDef(token.InstanceId, studyKey, surveyKey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := utils.CheckIfProfileIDinToken(token, profileID); err != nil {
		s.SaveLogEvent(token.InstanceId, token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_WRONG_PROFILE_ID, "get assigned survey:"+profileID)
		return nil, status.Error(codes.Internal, "permission denied")
	}

	participantID, err := s.profileIDToParticipantID(token.InstanceId, studyKey, profileID)
	if err != nil {
		logger.Debug.Println(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !s.checkIfParticipantExists(token.InstanceId, studyKey, participantID, "") {
		resp, err := s.prepareSurveyWithoutParticipant(token.InstanceId, studyKey, surveyDef)
		if err != nil {
			logger.Debug.Println(err)
		}
		return resp, err
	} else {
		resp, err := s.prepareSurveyForParticipant(token.InstanceId, studyKey, participantID, surveyDef)
		if err != nil {
			logger.Debug.Println(err)
		}
		return resp, err
	}

}
