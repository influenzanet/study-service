package utils

import (
	"testing"

	"github.com/influenzanet/go-utils/pkg/api_types"
)

func TestCheckIfProfileIDinToken(t *testing.T) {
	token := &api_types.TokenInfos{
		Id:               "userid",
		InstanceId:       "instanceid",
		AccountConfirmed: true,
		ProfilId:         "mainprofileid",
		OtherProfileIds:  []string{"2ndprofile", "3rdprofile"},
	}
	t.Run("not in token", func(t *testing.T) {
		err := CheckIfProfileIDinToken(token, "notintoken")
		if err == nil {
			t.Error("should fail")
		}
	})
	t.Run("main profile", func(t *testing.T) {
		err := CheckIfProfileIDinToken(token, "mainprofileid")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	t.Run("other profile", func(t *testing.T) {
		err := CheckIfProfileIDinToken(token, "2ndprofile")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		err = CheckIfProfileIDinToken(token, "3rdprofile")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}
