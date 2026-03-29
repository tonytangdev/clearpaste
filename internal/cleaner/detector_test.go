package cleaner

import "testing"

func TestNeedsCleaning_BoxDrawingChars(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"light vertical │", "│ hello world", true},
		{"heavy vertical ┃", "┃ some text", true},
		{"corner ┌", "┌──────┐", true},
		{"corner └", "└──────┘", true},
		{"tee ├", "├── item", true},
		{"cross ┼", "──┼──", true},
		{"block element ▌", "▌text", true},
		{"mixed with normal text", "hello │ world", true},
		{"normal text no box chars", "hello world", false},
		{"empty string", "", false},
		{"ascii pipe only", "cat foo | grep bar", false},
		{"markdown table", "| col1 | col2 |", false},
		{"code with pipe", "if a || b {}", false},
		{"newlines only", "\n\n\n", false},
		{"broken line wrap", "The editing part is harder than people think. You need to remember why you\n   made certain choices", true},
		{"intentional line break after period", "First sentence.\nSecond sentence.", false},
		{"single line", "Just a single line of text", false},
		{"paragraph break not broken wrap", "First paragraph.\n\nSecond paragraph.", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NeedsCleaning(tt.input)
			if got != tt.want {
				t.Errorf("NeedsCleaning(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
