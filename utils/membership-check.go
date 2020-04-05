package utils

import "github.com/influenzanet/study-service/models"

// CheckIfMember finds user in the member array and if role is contained in hasRoleFrom slice (ignored when empty)
func CheckIfMember(userID string, members []models.StudyMember, hasRoleFrom []string) bool {
	for _, member := range members {
		if member.UserID == userID {
			if len(hasRoleFrom) > 0 {
				for _, r := range hasRoleFrom {
					if member.Role == r {
						return true
					}
				}
				return false
			}
			return true
		}
	}
	return false
}
