package studytimer

import (
	"testing"

	"github.com/influenzanet/study-service/pkg/types"
)

func TestHasRuleForEventType(t *testing.T) {
	event := types.StudyEvent{
		Type: "TIMER",
	}
	s := StudyTimerService{}

	t.Run("with no timer event", func(t *testing.T) {
		rules := []types.Expression{
			{Name: "IFTHEN", Data: []types.ExpressionArg{
				{Exp: &types.Expression{Name: "checkEventType", Data: []types.ExpressionArg{{
					Str: "SUBMIT",
				}}}},
			}},
		}
		hasRule := s.hasRuleForEventType(rules, event)
		if hasRule {
			t.Error("should return false")
		}
	})

	t.Run("with strange rule", func(t *testing.T) {
		rules := []types.Expression{
			{Name: "IFTHEN", Data: []types.ExpressionArg{
				{Exp: &types.Expression{Name: "checkEventType", Data: []types.ExpressionArg{{
					Str: "SUBMIT",
				}}}},
			}},
			{Name: "Else", Data: []types.ExpressionArg{
				{Exp: &types.Expression{Name: "other", Data: []types.ExpressionArg{{
					Num: 0,
				}}}},
			}},
		}
		hasRule := s.hasRuleForEventType(rules, event)
		if hasRule {
			t.Error("should return false")
		}
	})

	t.Run("with timer event", func(t *testing.T) {
		rules := []types.Expression{
			{Name: "IFTHEN", Data: []types.ExpressionArg{
				{Exp: &types.Expression{Name: "checkEventType", Data: []types.ExpressionArg{{
					Str: "SUBMIT",
				}}}},
			}},
			{Name: "Else", Data: []types.ExpressionArg{
				{Exp: &types.Expression{Name: "other", Data: []types.ExpressionArg{{
					Num: 0,
				}}}},
			}},
			{Name: "IFTHEN", Data: []types.ExpressionArg{
				{Exp: &types.Expression{Name: "checkEventType", Data: []types.ExpressionArg{{
					Str: "TIMER",
				}}}},
			}},
		}
		hasRule := s.hasRuleForEventType(rules, event)
		if !hasRule {
			t.Error("should return true")
		}
	})

}
