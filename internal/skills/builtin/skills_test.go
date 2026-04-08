package builtin_test

import (
	"strings"
	"testing"

	"github.com/akshaymemane/superclaw/internal/skills/builtin"
)

// skillCase describes the minimal contract we verify for every built-in skill.
type skillCase struct {
	skill        interface {
		Name() string
		Description() string
		Instructions() string
	}
	wantName        string
	wantInstructions []string // substrings that must appear in Instructions()
}

func TestBuiltinSkills(t *testing.T) {
	cases := []skillCase{
		{
			skill:    builtin.NewSummarizeSkill(),
			wantName: "summarize",
			wantInstructions: []string{
				"Lead with",
				"filler",
			},
		},
		{
			skill:    builtin.NewResearchSkill(),
			wantName: "research",
			wantInstructions: []string{
				"fetch",
				"cross-reference",
			},
		},
		{
			skill:    builtin.NewExtractSkill(),
			wantName: "extract",
			wantInstructions: []string{
				"schema",
				"NOT_FOUND",
			},
		},
		{
			skill:    builtin.NewCodingSkill(),
			wantName: "coding",
			wantInstructions: []string{
				"read_file",
				"patch_file",
				"smallest change",
			},
		},
		{
			skill:    builtin.NewGitHubSkill(),
			wantName: "github",
			wantInstructions: []string{
				"gh",
				"force",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.wantName, func(t *testing.T) {
			if tc.skill.Name() != tc.wantName {
				t.Errorf("Name() = %q, want %q", tc.skill.Name(), tc.wantName)
			}

			if tc.skill.Description() == "" {
				t.Error("Description() must not be empty")
			}

			instr := tc.skill.Instructions()
			if instr == "" {
				t.Fatal("Instructions() must not be empty")
			}

			for _, sub := range tc.wantInstructions {
				if !strings.Contains(strings.ToLower(instr), strings.ToLower(sub)) {
					t.Errorf("Instructions() missing %q\ngot:\n%s", sub, instr)
				}
			}
		})
	}
}

func TestSkillsInterface(t *testing.T) {
	// Verify all skills satisfy the skills.Skill interface at compile time
	// by calling the methods through a local interface.
	type skillIface interface {
		Name() string
		Description() string
		Instructions() string
	}

	skills := []skillIface{
		builtin.NewSummarizeSkill(),
		builtin.NewResearchSkill(),
		builtin.NewExtractSkill(),
		builtin.NewCodingSkill(),
		builtin.NewGitHubSkill(),
	}

	seen := map[string]bool{}
	for _, s := range skills {
		if seen[s.Name()] {
			t.Errorf("duplicate skill name %q", s.Name())
		}
		seen[s.Name()] = true
	}
}
