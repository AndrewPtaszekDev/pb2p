package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"go.yaml.in/yaml/v3"
)

func GetYAMLKey(path string, key string) (string, error) {
	data := map[string]any{}

	contents, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", path, err)
	}

	if err := yaml.Unmarshal(contents, &data); err != nil {
		return "", fmt.Errorf("parsing %s: %w", path, err)
	}

	raw, ok := data[key]
	if !ok {
		return "", fmt.Errorf("key %q not found in %s", key, path)
	}

	value, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("key %q in %s is not a string (got %T)", key, path, raw)
	}

	return value, nil
}

// readYAMLMap loads path into a top-level string-keyed map. A missing file is
// not an error; the caller starts from an empty map.
func readYAMLMap(path string) (map[string]any, error) {
	data := map[string]any{}

	existing, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return data, nil
		}
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	if len(existing) == 0 {
		return data, nil
	}

	if err := yaml.Unmarshal(existing, &data); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	return data, nil
}

// writeYAMLMap marshals data and writes it to path atomically.
func writeYAMLMap(path string, data map[string]any) error {
	out, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshaling yaml: %w", err)
	}

	if err := writeFileAtomic(path, out, 0644, false); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}

	return nil
}
