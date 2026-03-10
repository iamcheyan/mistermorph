package llmutil

import (
	"strings"
	"testing"
)

func TestResolveStructEnvRefs_RecursivePairing(t *testing.T) {
	t.Setenv("TEST_CHILD_API_KEY", "env-child-key")

	type child struct {
		APIKey       string `config:"llm.child.api_key"`
		APIKeyEnvRef string `config:"llm.child.api_key_env_ref"`
	}
	type sample struct {
		Child child
	}

	v := sample{
		Child: child{
			APIKey:       "plain-child-key",
			APIKeyEnvRef: "TEST_CHILD_API_KEY",
		},
	}
	if err := resolveStructEnvRefs(&v); err != nil {
		t.Fatalf("resolveStructEnvRefs() error = %v", err)
	}
	if v.Child.APIKey != "env-child-key" {
		t.Fatalf("api key = %q, want env-child-key", v.Child.APIKey)
	}
}

func TestResolveStructEnvRefs_MissingEnvUsesConfigPath(t *testing.T) {
	type sample struct {
		APIKey       string `config:"llm.api_key"`
		APIKeyEnvRef string `config:"llm.api_key_env_ref"`
	}

	v := sample{APIKeyEnvRef: "MISSING_LLM_API_KEY"}
	err := resolveStructEnvRefs(&v)
	if err == nil {
		t.Fatalf("expected missing env error")
	}
	if got := err.Error(); !strings.HasPrefix(got, "llm.api_key_env_ref") {
		t.Fatalf("unexpected error: %v", err)
	}
}
