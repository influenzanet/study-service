package utils

import "github.com/influenzanet/study-service/models"

// CheckIfMember finds user in the member array and if role is equal to requiredRole (ignored when empty)
func CheckIfMember(userID string, members []models.StudyMember, requiredRole string) bool {
	for _, member := range members {
		if member.UserID == userID {
			if requiredRole != "" {
				return member.Role == requiredRole
			}
			return true
		}
	}
	return false
}
