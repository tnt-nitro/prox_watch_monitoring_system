package rules

import (
	"testing"
	"time"

	"prox-watch/internal/config"
	"prox-watch/internal/state"
)

func TestEvaluator_Evaluate(t *testing.T) {
	cfg := config.DefaultConfig()
	eval := NewEvaluator(cfg)

	now := time.Now()
	baseTime := now.Add(-5 * time.Minute)

	tests := []struct {
		name        string
		cs          state.CountState
		now         time.Time
		isHardError bool
		expected    Severity
	}{
		{
			name: "hard_error_immediate_crit",
			cs: state.CountState{
				EventID:   "host.network.link_down",
				Count:     1,
				FirstSeen: baseTime,
				LastSeen:  baseTime,
			},
			now:         now,
			isHardError: true,
			expected:    SeverityCrit,
		},
		{
			name: "info_single_event",
			cs: state.CountState{
				EventID:   "test.event",
				Count:     1,
				FirstSeen: baseTime,
				LastSeen:  baseTime,
			},
			now:         now,
			isHardError: false,
			expected:    SeverityInfo,
		},
		{
			name: "warn_threshold_reached",
			cs: state.CountState{
				EventID:   "test.event",
				Count:     3, // warn threshold
				FirstSeen: baseTime,
				LastSeen:  baseTime,
			},
			now:         baseTime.Add(5 * time.Minute), // within 10m window
			isHardError: false,
			expected:    SeverityWarn,
		},
		{
			name: "crit_threshold_reached",
			cs: state.CountState{
				EventID:   "test.event",
				Count:     10, // crit threshold
				FirstSeen: baseTime,
				LastSeen:  baseTime,
			},
			now:         baseTime.Add(10 * time.Minute), // within 15m window
			isHardError: false,
			expected:    SeverityCrit,
		},
		{
			name: "warn_window_expired",
			cs: state.CountState{
				EventID:   "test.event",
				Count:     3,
				FirstSeen: baseTime,
				LastSeen:  baseTime,
			},
			now:         baseTime.Add(11 * time.Minute), // after 10m window
			isHardError: false,
			expected:    SeverityInfo, // Reset to INFO
		},
		{
			name: "crit_window_expired",
			cs: state.CountState{
				EventID:   "test.event",
				Count:     10,
				FirstSeen: baseTime,
				LastSeen:  baseTime,
			},
			now:         baseTime.Add(16 * time.Minute), // after 15m window
			isHardError: false,
			expected:    SeverityInfo, // Reset to INFO
		},
		{
			name: "boundary_warn_2_to_3",
			cs: state.CountState{
				EventID:   "test.event",
				Count:     2, // just below warn threshold
				FirstSeen: baseTime,
				LastSeen:  baseTime,
			},
			now:         baseTime.Add(5 * time.Minute),
			isHardError: false,
			expected:    SeverityInfo,
		},
		{
			name: "boundary_crit_9_to_10",
			cs: state.CountState{
				EventID:   "test.event",
				Count:     9, // just below crit threshold
				FirstSeen: baseTime,
				LastSeen:  baseTime,
			},
			now:         baseTime.Add(10 * time.Minute),
			isHardError: false,
			expected:    SeverityWarn, // Still WARN, not CRIT
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := eval.Evaluate(tt.cs, tt.now, tt.isHardError)
			if result != tt.expected {
				t.Errorf("Evaluate() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEvaluator_IsHardError(t *testing.T) {
	cfg := config.DefaultConfig()
	eval := NewEvaluator(cfg)

	tests := []struct {
		name     string
		eventID  string
		expected bool
	}{
		{
			name:     "hard_error_link_down",
			eventID:  "host.network.link_down",
			expected: true,
		},
		{
			name:     "hard_error_kernel_panic",
			eventID:  "host.kernel.panic",
			expected: true,
		},
		{
			name:     "hard_error_container_crash",
			eventID:  "container.lifecycle.crash",
			expected: true,
		},
		{
			name:     "not_hard_error",
			eventID:  "test.event",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := eval.IsHardError(tt.eventID)
			if result != tt.expected {
				t.Errorf("IsHardError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEvaluator_GetThreshold(t *testing.T) {
	cfg := config.DefaultConfig()
	eval := NewEvaluator(cfg)

	if eval.GetThreshold(SeverityInfo) != 1 {
		t.Errorf("GetThreshold(INFO) = %d, want 1", eval.GetThreshold(SeverityInfo))
	}
	if eval.GetThreshold(SeverityWarn) != 3 {
		t.Errorf("GetThreshold(WARN) = %d, want 3", eval.GetThreshold(SeverityWarn))
	}
	if eval.GetThreshold(SeverityCrit) != 10 {
		t.Errorf("GetThreshold(CRIT) = %d, want 10", eval.GetThreshold(SeverityCrit))
	}
}

func TestEvaluator_GetWindow(t *testing.T) {
	cfg := config.DefaultConfig()
	eval := NewEvaluator(cfg)

	if eval.GetWindow(SeverityInfo) != 0 {
		t.Errorf("GetWindow(INFO) = %v, want 0", eval.GetWindow(SeverityInfo))
	}

	warnWindow := eval.GetWindow(SeverityWarn)
	expectedWarn := 10 * time.Minute
	if warnWindow != expectedWarn {
		t.Errorf("GetWindow(WARN) = %v, want %v", warnWindow, expectedWarn)
	}

	critWindow := eval.GetWindow(SeverityCrit)
	expectedCrit := 15 * time.Minute
	if critWindow != expectedCrit {
		t.Errorf("GetWindow(CRIT) = %v, want %v", critWindow, expectedCrit)
	}
}
