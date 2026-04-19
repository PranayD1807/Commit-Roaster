package roaster

import (
	"testing"
)

func TestAnalyze(t *testing.T) {
	tests := []struct {
		name          string
		msg           string
		expectScore   int
		expectPenalty bool
	}{
		{
			name:          "Perfect commit message",
			msg:           "feat: add support for git commit roasting\n\n- Added new hook installer\n- Tested with unit tests\n",
			expectScore:   10,
			expectPenalty: false,
		},
		{
			name:          "Terrible commit message",
			msg:           "WIP fixes",
			expectPenalty: true, // Should match multiple rules (too short, no scope, etc.)
		},
		{
			name:          "All caps shouting",
			msg:           "FIX BUG IN PRODUCTION",
			expectPenalty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Analyze(tt.msg)

			if !tt.expectPenalty && result.Score != 10 {
				t.Errorf("Expected perfect score of 10 for clean commit, got %d. Violations: %v", result.Score, result.Violations)
			}

			if tt.expectPenalty && len(result.Violations) == 0 {
				t.Errorf("Expected violations for a bad commit, got none.")
			}

			if tt.expectPenalty && result.Score == 10 {
				t.Errorf("Expected penalized score, got 10")
			}
		})
	}
}
