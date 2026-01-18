package shortcodes

import (
	"testing"

	"chameth.com/chameth.com/cmd/serve/content/shortcodes/context"
)

func TestSplitArguments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
		wantErr  bool
	}{
		{
			name:     "single word",
			input:    "hello",
			expected: []string{"hello"},
			wantErr:  false,
		},
		{
			name:     "multiple words",
			input:    "hello world test",
			expected: []string{"hello", "world", "test"},
			wantErr:  false,
		},
		{
			name:     "quoted string",
			input:    `"hello world" test`,
			expected: []string{"hello world", "test"},
			wantErr:  false,
		},
		{
			name:     "escaped quote",
			input:    `"hello \"world\" test"`,
			expected: []string{`hello "world" test`},
			wantErr:  false,
		},
		{
			name:     "double escaped quote",
			input:    `"hello \\"world\\" test"`,
			expected: []string{`hello "world" test`},
			wantErr:  false,
		},
		{
			name:     "mixed",
			input:    `arg1 "argument 2" arg3`,
			expected: []string{"arg1", "argument 2", "arg3"},
			wantErr:  false,
		},
		{
			name:     "unclosed quote",
			input:    `"hello world`,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := splitArguments(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("splitArguments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(result) != len(tt.expected) {
					t.Errorf("splitArguments() returned %d args, expected %d", len(result), len(tt.expected))
					return
				}
				for i := range tt.expected {
					if result[i] != tt.expected[i] {
						t.Errorf("splitArguments()[%d] = %q, want %q", i, result[i], tt.expected[i])
					}
				}
			}
		})
	}
}

func TestRenderUnknownShortcode(t *testing.T) {
	input := "prefix {%unknown%} suffix"
	result := Render(input, &context.Context{})
	expected := "prefix " + shortcodesError + " suffix"
	if result != expected {
		t.Errorf("Render() = %q, want %q (unknown shortcode should be replaced with error)", result, expected)
	}
}

func TestRenderEndTagWhitespace(t *testing.T) {
	input := "prefix {%warning%} content {%end warning%} suffix"
	result := Render(input, &context.Context{})
	if result == input {
		t.Errorf("Render() returned unchanged input, end tag with whitespace not recognized")
	}
}
