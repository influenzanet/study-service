package utils

import "github.com/influenzanet/study-service/api"

// IsTokenEmpty check a token from api if it's empty
func IsTokenEmpty(t *api.TokenInfos) bool {
	if t == nil || t.Id == "" || t.InstanceId == "" {
		return true
	}
	return false
}
