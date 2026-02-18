package bundles

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// DerefString safely dereferences a *string, returning empty string if nil.
func DerefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// DerefFloat64 safely dereferences a *float64, returning 0 if nil.
func DerefFloat64(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}

// ReadJSON reads a JSON file and unmarshals it into the provided value.
func ReadJSON(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

// WriteJSON writes a value as indented JSON to a file.
// Creates parent directories if they don't exist.
func WriteJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
