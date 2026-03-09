package query

import (
	"strings"
	"testing"
)

func TestEvaluateFormula(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		vars    map[string]string
		want    string
		wantErr string
	}{
		{
			name: "simple variable substitution",
			expr: "a + b",
			vars: map[string]string{"a": "sel_a", "b": "sel_b"},
			want: "(sel_a) + (sel_b)",
		},
		{
			name: "formula with constant",
			expr: "a * 100",
			vars: map[string]string{"a": "sel_a"},
			want: "(sel_a) * 100",
		},
		{
			name: "formula with parentheses",
			expr: "(a + b) / c",
			vars: map[string]string{"a": "sel_a", "b": "sel_b", "c": "sel_c"},
			want: "((sel_a) + (sel_b)) / (sel_c)",
		},
		{
			name: "single variable reference",
			expr: "a",
			vars: map[string]string{"a": "sel_a"},
			want: "(sel_a)",
		},
		{
			name:    "unknown variable returns error",
			expr:    "a + z",
			vars:    map[string]string{"a": "sel_a"},
			wantErr: "unknown variable",
		},
		{
			name:    "empty expression returns error",
			expr:    "",
			vars:    map[string]string{"a": "sel_a"},
			wantErr: "empty formula",
		},
		{
			name: "formula with division",
			expr: "a / b",
			vars: map[string]string{"a": "sel_a", "b": "sel_b"},
			want: "(sel_a) / (sel_b)",
		},
		{
			name: "complex nested formula",
			expr: "(a + b) / (c - d) * 100",
			vars: map[string]string{"a": "sel_a", "b": "sel_b", "c": "sel_c", "d": "sel_d"},
			want: "((sel_a) + (sel_b)) / ((sel_c) - (sel_d)) * 100",
		},
		{
			name: "multi-character variable names",
			expr: "query0 + query1",
			vars: map[string]string{"query0": "sel0", "query1": "sel1"},
			want: "(sel0) + (sel1)",
		},
		{
			name: "mixed single and multi-char variables",
			expr: "a + query0",
			vars: map[string]string{"a": "sel_a", "query0": "sel0"},
			want: "(sel_a) + (sel0)",
		},
		{
			name: "underscore in variable name",
			expr: "cpu_total / cpu_idle",
			vars: map[string]string{"cpu_total": "total_sel", "cpu_idle": "idle_sel"},
			want: "(total_sel) / (idle_sel)",
		},
		{
			name: "multi-char with constant",
			expr: "query0 * 100",
			vars: map[string]string{"query0": "sel0"},
			want: "(sel0) * 100",
		},
		{
			name:    "unknown multi-char variable",
			expr:    "query0 + query2",
			vars:    map[string]string{"query0": "sel0"},
			wantErr: "unknown variable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EvaluateFormula(tt.expr, tt.vars)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("EvaluateFormula(%q) = %q, want %q", tt.expr, got, tt.want)
			}
		})
	}
}
