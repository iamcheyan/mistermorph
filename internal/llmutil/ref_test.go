package llmutil

import (
	"strings"
	"testing"
)

func TestResolveStructRefs_RecursivePairing(t *testing.T) {
	t.Setenv("TEST_CHILD_API_KEY", "env-child-key")

	type child struct {
		APIKey    string `config:"llm.child.api_key"`
		APIKeyRef string `config:"llm.child.api_key_ref"`
	}
	type sample struct {
		Child child
	}

	v := sample{
		Child: child{
			APIKey:    "plain-child-key",
			APIKeyRef: "TEST_CHILD_API_KEY",
		},
	}
	if err := resolveStructRefs(&v); err != nil {
		t.Fatalf("resolveStructRefs() error = %v", err)
	}
	if v.Child.APIKey != "env-child-key" {
		t.Fatalf("api key = %q, want env-child-key", v.Child.APIKey)
	}
}

func TestResolveStructRefs_MissingEnvUsesConfigPath(t *testing.T) {
	type sample struct {
		APIKey    string `config:"llm.api_key"`
		APIKeyRef string `config:"llm.api_key_ref"`
	}

	v := sample{APIKeyRef: "MISSING_LLM_API_KEY"}
	err := resolveStructRefs(&v)
	if err == nil {
		t.Fatalf("expected missing env error")
	}
	if got := err.Error(); !strings.HasPrefix(got, "llm.api_key_ref") {
		t.Fatalf("unexpected error: %v", err)
	}
}
