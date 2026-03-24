package skillscmd

import (
	"context"
	"strings"
	"testing"

	"github.com/quailyquaily/mistermorph/llm"
	"github.com/spf13/viper"
)

type stubRemoteSkillReviewClient struct{}

func (stubRemoteSkillReviewClient) Chat(context.Context, llm.Request) (llm.Result, error) {
	return llm.Result{}, nil
}

func TestSkillsInstallCommandExposesYesFlagShorthand(t *testing.T) {
	cmd := NewSkillsInstallBuiltinCmd()
	flag := cmd.Flags().Lookup("yes")
	if flag == nil {
		t.Fatalf("expected --yes flag to exist")
	}
	if flag.Shorthand != "y" {
		t.Fatalf("expected --yes shorthand to be -y, got %q", flag.Shorthand)
	}
}

func TestLLMClientForRemoteSkillReviewRequiresModel(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	viper.Set("llm.provider", "openai")
	viper.Set("llm.api_key", "sk-test")
	viper.Set("llm.model", "")

	_, _, err := llmClientForRemoteSkillReview()
	if err == nil {
		t.Fatal("expected missing llm.model error")
	}
	if !strings.Contains(err.Error(), "missing llm.model") {
		t.Fatalf("error = %q, want missing llm.model", err.Error())
	}
}

func TestReviewRemoteSkillRequiresModel(t *testing.T) {
	_, err := reviewRemoteSkill(context.Background(), stubRemoteSkillReviewClient{}, "", "https://example.com/SKILL.md", []byte("body"))
	if err == nil {
		t.Fatal("expected missing llm.model error")
	}
	if !strings.Contains(err.Error(), "missing llm.model") {
		t.Fatalf("error = %q, want missing llm.model", err.Error())
	}
}
