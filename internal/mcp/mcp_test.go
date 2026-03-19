package mcp

import (
	"testing"
)

func TestToolName(t *testing.T) {
	tests := []struct {
		namespace string
		tool      string
		want      string
	}{
		{"code", "search_graph", "forgia_code_search_graph"},
		{"code", "detect_changes", "forgia_code_detect_changes"},
		{"custom", "my_tool", "forgia_custom_my_tool"},
	}

	for _, tt := range tests {
		got := ToolName(tt.namespace, tt.tool)
		if got != tt.want {
			t.Errorf("ToolName(%q, %q) = %q, want %q", tt.namespace, tt.tool, got, tt.want)
		}
	}
}

func TestParseNamespacedTool(t *testing.T) {
	tests := []struct {
		tool      string
		wantNS    string
		wantName  string
		wantError bool
	}{
		{"forgia_code_search_graph", "code", "search_graph", false},
		{"forgia_code_detect_changes", "code", "detect_changes", false},
		{"forgia_custom_my_tool", "custom", "my_tool", false},
		{"forgia_ns_a_b_c", "ns", "a_b_c", false},

		// Error cases.
		{"search_graph", "", "", true},           // missing forgia_ prefix
		{"forgia_nounderscoreafter", "", "", true}, // no separator after namespace
		{"", "", "", true},                        // empty
		{"other_prefix_tool", "", "", true},       // wrong prefix
	}

	for _, tt := range tests {
		ns, name, err := parseNamespacedTool(tt.tool)
		if tt.wantError {
			if err == nil {
				t.Errorf("parseNamespacedTool(%q) expected error, got ns=%q name=%q", tt.tool, ns, name)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseNamespacedTool(%q) unexpected error: %v", tt.tool, err)
			continue
		}
		if ns != tt.wantNS || name != tt.wantName {
			t.Errorf("parseNamespacedTool(%q) = (%q, %q), want (%q, %q)", tt.tool, ns, name, tt.wantNS, tt.wantName)
		}
	}
}
