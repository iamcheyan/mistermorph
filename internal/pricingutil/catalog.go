package pricingutil

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	uniaiapi "github.com/quailyquaily/uniai"
)

func LoadCatalog(path, configPath string) (*uniaiapi.PricingCatalog, string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return defaultCatalog()
	}
	resolved := resolveConfigRelativePath(path, configPath)
	data, err := os.ReadFile(resolved)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultCatalog()
		}
		return nil, "", fmt.Errorf("read llm.pricing_file %q: %w", path, err)
	}
	pricing, err := uniaiapi.ParsePricingYAML(data)
	if err != nil {
		return nil, "", fmt.Errorf("parse llm.pricing_file %q: %w", path, err)
	}
	sum := sha256.Sum256(data)
	return pricing, hex.EncodeToString(sum[:]), nil
}

func resolveConfigRelativePath(path, configPath string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	configPath = strings.TrimSpace(configPath)
	if configPath == "" {
		return filepath.Clean(path)
	}
	return filepath.Clean(filepath.Join(filepath.Dir(configPath), path))
}

func defaultCatalog() (*uniaiapi.PricingCatalog, string, error) {
	catalog := uniaiapi.DefaultPricingCatalog()
	digest, err := catalogDigest(catalog)
	if err != nil {
		return nil, "", err
	}
	return catalog, digest, nil
}

func catalogDigest(catalog *uniaiapi.PricingCatalog) (string, error) {
	if catalog == nil {
		return "", nil
	}
	data, err := json.Marshal(catalog)
	if err != nil {
		return "", fmt.Errorf("marshal pricing catalog digest: %w", err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}
