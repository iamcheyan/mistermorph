package skillsutil

import (
	"strings"
	"testing"
)

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
