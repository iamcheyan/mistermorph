package skillsutil

import (
	"strings"
	"testing"
)

func TestRenderSkillStatusShowsLoadedAndAvailable(t *testing.T) {
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
	for _, want := range []string{"**Loaded Skills (1)**", "* `alpha`:", "**Available Skills (1)**", "* `beta`:"} {
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
	if !strings.Contains(out, "**Loaded Skills (1)**") || !strings.Contains(out, "* `beta`:") {
		t.Fatalf("status missing current loaded skill:\n%s", out)
	}
	if !strings.Contains(out, "**Available Skills (1)**") || !strings.Contains(out, "* `alpha`:") {
		t.Fatalf("status missing available skill:\n%s", out)
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
	for _, want := range []string{"> Skill is disabled"} {
		if !strings.Contains(out, want) {
			t.Fatalf("status missing %q:\n%s", want, out)
		}
	}
}
