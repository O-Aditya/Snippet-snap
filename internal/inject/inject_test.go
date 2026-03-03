package inject

import (
	"os"
	"reflect"
	"testing"
)

func TestFindVars(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect []string
	}{
		{
			name:   "no vars",
			input:  "no vars here",
			expect: nil,
		},
		{
			name:   "single var",
			input:  "echo {{NAME}}",
			expect: []string{"NAME"},
		},
		{
			name:   "two vars",
			input:  "{{HOST}}:{{PORT}}",
			expect: []string{"HOST", "PORT"},
		},
		{
			name:   "deduped and ordered",
			input:  "{{A}} {{A}} {{B}}",
			expect: []string{"A", "B"},
		},
		{
			name:   "lowercase not matched",
			input:  "{{lowercase}}",
			expect: nil,
		},
		{
			name:   "mixed valid and invalid",
			input:  "{{A1}} {{B_C}} {{123}}",
			expect: []string{"A1", "B_C"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindVars(tt.input)
			if !reflect.DeepEqual(got, tt.expect) {
				t.Errorf("FindVars(%q) = %v, want %v", tt.input, got, tt.expect)
			}
		})
	}
}

func TestResolveVars_NoVars(t *testing.T) {
	input := "hello world — no placeholders"
	got, err := ResolveVars(input)
	if err != nil {
		t.Fatalf("ResolveVars() unexpected error: %v", err)
	}
	if got != input {
		t.Errorf("ResolveVars() = %q, want %q", got, input)
	}
}

func TestResolveVars_WithVars(t *testing.T) {
	// Create a pipe to simulate stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error: %v", err)
	}

	// Write simulated input before calling ResolveVars
	_, err = w.WriteString("myhost\n8080\n")
	if err != nil {
		t.Fatalf("write to pipe: %v", err)
	}
	w.Close()

	// Replace stdin with our pipe
	origStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = origStdin }()

	input := "curl https://{{HOST}}:{{PORT}}/api"
	got, resolveErr := ResolveVars(input)
	if resolveErr != nil {
		t.Fatalf("ResolveVars() unexpected error: %v", resolveErr)
	}

	expected := "curl https://myhost:8080/api"
	if got != expected {
		t.Errorf("ResolveVars() = %q, want %q", got, expected)
	}
}
