package llmutil

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"unicode"
)

func resolveStructRefs(target any) error {
	rv := reflect.ValueOf(target)
	if !rv.IsValid() || rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer")
	}
	if rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target must point to a struct")
	}
	return resolveStructRefValues(rv.Elem())
}

func resolveStructRefValues(v reflect.Value) error {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)
		if !fieldValue.CanSet() {
			continue
		}
		switch fieldValue.Kind() {
		case reflect.Struct:
			if err := resolveStructRefValues(fieldValue); err != nil {
				return err
			}
		case reflect.String:
			if !strings.HasSuffix(field.Name, "Ref") {
				continue
			}
			baseName := strings.TrimSuffix(field.Name, "Ref")
			baseValue := v.FieldByName(baseName)
			if !baseValue.IsValid() || !baseValue.CanSet() || baseValue.Kind() != reflect.String {
				continue
			}
			resolved, err := envRefOrValue(fieldValue.String(), baseValue.String())
			if err != nil {
				return fmt.Errorf("%s: %w", configPathForField(field), err)
			}
			baseValue.SetString(resolved)
		}
	}
	return nil
}

func envRefOrValue(envRef, value string) (string, error) {
	envRef = strings.TrimSpace(envRef)
	if envRef == "" {
		return strings.TrimSpace(value), nil
	}
	val, ok := os.LookupEnv(envRef)
	if !ok {
		return "", fmt.Errorf("env %q is not set", envRef)
	}
	val = strings.TrimSpace(val)
	if val == "" {
		return "", fmt.Errorf("env %q is empty", envRef)
	}
	return val, nil
}

func configPathForField(field reflect.StructField) string {
	if path := strings.TrimSpace(field.Tag.Get("config")); path != "" {
		return path
	}
	if path := strings.TrimSpace(field.Tag.Get("mapstructure")); path != "" && path != ",squash" {
		return path
	}
	return toSnakeCase(field.Name)
}

func toSnakeCase(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
