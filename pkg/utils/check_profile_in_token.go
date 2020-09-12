package utils

import (
	"errors"

	"github.com/influenzanet/go-utils/pkg/api_types"
)

func CheckIfProfileIDinToken(token *api_types.TokenInfos, profileID string) error {
	if token == nil {
		return errors.New("invalid token")
	}
	if token.ProfilId == profileID {
		return nil
	}
	for _, p := range token.OtherProfileIds {
		if p == profileID {
			return nil
		}
	}
	return errors.New("profile id not valid for this user")
}
