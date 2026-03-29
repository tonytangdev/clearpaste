package cleaner

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	boxCharsRe       = regexp.MustCompile(`[\x{2500}-\x{257F}\x{2580}-\x{259F}]`)
	boxAndPipeRe     = regexp.MustCompile(`[\x{2500}-\x{257F}\x{2580}-\x{259F}|]`)
	multiSpaceRe     = regexp.MustCompile(` {2,}`)
	excessNewlinesRe = regexp.MustCompile(`\n{3,}`)
	numberedListRe   = regexp.MustCompile(`^\d+\.`)
	codeLikeCharsRe  = regexp.MustCompile(`[{}();=]`)
)

const maxSize = 1024 * 1024 // 1MB

// Clean applies the full cleaning pipeline to text.
func Clean(text string) string {
	if len(text) > maxSize {
		return text
	}

	lines := strings.Split(text, "\n")

	// Step 1: strip box-drawing chars (and | on lines that contain box chars)
	for i, line := range lines {
		if boxCharsRe.MatchString(line) {
			lines[i] = boxAndPipeRe.ReplaceAllString(line, "")
		}
	}

	// Detect code blocks before collapsing spaces (indentation is significant)
	codeBlock := make([]bool, len(lines))
	for i, line := range lines {
		codeBlock[i] = isCodeBlock(line)
	}

	// Step 2: collapse multiple spaces; preserve leading spaces on code block lines
	for i, line := range lines {
		if codeBlock[i] {
			// Only collapse internal spaces, not leading whitespace
			trimmed := strings.TrimLeft(line, " \t")
			leading := line[:len(line)-len(trimmed)]
			lines[i] = leading + multiSpaceRe.ReplaceAllString(trimmed, " ")
		} else {
			lines[i] = multiSpaceRe.ReplaceAllString(line, " ")
		}
	}

	// Step 3: rejoin broken lines
	var result []string
	var resultIsCode []bool
	for i, line := range lines {
		if i == 0 {
			result = append(result, strings.TrimSpace(line))
			resultIsCode = append(resultIsCode, false)
			continue
		}

		if shouldKeepSeparate(lines, codeBlock, i) {
			// Preserve leading whitespace for code blocks
			if codeBlock[i] {
				result = append(result, trimBoxLeading(line))
				resultIsCode = append(resultIsCode, true)
			} else {
				result = append(result, strings.TrimSpace(line))
				resultIsCode = append(resultIsCode, false)
			}
			continue
		}

		// Merge with previous line
		result[len(result)-1] = result[len(result)-1] + " " + strings.TrimSpace(line)
	}

	// Step 4: trim non-code lines
	for i, line := range result {
		if !resultIsCode[i] {
			result[i] = strings.TrimSpace(line)
		}
	}

	// Step 5: collapse excessive newlines (3+ → 2)
	joined := strings.Join(result, "\n")
	joined = excessNewlinesRe.ReplaceAllString(joined, "\n\n")

	return strings.TrimSpace(joined)
}

// shouldKeepSeparate returns true if lines[i] should NOT be merged with the previous line.
func shouldKeepSeparate(lines []string, codeBlock []bool, i int) bool {
	line := lines[i]
	prevLine := ""
	if i > 0 {
		prevLine = strings.TrimSpace(lines[i-1])
	}

	// Empty line = paragraph break
	if strings.TrimSpace(line) == "" {
		return true
	}

	// Previous line was empty = paragraph break, keep separate
	if i > 0 && strings.TrimSpace(lines[i-1]) == "" {
		return true
	}

	trimmed := strings.TrimSpace(line)

	// List markers: -, *, • (followed by space)
	if len(trimmed) > 1 && (trimmed[0] == '-' || trimmed[0] == '*') && trimmed[1] == ' ' {
		return true
	}
	if strings.HasPrefix(trimmed, "• ") {
		return true
	}

	// Numbered list: 1., 2., etc.
	if numberedListRe.MatchString(trimmed) {
		return true
	}

	// Emoji at start
	if startsWithEmoji(trimmed) {
		return true
	}

	// Previous line ends with sentence-ending punctuation AND this line starts with uppercase
	if len(prevLine) > 0 && len(trimmed) > 0 {
		lastChar := rune(prevLine[len(prevLine)-1])
		firstRune := firstRuneOf(trimmed)
		if (lastChar == '.' || lastChar == '!' || lastChar == '?') && unicode.IsUpper(firstRune) {
			return true
		}
	}

	// Code block: starts with 4+ spaces or tab (use pre-computed flag)
	if codeBlock[i] {
		return true
	}

	// Code-like: 3+ occurrences of {}();=
	if len(codeLikeCharsRe.FindAllString(trimmed, -1)) >= 3 {
		return true
	}

	return false
}

func isCodeBlock(line string) bool {
	return strings.HasPrefix(line, "    ") || strings.HasPrefix(line, "\t")
}

func trimBoxLeading(line string) string {
	stripped := strings.TrimLeft(line, " \t")
	indent := len(line) - len(stripped)
	preserved := (indent / 4) * 4
	if preserved < 4 {
		preserved = 4
	}
	if preserved > indent {
		preserved = indent
	}
	return strings.Repeat(" ", preserved) + stripped
}

func firstRuneOf(s string) rune {
	for _, r := range s {
		return r
	}
	return 0
}

func startsWithEmoji(s string) bool {
	for _, r := range s {
		if (r >= 0x2600 && r <= 0x27BF) ||
			(r >= 0x1F300 && r <= 0x1F9FF) ||
			(r >= 0x1FA00 && r <= 0x1FA6F) ||
			(r >= 0x1FA70 && r <= 0x1FAFF) {
			return true
		}
		return false
	}
	return false
}
