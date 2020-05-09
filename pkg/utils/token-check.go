package utils

import (
	"strings"

	"github.com/influenzanet/study-service/pkg/api"
)

// IsTokenEmpty check a token from api if it's empty
func IsTokenEmpty(t *api.TokenInfos) bool {
	if t == nil || t.Id == "" || t.InstanceId == "" {
		return true
	}
	return false
}

func GetUsernameFromToken(t *api.TokenInfos) string {
	if t == nil {
		return ""
	}
	username, ok := t.Payload["username"]
	if !ok {
		return ""
	}
	return username
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

// CheckIfAnyRolesInToken if token contains any of the roles
func CheckIfAnyRolesInToken(t *api.TokenInfos, requiredRoles []string) bool {
	if t == nil {
		return false
	}
	if val, ok := t.Payload["roles"]; ok {
		roles := strings.Split(val, ",")
		for _, r := range roles {
			for _, rr := range requiredRoles {
				if r == rr {
					return true
				}
			}
		}
	}
	return false
}
