package runner

import (
	"testing"
)

func TestParseDryRunReport_GO(t *testing.T) {
	t.Parallel()

	report := `=== SDD Dry Run: SDD-001 ===

  Feasibility: GO
  Confidence: 92%
  Complexity: low

  BLOCKERS (must fix before execution):
    (none)

  WARNINGS (review recommended):
    (none)

  PASS:
    v [Pass 1] All preconditions met
    v [Pass 2] No guardrail conflicts

  RECOMMENDATION:
    Esecuzione sicura, nessun blocco rilevato.
`

	result := ParseDryRunReport(report)

	if result.Feasibility != FeasibilityGo {
		t.Errorf("expected GO, got %s", result.Feasibility)
	}
	if result.Confidence != "92%" {
		t.Errorf("expected confidence 92%%, got %s", result.Confidence)
	}
	if result.Complexity != "low" {
		t.Errorf("expected complexity low, got %s", result.Complexity)
	}
	if !result.IsGo() {
		t.Error("expected IsGo() to be true for GO")
	}
	if len(result.Blockers) != 0 {
		t.Errorf("expected 0 blockers, got %d", len(result.Blockers))
	}
}

func TestParseDryRunReport_NOGO(t *testing.T) {
	t.Parallel()

	report := `=== SDD Dry Run: SDD-002 ===

  Feasibility: NO-GO
  Confidence: 25%
  Complexity: very-high

  BLOCKERS (must fix before execution):
    x [Pass 1] Missing context file: internal/auth/handler.go
    x [Pass 2] Guardrail conflict: SDD requires reading .env

  WARNINGS (review recommended):
    ! [Pass 4] Cross-SDD dependency on SDD-001 not yet complete

  RECOMMENDATION:
    Blocchi critici impediscono l'esecuzione. Risolvere prima i blocchi.
`

	result := ParseDryRunReport(report)

	if result.Feasibility != FeasibilityNoGo {
		t.Errorf("expected NO-GO, got %s", result.Feasibility)
	}
	if result.Confidence != "25%" {
		t.Errorf("expected confidence 25%%, got %s", result.Confidence)
	}
	if result.Complexity != "very-high" {
		t.Errorf("expected complexity very-high, got %s", result.Complexity)
	}
	if result.IsGo() {
		t.Error("expected IsGo() to be false for NO-GO")
	}
	if len(result.Blockers) != 2 {
		t.Errorf("expected 2 blockers, got %d", len(result.Blockers))
	}
	if len(result.Warnings) != 1 {
		t.Errorf("expected 1 warning, got %d", len(result.Warnings))
	}
}

func TestParseDryRunReport_CAUTION(t *testing.T) {
	t.Parallel()

	report := `=== SDD Dry Run: SDD-003 ===

  Feasibility: CAUTION
  Confidence: 65%
  Complexity: high

  BLOCKERS (must fix before execution):
    (none)

  WARNINGS (review recommended):
    ! [Pass 3] Ambiguous scope: "configure database" could require .env access
    ! [Pass 4] Cross-SDD dependency on SDD-001
    ! [Pass 4] 8 files to modify — high complexity

  RECOMMENDATION:
    Procedere con cautela. Verificare manualmente le dipendenze.
`

	result := ParseDryRunReport(report)

	if result.Feasibility != FeasibilityCaution {
		t.Errorf("expected CAUTION, got %s", result.Feasibility)
	}
	if result.Confidence != "65%" {
		t.Errorf("expected confidence 65%%, got %s", result.Confidence)
	}
	if result.Complexity != "high" {
		t.Errorf("expected complexity high, got %s", result.Complexity)
	}
	if !result.IsGo() {
		t.Error("expected IsGo() to be true for CAUTION")
	}
	if len(result.Blockers) != 0 {
		t.Errorf("expected 0 blockers, got %d", len(result.Blockers))
	}
	if len(result.Warnings) != 3 {
		t.Errorf("expected 3 warnings, got %d", len(result.Warnings))
	}
}

func TestParseDryRunReport_UnparseableDefaults(t *testing.T) {
	t.Parallel()

	result := ParseDryRunReport("no structured output here")

	if result.Feasibility != FeasibilityNoGo {
		t.Errorf("expected default NO-GO for unparseable report, got %s", result.Feasibility)
	}
	if result.IsGo() {
		t.Error("expected IsGo() to be false for unparseable report")
	}
}
