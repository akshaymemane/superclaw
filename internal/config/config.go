// Package config handles loading and merging superclaw.json config files.
package config

import (
	"encoding/json"
	"errors"
	"os"
)

// DefaultPath is the filename looked up in the current directory.
const DefaultPath = "superclaw.json"

// File represents the superclaw.json config format.
// All fields are optional; zero values mean "use the default".
type File struct {
	Model          string   `json:"model"`
	MaxSteps       int      `json:"max_steps"`
	MaxTokens      int64    `json:"max_tokens"`
	WorkDir        string   `json:"work_dir"`
	Tools          []string `json:"tools"`
	Skills         []string `json:"skills"`
	SystemPrompt   string   `json:"system_prompt"`
	TimeoutSeconds int      `json:"timeout_seconds"`
	MaxFetchCalls  int      `json:"max_fetch_calls"`
}

// Load reads and parses the file at path.
func Load(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var f File
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	return &f, nil
}

// LoadOrDefault loads path if it exists.
// Returns an empty File (not an error) when path does not exist.
func LoadOrDefault(path string) (*File, error) {
	f, err := Load(path)
	if errors.Is(err, os.ErrNotExist) {
		return &File{}, nil
	}
	return f, err
}
