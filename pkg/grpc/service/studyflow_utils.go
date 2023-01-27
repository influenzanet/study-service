package service

import (
	"errors"
	"fmt"
	"strings"
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

func (s *studyServiceServer) profileIDToParticipantID(instanceID string, studyKey string, userID string, ignoreForConfidentialResponses bool) (string, string, error) {
	idMappingMethod, studySecret, err := s.studyDBservice.GetStudySecretKey(instanceID, studyKey)
	if err != nil {
		return "", "", err
	}
	pID, err := utils.ProfileIDtoParticipantID(userID, s.StudyGlobalSecret, studySecret, idMappingMethod)
	if err != nil {
		return "", "", err
	}
	if ignoreForConfidentialResponses {
		return pID, "", nil
	}

	// for confidential responses
	pID2, err := utils.ProfileIDtoParticipantID(pID, s.StudyGlobalSecret, studySecret, idMappingMethod)
	if err != nil {
		return "", "", err
	}
	return pID, pID2, nil
}

func (s *studyServiceServer) checkIfParticipantExists(instanceID string, studyKey string, participantID string, withStatus string) bool {
	pState, err := s.studyDBservice.FindParticipantState(instanceID, studyKey, participantID)
	if err != nil || (withStatus != "" && pState.StudyStatus != withStatus) {
		return false
	}
	return err == nil
}

func (s *studyServiceServer) getAndPerformStudyRules(instanceID string, studyKey string, pState types.ParticipantState, event types.StudyEvent) (newState studyengine.ActionData, err error) {
	newState = studyengine.ActionData{
		PState:          pState,
		ReportsToCreate: map[string]types.Report{},
	}

	rules, err := s.studyDBservice.GetStudyRules(instanceID, studyKey)
	if err != nil {
		return
	}
	for _, rule := range rules {
		newState, err = studyengine.ActionEval(rule, newState, event, studyengine.ActionConfigs{
			DBService:              s.studyDBservice,
			ExternalServiceConfigs: s.studyEngineExternalServices,
		})
		if err != nil {
			return
		}
	}

	return newState, nil
}

func (s *studyServiceServer) resolveContextRules(instanceID string, studyKey string, pState types.ParticipantState, rules *types.SurveyContextDef) (sCtx types.SurveyContext, err error) {
	participantID := pState.ParticipantID

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
		case "PREFILL_SLOT_WITH_VALUE":
			if len(rule.Data) < 3 {
				logger.Error.Printf("not enough arguments in %v", rule)
				continue
			}
			itemKey := rule.Data[0].Str
			slotKey := rule.Data[1].Str
			targetValue := rule.Data[2]

			prefillItem := types.SurveyItemResponse{
				Key: itemKey,
			}

			// Find item if already exits
			pItemIndex := -1
			for i, p := range prefills.Responses {
				if p.Key == itemKey {
					prefillItem = p
					pItemIndex = i
					break
				}
			}

			slotKeyParts := strings.Split(slotKey, ".")
			if len(slotKeyParts) < 1 {
				logger.Error.Printf("prefill rule has invalid slot key: %v", rule)
				return
			}

			respItem := prefillItem.Response
			if respItem == nil {
				respItem = &types.ResponseItem{Key: slotKeyParts[0], Items: []*types.ResponseItem{}}
			}

			var currentRespItem *types.ResponseItem
			for _, rKey := range slotKeyParts {
				if currentRespItem == nil {
					currentRespItem = respItem
					continue
				}

				found := false
				for _, item := range currentRespItem.Items {
					if item.Key == rKey {
						found = true
						currentRespItem = item
						break
					}
				}
				if !found {
					newItem := types.ResponseItem{Key: rKey, Items: []*types.ResponseItem{}}
					currentRespItem.Items = append(currentRespItem.Items, &newItem)
					currentRespItem = currentRespItem.Items[len(currentRespItem.Items)-1]
				}
			}

			if targetValue.DType == "num" {
				currentRespItem.Dtype = "number"
				currentRespItem.Value = fmt.Sprintf("%f", targetValue.Num)
			} else {
				currentRespItem.Value = targetValue.Str
			}
			prefillItem.Response = respItem

			if pItemIndex > -1 {
				prefills.Responses[pItemIndex] = prefillItem
			} else {
				prefills.Responses = append(prefills.Responses, prefillItem)
			}
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

func (s *studyServiceServer) prepareSurveyWithoutParticipant(instanceID string, studyKey string, surveyDef *types.Survey) (*api.SurveyAndContext, error) {
	// empty irrelevant fields for this purpose
	surveyDef.ContextRules = nil
	surveyDef.PrefillRules = []types.Expression{}

	resp := &api.SurveyAndContext{
		Survey: surveyDef.ToAPI(),
	}
	return resp, nil
}

func isSurveyAssignedAndActive(pState types.ParticipantState, surveyKey string) bool {
	now := time.Now().Unix()

	for _, as := range pState.AssignedSurveys {
		if as.SurveyKey != surveyKey {
			continue
		}

		if as.ValidFrom > 0 && now < as.ValidFrom {
			continue
		}

		if as.ValidUntil > 0 && now > as.ValidUntil {
			continue
		}

		// --> survey is currently active
		return true
	}

	return false
}

func (s *studyServiceServer) prepareSurveyForParticipant(instanceID string, studyKey string, participantID string, surveyDef *types.Survey) (*api.SurveyAndContext, error) {
	pState, err := s.studyDBservice.FindParticipantState(instanceID, studyKey, participantID)
	if err != nil {
		return nil, errors.New("no participant with this id in this study")
	}

	// If survey requires assigned state, ensure it's assigned
	if surveyDef.AvailableFor == types.SURVEY_AVAILABLE_FOR_PARTICIPANTS_IF_ASSIGNED {
		if !isSurveyAssignedAndActive(pState, surveyDef.SurveyDefinition.Key) {
			msg := fmt.Sprintf("Participant %s trying to access survey (%s) when it's not assigned/active", participantID, surveyDef.SurveyDefinition.Key)
			logger.Warning.Println(msg)
			return nil, errors.New(msg)
		}
	}

	// Prepare context
	surveyContext, err := s.resolveContextRules(instanceID, studyKey, pState, surveyDef.ContextRules)
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
	surveyDef, err := s.studyDBservice.FindCurrentSurveyDef(instanceID, studyKey, surveyKey, false)
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
		if surveyDef.AvailableFor != types.SURVEY_AVAILABLE_FOR_PUBLIC {
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
	surveyDef, err := s.studyDBservice.FindCurrentSurveyDef(token.InstanceId, studyKey, surveyKey, false)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := utils.CheckIfProfileIDinToken(token, profileID); err != nil {
		s.SaveLogEvent(token.InstanceId, token.Id, loggingAPI.LogEventType_SECURITY, constants.LOG_EVENT_WRONG_PROFILE_ID, "get assigned survey:"+profileID)
		return nil, status.Error(codes.Internal, "permission denied")
	}

	participantID, _, err := s.profileIDToParticipantID(token.InstanceId, studyKey, profileID, true)
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
