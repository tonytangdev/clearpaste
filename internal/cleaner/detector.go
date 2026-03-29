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
	return hasBrokenLineWraps(text)
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
