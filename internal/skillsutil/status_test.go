package skillsutil

import (
	"strings"
	"testing"
)

func TestRenderSkillStatusShowsLoadedAndNotLoaded(t *testing.T) {
	root := t.TempDir()
	writeSkill(t, root, "alpha")
	writeSkill(t, root, "beta")

	out, err := RenderSkillStatus(SkillsConfig{
		Roots:     []string{root},
		Enabled:   true,
		Requested: []string{"alpha"},
	}, nil)
	if err != nil {
		t.Fatalf("RenderSkillStatus() error = %v", err)
	}
	for _, want := range []string{"Skills: enabled", "Loaded (1):", "- alpha", "Not loaded (1):", "- beta"} {
		if !strings.Contains(out, want) {
			t.Fatalf("status missing %q:\n%s", want, out)
		}
	}
}

func TestRenderSkillStatusUsesCurrentLoadedSkills(t *testing.T) {
	root := t.TempDir()
	writeSkill(t, root, "alpha")
	writeSkill(t, root, "beta")

	out, err := RenderSkillStatus(SkillsConfig{
		Roots:   []string{root},
		Enabled: true,
	}, []string{"beta"})
	if err != nil {
		t.Fatalf("RenderSkillStatus() error = %v", err)
	}
	if !strings.Contains(out, "Loaded (1):") || !strings.Contains(out, "- beta") {
		t.Fatalf("status missing current loaded skill:\n%s", out)
	}
	if !strings.Contains(out, "Not loaded (1):") || !strings.Contains(out, "- alpha") {
		t.Fatalf("status missing not loaded skill:\n%s", out)
	}
}

func TestRenderSkillStatusDisabled(t *testing.T) {
	root := t.TempDir()
	writeSkill(t, root, "alpha")

	out, err := RenderSkillStatus(SkillsConfig{
		Roots:   []string{root},
		Enabled: false,
	}, nil)
	if err != nil {
		t.Fatalf("RenderSkillStatus() error = %v", err)
	}
	for _, want := range []string{"Skills: disabled", "Loaded: none", "Not loaded (1):", "- alpha"} {
		if !strings.Contains(out, want) {
			t.Fatalf("status missing %q:\n%s", want, out)
		}
	}
}
