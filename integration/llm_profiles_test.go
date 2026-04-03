package integration

import "testing"

func TestRuntimeLLMProfileSelectionIsInstanceScoped(t *testing.T) {
	cfg := Config{
		Overrides: map[string]any{
			"llm.provider": "openai",
			"llm.model":    "gpt-5.2",
			"llm.profiles": map[string]any{
				"cheap": map[string]any{
					"model": "gpt-4.1-mini",
				},
			},
		},
	}
	rtA := New(cfg)
	rtB := New(cfg)

	if err := rtA.SetLLMProfile("cheap"); err != nil {
		t.Fatalf("rtA.SetLLMProfile() error = %v", err)
	}

	selA, err := rtA.GetLLMProfileSelection()
	if err != nil {
		t.Fatalf("rtA.GetLLMProfileSelection() error = %v", err)
	}
	if selA.Mode != "manual" || selA.ManualProfile != "cheap" {
		t.Fatalf("selA = %#v, want manual cheap", selA)
	}

	selB, err := rtB.GetLLMProfileSelection()
	if err != nil {
		t.Fatalf("rtB.GetLLMProfileSelection() error = %v", err)
	}
	if selB.Mode != "auto" {
		t.Fatalf("selB.Mode = %q, want auto", selB.Mode)
	}
	if selB.ManualProfile != "" {
		t.Fatalf("selB.ManualProfile = %q, want empty", selB.ManualProfile)
	}
}
