package cleaner

import (
	"strings"
	"unicode"
)

// NeedsCleaning returns true if text contains Unicode Box Drawing (U+2500–U+257F)
// or Block Elements (U+2580–U+259F) characters, or has broken line wraps
// (lines ending mid-sentence followed by continuation text).
func NeedsCleaning(text string) bool {
	for _, r := range text {
		if (r >= 0x2500 && r <= 0x257F) || (r >= 0x2580 && r <= 0x259F) {
			return true
		}
	}
	return hasBrokenLineWraps(text) || hasUniformIndent(text)
}

// hasUniformIndent detects text where every non-empty line shares the same
// leading whitespace prefix (≥2 spaces). This pattern is an artifact of
// copying from TUI tools (e.g. Claude Code quote blocks) where the block
// character may not survive clipboard copy.
func hasUniformIndent(text string) bool {
	lines := strings.Split(text, "\n")

	// Find common leading whitespace prefix across non-empty lines
	prefix := ""
	nonEmpty := 0
	proseLines := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		nonEmpty++

		// Count leading spaces
		trimmed := strings.TrimLeft(line, " ")
		indent := line[:len(line)-len(trimmed)]
		if len(indent) < 2 {
			return false // a non-empty line without indent → not uniform
		}

		if prefix == "" {
			prefix = indent
		} else if indent != prefix {
			return false // inconsistent indent
		}

		// Check if line content (after indent) looks like prose (4+ words)
		words := strings.Fields(trimmed)
		if len(words) >= 4 {
			proseLines++
		}
	}

	// Need at least 3 non-empty lines and some prose content
	return nonEmpty >= 3 && proseLines >= 1
}

// hasBrokenLineWraps detects text that was hard-wrapped at a column width.
// Returns true if at least one line ends without terminal punctuation and
// the next non-empty line starts with a lowercase letter.
func hasBrokenLineWraps(text string) bool {
	lines := strings.Split(text, "\n")
	if len(lines) < 2 {
		return false
	}

	for i := 0; i < len(lines)-1; i++ {
		trimmed := strings.TrimSpace(lines[i])
		if len(trimmed) == 0 {
			continue
		}

		lastChar := rune(trimmed[len(trimmed)-1])
		if lastChar == '.' || lastChar == '!' || lastChar == '?' || lastChar == ':' || lastChar == ',' {
			continue
		}

		// Find next non-empty line
		for j := i + 1; j < len(lines); j++ {
			next := strings.TrimSpace(lines[j])
			if len(next) == 0 {
				break // paragraph break, not a broken wrap
			}
			firstRune := []rune(next)[0]
			if unicode.IsLower(firstRune) {
				return true
			}
			break
		}
	}
	return false
}
