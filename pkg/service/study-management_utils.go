package service

import (
	"errors"

	"github.com/influenzanet/study-service/pkg/utils"
)

func (s *studyServiceServer) HasRoleInStudy(instanceID string, studyKey string, userID string, hasAnyOfRoles []string) error {
	members, err := s.studyDBservice.GetStudyMembers(instanceID, studyKey)
	if err != nil {
		return err
	}
	if !utils.CheckIfMember(userID, members, hasAnyOfRoles) {
		return errors.New("not authorized to access this study")
	}
	return nil
}
