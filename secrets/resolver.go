package secrets

import (
	"context"
	"fmt"
	"os"
	"strings"
)

type Resolver interface {
	Resolve(ctx context.Context, secretRef string) (string, error)
}

// EnvResolver resolves secret_ref values directly from environment variables.
//
// The MVP behavior is fail-closed:
// - missing/unset env var => error
// - empty value => error
type EnvResolver struct{}

func (r *EnvResolver) Resolve(ctx context.Context, secretRef string) (string, error) {
	_ = ctx

	envName := strings.TrimSpace(secretRef)
	if envName == "" {
		return "", fmt.Errorf("empty secret_ref")
	}

	val, ok := os.LookupEnv(envName)
	if !ok {
		return "", fmt.Errorf("secret not found (env var %q is not set)", envName)
	}
	if strings.TrimSpace(val) == "" {
		return "", fmt.Errorf("secret is empty (env var %q)", envName)
	}
	return val, nil
}
