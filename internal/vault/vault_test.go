package vault

import (
	"testing"
)

func TestNewFDID(t *testing.T) {
	id := NewFDID("Test Feature", "testuser")
	if len(id) != 7 { // "FD-" + 4 hex chars
		t.Errorf("expected 7 char ID, got %d: %q", len(id), id)
	}
	if id[:3] != "FD-" {
		t.Errorf("expected FD- prefix, got %q", id)
	}
}

func TestNewFDIDUniqueness(t *testing.T) {
	// Two calls with same input should produce different IDs (timestamp differs).
	id1 := NewFDID("Same Title", "same-author")
	id2 := NewFDID("Same Title", "same-author")
	if id1 == id2 {
		t.Errorf("expected unique IDs, both got %q", id1)
	}
}

func TestFDStatusConstants(t *testing.T) {
	statuses := []FDStatus{
		FDPlanned, FDApproved, FDInProgress,
		FDComplete, FDClosed, FDRejected, FDAbandoned,
	}
	for _, s := range statuses {
		if s == "" {
			t.Error("empty FD status constant")
		}
	}
}

func TestSDDStatusConstants(t *testing.T) {
	statuses := []SDDStatus{
		SDDPlanned, SDDValidated, SDDInProgress, SDDDone, SDDFailed, SDDCancelled,
	}
	for _, s := range statuses {
		if s == "" {
			t.Error("empty SDD status constant")
		}
	}
}
