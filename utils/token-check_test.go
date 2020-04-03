package utils

import (
	"strings"
	"testing"

	"github.com/influenzanet/study-service/api"
)

func TestIsTokenEmpty(t *testing.T) {
	t.Run("check with nil input", func(t *testing.T) {
		if !IsTokenEmpty(nil) {
			t.Error("should be true")
		}
	})

	t.Run("check with empty id", func(t *testing.T) {
		if !IsTokenEmpty(&api.TokenInfos{Id: "", InstanceId: "testid"}) {
			t.Error("should be true")
		}
	})

	t.Run("check with empty InstanceId", func(t *testing.T) {
		if !IsTokenEmpty(&api.TokenInfos{InstanceId: "", Id: "testid"}) {
			t.Error("should be true")
		}
	})

	t.Run("check with not empty id", func(t *testing.T) {
		if IsTokenEmpty(&api.TokenInfos{Id: "testid", InstanceId: "testid"}) {
			t.Error("should be false")
		}
	})
}

func TestCheckRoleInToken(t *testing.T) {
	t.Run("check with nil input", func(t *testing.T) {
		if CheckRoleInToken(nil, "") {
			t.Error("should be false")
		}
	})

	t.Run("check with no payload ", func(t *testing.T) {
		tokenInf := &api.TokenInfos{
			Id: "testid",
		}
		if CheckRoleInToken(tokenInf, "testrole") {
			t.Error("should be false")
		}
	})

	t.Run("check with single role - wrong", func(t *testing.T) {
		payload := map[string]string{}
		payload["roles"] = strings.Join([]string{"notthesame"}, ",")

		tokenInf := &api.TokenInfos{
			Id:      "testid",
			Payload: payload,
		}
		if CheckRoleInToken(tokenInf, "testrole") {
			t.Error("should be false")
		}
	})

	t.Run("check with single role - right", func(t *testing.T) {
		payload := map[string]string{}
		payload["roles"] = strings.Join([]string{"testrole"}, ",")
		tokenInf := &api.TokenInfos{
			Id:      "testid",
			Payload: payload,
		}
		if !CheckRoleInToken(tokenInf, "testrole") {
			t.Error("should be true")
		}
	})

	t.Run("check with multiple roles - wrong", func(t *testing.T) {
		payload := map[string]string{}
		payload["roles"] = strings.Join([]string{"r1", "r2", "r4"}, ",")
		tokenInf := &api.TokenInfos{
			Id:      "testid",
			Payload: payload,
		}
		if CheckRoleInToken(tokenInf, "testrole") {
			t.Error("should be false")
		}
	})

	t.Run("check with multiple roles - right", func(t *testing.T) {
		payload := map[string]string{}
		payload["roles"] = strings.Join([]string{"r1", "r2", "r4", "testrole"}, ",")
		tokenInf := &api.TokenInfos{
			Id:      "testid",
			Payload: payload,
		}
		if !CheckRoleInToken(tokenInf, "testrole") {
			t.Error("should be true")
		}
	})
}
