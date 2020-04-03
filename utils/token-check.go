package utils

import (
	"strings"

	"github.com/influenzanet/study-service/api"
)

// IsTokenEmpty check a token from api if it's empty
func IsTokenEmpty(t *api.TokenInfos) bool {
	if t == nil || t.Id == "" || t.InstanceId == "" {
		return true
	}
	return false
}

// CheckRoleInToken Check if role is present in the token
func CheckRoleInToken(t *api.TokenInfos, role string) bool {
	if t == nil {
		return false
	}
	if val, ok := t.Payload["roles"]; ok {
		roles := strings.Split(val, ",")
		for _, r := range roles {
			if r == role {
				return true
			}
		}
	}
	return false
}
