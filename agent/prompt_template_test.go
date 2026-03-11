package agent

import (
	"strings"
	"testing"
)

func TestBuildSystemPrompt_SkillItemsUseAuthProfilesWithoutRequirements(t *testing.T) {
	prompt := BuildSystemPrompt(nil, PromptSpec{
		Identity: "identity",
		Skills: []PromptSkill{
			{
				Name:         "jsonbill",
				FilePath:     "file_state_dir/skills/jsonbill/SKILL.md",
				Description:  "Generate invoice PDF.",
				AuthProfiles: []string{"jsonbill"},
			},
			{
				Name:        "alpha",
				FilePath:    "file_state_dir/skills/alpha/SKILL.md",
				Description: "No auth needed.",
			},
		},
	})

	if !strings.Contains(prompt, "auth_profiles: jsonbill") {
		t.Fatalf("prompt missing auth profiles line: %s", prompt)
	}
	if strings.Contains(prompt, "Requirements:") {
		t.Fatalf("prompt should not contain Requirements: %s", prompt)
	}
	if strings.Contains(prompt, "auth_profiles: \n") {
		t.Fatalf("prompt should not emit empty AuthProfiles lines: %s", prompt)
	}
}
