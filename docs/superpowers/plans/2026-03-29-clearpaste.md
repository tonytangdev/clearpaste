# ClearPaste Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a cross-platform system tray app that monitors the clipboard and auto-cleans terminal formatting artifacts (box-drawing chars, broken line wraps).

**Architecture:** Go binary with 3 internal packages: `cleaner` (pure text transforms), `clipboard` (read/write with polling monitor), `tray` (system tray UI). `main.go` wires them together. TDD — cleaner is tested thoroughly, clipboard/tray are integration-tested manually.

**Tech Stack:** Go 1.22+, `fyne.io/systray` (tray), `github.com/atotto/clipboard` (clipboard read/write — simpler than `golang.design/x/clipboard` since we only need text + polling), GoReleaser (build/release).

**Spec:** `docs/superpowers/specs/2026-03-29-clearpaste-design.md`

---

## File Structure

```
clearpaste/
├── cmd/clearpaste/main.go           — entry point, wires packages together
├── internal/cleaner/detector.go     — NeedsCleaning() function
├── internal/cleaner/detector_test.go
├── internal/cleaner/cleaner.go      — Clean() pipeline
├── internal/cleaner/cleaner_test.go
├── internal/clipboard/clipboard.go  — Reader/Writer interfaces + impl
├── internal/clipboard/monitor.go    — polling loop, loop prevention, undo
├── internal/tray/tray.go            — system tray setup, menu, icon states
├── internal/tray/icons.go           — embedded icon assets via go:embed
├── icons/                           — icon PNGs at repo root (for go:embed)
│   ├── icon.png                     — normal tray icon (22x22)
│   ├── icon_active.png              — "just cleaned" flash icon
│   └── icon_disabled.png            — disabled/grayed icon
├── scripts/gen_icons.go             — generates placeholder icons
├── go.mod
├── go.sum
├── Makefile
├── .goreleaser.yml
└── .github/workflows/release.yml
```

---

## Task 1: Project Scaffolding

**Files:**
- Create: `go.mod`
- Create: `cmd/clearpaste/main.go`
- Create: `Makefile`

- [ ] **Step 1: Initialize Go module**

```bash
cd /Users/tonytang/Documents/github/tonytangdev/clean-copy-terminal
go mod init github.com/tonytangdev/clearpaste
```

- [ ] **Step 2: Create minimal main.go**

Create `cmd/clearpaste/main.go`:

```go
package main

import "fmt"

func main() {
	fmt.Println("ClearPaste starting...")
}
```

- [ ] **Step 3: Create Makefile**

Create `Makefile`:

```makefile
.PHONY: build run test clean

build:
	go build -o bin/clearpaste ./cmd/clearpaste

run:
	go run ./cmd/clearpaste

test:
	go test ./... -v

clean:
	rm -rf bin/
```

- [ ] **Step 4: Verify it builds and runs**

```bash
make build && ./bin/clearpaste
```

Expected: prints "ClearPaste starting..."

- [ ] **Step 5: Commit**

```bash
git init
git add go.mod cmd/clearpaste/main.go Makefile
git commit -m "init: scaffold go project"
```

---

## Task 2: Detector — `NeedsCleaning()`

**Files:**
- Create: `internal/cleaner/detector.go`
- Create: `internal/cleaner/detector_test.go`

- [ ] **Step 1: Write failing tests for NeedsCleaning**

Create `internal/cleaner/detector_test.go`:

```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
make test
```

Expected: FAIL — `NeedsCleaning` not defined.

- [ ] **Step 3: Implement NeedsCleaning**

Create `internal/cleaner/detector.go`:

```go
package cleaner

// NeedsCleaning returns true if text contains Unicode Box Drawing (U+2500–U+257F)
// or Block Elements (U+2580–U+259F) characters.
// Does NOT trigger on ASCII pipe '|' alone.
func NeedsCleaning(text string) bool {
	for _, r := range text {
		if (r >= 0x2500 && r <= 0x257F) || (r >= 0x2580 && r <= 0x259F) {
			return true
		}
	}
	return false
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
make test
```

Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/cleaner/detector.go internal/cleaner/detector_test.go
git commit -m "feat: add NeedsCleaning detector for box-drawing chars"
```

---

## Task 3: Cleaner — `Clean()` Pipeline

**Files:**
- Create: `internal/cleaner/cleaner.go`
- Create: `internal/cleaner/cleaner_test.go`

- [ ] **Step 1: Write failing tests for Clean**

Create `internal/cleaner/cleaner_test.go`:

```go
package cleaner

import "testing"

func TestClean_StripBoxChars(t *testing.T) {
	input := "│ hello world │"
	want := "hello world"
	got := Clean(input)
	if got != want {
		t.Errorf("Clean(%q) = %q, want %q", input, got, want)
	}
}

func TestClean_StripPipeOnlyWhenBoxCharsPresent(t *testing.T) {
	// Line has both │ and |, so | should be stripped too
	input := "│ hello | world │"
	want := "hello world"
	got := Clean(input)
	if got != want {
		t.Errorf("Clean(%q) = %q, want %q", input, got, want)
	}
}

func TestClean_PreservePipeOnLinesWithoutBoxChars(t *testing.T) {
	// First line has box chars, second line has only ascii pipe
	input := "│ hello\nfoo | bar"
	want := "hello foo | bar"
	got := Clean(input)
	if got != want {
		t.Errorf("Clean(%q) = %q, want %q", input, got, want)
	}
}

func TestClean_CollapseMultipleSpaces(t *testing.T) {
	input := "│ hello    world │"
	want := "hello world"
	got := Clean(input)
	if got != want {
		t.Errorf("Clean(%q) = %q, want %q", input, got, want)
	}
}

func TestClean_RejoinBrokenLines(t *testing.T) {
	input := "│ This is a long sentence that was\n│ broken across two lines"
	want := "This is a long sentence that was broken across two lines"
	got := Clean(input)
	if got != want {
		t.Errorf("Clean(%q) = %q, want %q", input, got, want)
	}
}

func TestClean_PreserveListMarkers(t *testing.T) {
	input := "│ Introduction text\n│ - item one\n│ - item two\n│ * star item"
	want := "Introduction text\n- item one\n- item two\n* star item"
	got := Clean(input)
	if got != want {
		t.Errorf("Clean(%q) = %q, want %q", input, got, want)
	}
}

func TestClean_PreserveNumberedLists(t *testing.T) {
	input := "│ Steps:\n│ 1. first\n│ 2. second"
	want := "Steps:\n1. first\n2. second"
	got := Clean(input)
	if got != want {
		t.Errorf("Clean(%q) = %q, want %q", input, got, want)
	}
}

func TestClean_PreserveParagraphBreaks(t *testing.T) {
	input := "│ First paragraph.\n│\n│ Second paragraph."
	want := "First paragraph.\n\nSecond paragraph."
	got := Clean(input)
	if got != want {
		t.Errorf("Clean(%q) = %q, want %q", input, got, want)
	}
}

func TestClean_PreserveNewSentenceAfterPeriod(t *testing.T) {
	input := "│ End of sentence.\n│ Start of new one."
	want := "End of sentence.\nStart of new one."
	got := Clean(input)
	if got != want {
		t.Errorf("Clean(%q) = %q, want %q", input, got, want)
	}
}

func TestClean_MergeAfterAbbreviation(t *testing.T) {
	// "e.g." followed by lowercase = not sentence ending, merge
	input := "│ Use tools e.g.\n│ hammer or wrench"
	want := "Use tools e.g. hammer or wrench"
	got := Clean(input)
	if got != want {
		t.Errorf("Clean(%q) = %q, want %q", input, got, want)
	}
}

func TestClean_PreserveCodeBlocks(t *testing.T) {
	input := "│ Example:\n│     func main() {\n│         fmt.Println()\n│     }"
	want := "Example:\n    func main() {\n        fmt.Println()\n    }"
	got := Clean(input)
	if got != want {
		t.Errorf("Clean(%q) = %q, want %q", input, got, want)
	}
}

func TestClean_PreserveCodeLikeLines(t *testing.T) {
	input := "│ Some text\n│ if (a && b) { return c; }"
	want := "Some text\nif (a && b) { return c; }"
	got := Clean(input)
	if got != want {
		t.Errorf("Clean(%q) = %q, want %q", input, got, want)
	}
}

func TestClean_CollapseExcessiveNewlines(t *testing.T) {
	input := "│ First\n│\n│\n│\n│ Second"
	want := "First\n\nSecond"
	got := Clean(input)
	if got != want {
		t.Errorf("Clean(%q) = %q, want %q", input, got, want)
	}
}

func TestClean_RealClaudeCodeOutput(t *testing.T) {
	input := `  │ Interesting framing. If agents are making stack decisions, the question
  │ becomes: what context are they pulling from? Right now it's mostly training
  │ data + whatever's in CLAUDE.md. Teams that structure their decision history
  │ so agents can query it will have a real edge here.`
	want := "Interesting framing. If agents are making stack decisions, the question becomes: what context are they pulling from? Right now it's mostly training data + whatever's in CLAUDE.md. Teams that structure their decision history so agents can query it will have a real edge here."
	got := Clean(input)
	if got != want {
		t.Errorf("Clean() =\n%q\nwant:\n%q", got, want)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
make test
```

Expected: FAIL — `Clean` not defined.

- [ ] **Step 3: Implement Clean pipeline**

Create `internal/cleaner/cleaner.go`:

```go
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

	// Step 2: collapse multiple spaces (but keep leading spaces for code-block detection)
	for i, line := range lines {
		lines[i] = multiSpaceRe.ReplaceAllString(line, " ")
	}

	// Step 3: rejoin broken lines (uses untrimmed lines to detect code blocks via leading spaces)
	var result []string
	for i, line := range lines {
		if i == 0 {
			result = append(result, strings.TrimSpace(line))
			continue
		}

		if shouldKeepSeparate(lines, i) {
			// Preserve leading whitespace for code blocks
			if isCodeBlock(line) {
				result = append(result, trimBoxLeading(line))
			} else {
				result = append(result, strings.TrimSpace(line))
			}
			continue
		}

		// Merge with previous line
		result[len(result)-1] = result[len(result)-1] + " " + strings.TrimSpace(line)
	}

	// Step 4: trim each line again (after merging)
	for i, line := range result {
		result[i] = strings.TrimSpace(line)
	}

	// Step 5: collapse excessive newlines (3+ → 2)
	joined := strings.Join(result, "\n")
	joined = excessNewlinesRe.ReplaceAllString(joined, "\n\n")

	return strings.TrimSpace(joined)
}

// shouldKeepSeparate returns true if lines[i] should NOT be merged with the previous line.
// Note: lines have NOT been trimmed yet — leading whitespace is preserved for code-block detection.
func shouldKeepSeparate(lines []string, i int) bool {
	line := lines[i]
	prevLine := ""
	if i > 0 {
		prevLine = strings.TrimSpace(lines[i-1])
	}

	// Empty line = paragraph break
	if strings.TrimSpace(line) == "" {
		return true
	}

	trimmed := strings.TrimSpace(line)

	// List markers: -, *, • (followed by space to avoid false positives like *emphasis*)
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

	// Emoji at start (common Unicode emoji ranges)
	if startsWithEmoji(trimmed) {
		return true
	}

	// Previous line ends with sentence-ending punctuation AND this line starts with uppercase
	// = new sentence/paragraph. But NOT if this line starts with lowercase (abbreviation case).
	if len(prevLine) > 0 && len(trimmed) > 0 {
		lastChar := rune(prevLine[len(prevLine)-1])
		firstRune := firstRuneOf(trimmed)
		if (lastChar == '.' || lastChar == '!' || lastChar == '?') && unicode.IsUpper(firstRune) {
			return true
		}
	}

	// Code block: starts with 4+ spaces or tab (using UNTRIMMED line)
	if isCodeBlock(line) {
		return true
	}

	// Code-like: 3+ occurrences of {}();=
	if len(codeLikeCharsRe.FindAllString(trimmed, -1)) >= 3 {
		return true
	}

	return false
}

// isCodeBlock returns true if the line starts with 4+ spaces or a tab.
func isCodeBlock(line string) bool {
	return strings.HasPrefix(line, "    ") || strings.HasPrefix(line, "\t")
}

// trimBoxLeading trims only the leading box-artifact whitespace but preserves
// meaningful indentation for code blocks.
func trimBoxLeading(line string) string {
	// Find the first non-space character to measure indentation
	stripped := strings.TrimLeft(line, " \t")
	indent := len(line) - len(stripped)
	// Keep at most the indentation minus any leading box artifact padding (typically 2-4 chars)
	// Heuristic: preserve indentation in multiples of 4
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
		// Miscellaneous Symbols, Dingbats, Emoticons, Transport/Map,
		// Supplemental Symbols, Playing Cards, etc.
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
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
make test
```

Expected: all PASS. If any fail, read the error, fix the specific transform, re-run.

- [ ] **Step 5: Commit**

```bash
git add internal/cleaner/cleaner.go internal/cleaner/cleaner_test.go
git commit -m "feat: add Clean pipeline with line rejoining and code preservation"
```

---

## Task 4: Clipboard Interface & Implementation

**Files:**
- Create: `internal/clipboard/clipboard.go`

- [ ] **Step 1: Install atotto/clipboard dependency**

```bash
go get github.com/atotto/clipboard
```

- [ ] **Step 2: Create clipboard interface and implementation**

Create `internal/clipboard/clipboard.go`:

```go
package clipboard

import atotto "github.com/atotto/clipboard"

// Reader reads text from the system clipboard.
type Reader interface {
	Read() (string, error)
}

// Writer writes text to the system clipboard.
type Writer interface {
	Write(text string) error
}

// System implements Reader and Writer using the real system clipboard.
type System struct{}

func (s *System) Read() (string, error) {
	return atotto.ReadAll()
}

func (s *System) Write(text string) error {
	return atotto.WriteAll(text)
}
```

- [ ] **Step 3: Verify it compiles**

```bash
go build ./internal/clipboard/
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add internal/clipboard/clipboard.go go.mod go.sum
git commit -m "feat: add clipboard reader/writer interface with atotto impl"
```

---

## Task 5: Clipboard Monitor

**Files:**
- Create: `internal/clipboard/monitor.go`

- [ ] **Step 1: Implement the monitor**

Create `internal/clipboard/monitor.go`:

```go
package clipboard

import (
	"crypto/sha256"
	"sync"
	"time"

	"github.com/tonytangdev/clearpaste/internal/cleaner"
)

const pollInterval = 300 * time.Millisecond

// Monitor polls the clipboard for changes and auto-cleans terminal text.
type Monitor struct {
	reader   Reader
	writer   Writer
	enabled  bool
	mu       sync.Mutex
	lastHash [32]byte
	skipNext bool

	// Undo state
	originalText string
	hasOriginal  bool

	// Callbacks
	OnCleaned func() // called after text is cleaned
	OnUndo    func() // called after undo
}

// NewMonitor creates a monitor with the given clipboard reader/writer.
func NewMonitor(reader Reader, writer Writer) *Monitor {
	return &Monitor{
		reader:  reader,
		writer:  writer,
		enabled: true,
	}
}

// Start begins polling the clipboard. Blocks until stop is called.
func (m *Monitor) Start(stop <-chan struct{}) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	// Initialize hash with current clipboard content
	if text, err := m.reader.Read(); err == nil {
		m.lastHash = sha256.Sum256([]byte(text))
	}

	for {
		select {
		case <-ticker.C:
			m.poll()
		case <-stop:
			return
		}
	}
}

func (m *Monitor) poll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.enabled {
		return
	}

	text, err := m.reader.Read()
	if err != nil || text == "" {
		return
	}

	// Size guard: skip text > 1MB
	if len(text) > 1024*1024 {
		return
	}

	hash := sha256.Sum256([]byte(text))
	if hash == m.lastHash {
		return
	}

	// Skip if flagged (undo bypass)
	if m.skipNext {
		m.skipNext = false
		m.lastHash = hash
		return
	}

	if !cleaner.NeedsCleaning(text) {
		m.lastHash = hash
		return
	}

	cleaned := cleaner.Clean(text)
	if cleaned == text {
		m.lastHash = hash
		return
	}

	// Store original for undo
	m.originalText = text
	m.hasOriginal = true

	// Write cleaned text
	if err := m.writer.Write(cleaned); err != nil {
		return
	}

	// Update hash to the cleaned text (loop prevention)
	m.lastHash = sha256.Sum256([]byte(cleaned))

	if m.OnCleaned != nil {
		m.OnCleaned()
	}
}

// SetEnabled toggles monitoring on/off.
func (m *Monitor) SetEnabled(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = enabled
}

// Enabled returns whether monitoring is on.
func (m *Monitor) Enabled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.enabled
}

// Undo restores the original text before the last clean.
// Returns false if there is nothing to undo.
func (m *Monitor) Undo() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.hasOriginal {
		return false
	}

	m.skipNext = true
	if err := m.writer.Write(m.originalText); err != nil {
		return false
	}

	m.hasOriginal = false
	m.originalText = ""

	if m.OnUndo != nil {
		m.OnUndo()
	}
	return true
}

// HasUndo returns whether an undo is available.
func (m *Monitor) HasUndo() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.hasOriginal
}
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./internal/clipboard/
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add internal/clipboard/monitor.go
git commit -m "feat: add clipboard monitor with polling, loop prevention, undo"
```

---

## Task 6: Tray Icons

**Files:**
- Create: `assets/icon.png`
- Create: `assets/icon_active.png`
- Create: `assets/icon_disabled.png`
- Create: `internal/tray/icons.go`

- [ ] **Step 1: Generate simple placeholder icons**

We need 3 PNG icons (22x22 pixels). For now, generate solid-color placeholders using Go. We'll replace these with proper icons later.

```bash
go install github.com/nicholasgasior/gogenimg/cmd/gogenimg@latest || true
```

If the tool isn't available, create simple 22x22 PNGs manually using a small Go script. Create `scripts/gen_icons.go`:

```go
//go:build ignore

package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

func createIcon(path string, c color.RGBA) {
	img := image.NewRGBA(image.Rect(0, 0, 22, 22))
	for x := 0; x < 22; x++ {
		for y := 0; y < 22; y++ {
			img.Set(x, y, c)
		}
	}
	f, _ := os.Create(path)
	defer f.Close()
	png.Encode(f, img)
}

func main() {
	os.MkdirAll("icons", 0o755)
	createIcon("icons/icon.png", color.RGBA{100, 100, 100, 255})        // gray
	createIcon("icons/icon_active.png", color.RGBA{0, 200, 100, 255})   // green
	createIcon("icons/icon_disabled.png", color.RGBA{180, 180, 180, 255}) // light gray
}
```

```bash
go run scripts/gen_icons.go
```

- [ ] **Step 2: Create embedded icons file**

Icons are embedded from `cmd/clearpaste/` (which can reach `icons/` at the module root) and passed to the tray package as parameters. Go's `go:embed` does not allow `..` traversal, so we embed in `main.go` and inject into the tray package.

Create `internal/tray/icons.go`:

```go
package tray

// Icon bytes are injected from main.go at startup.
var (
	IconDefault  []byte
	IconActive   []byte
	IconDisabled []byte
)
```

- [ ] **Step 3: Verify it compiles**

```bash
go build ./internal/tray/
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add icons/ internal/tray/icons.go scripts/gen_icons.go
git commit -m "feat: add placeholder tray icons"
```

---

## Task 7: System Tray

**Files:**
- Create: `internal/tray/tray.go`

- [ ] **Step 1: Install systray dependency**

```bash
go get fyne.io/systray
```

- [ ] **Step 2: Implement tray**

Create `internal/tray/tray.go`:

```go
package tray

import (
	"time"

	"fyne.io/systray"
)

// Callbacks for tray actions.
type Callbacks struct {
	OnToggle func(enabled bool)
	OnUndo   func() bool // returns false if nothing to undo
	HasUndo  func() bool
}

// Run starts the system tray. Blocks until quit.
func Run(cb Callbacks) {
	systray.Run(func() { onReady(cb) }, onExit)
}

func onReady(cb Callbacks) {
	systray.SetIcon(IconDefault)
	systray.SetTitle("ClearPaste")
	systray.SetTooltip("ClearPaste — clipboard cleaner")

	mEnabled := systray.AddMenuItemCheckbox("Enabled", "Toggle clipboard monitoring", true)
	mUndo := systray.AddMenuItem("Undo last clean", "Restore original clipboard text")
	mUndo.Disable()
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit ClearPaste")

	go func() {
		for {
			select {
			case <-mEnabled.ClickedCh:
				if mEnabled.Checked() {
					mEnabled.Uncheck()
					systray.SetIcon(IconDisabled)
					cb.OnToggle(false)
				} else {
					mEnabled.Check()
					systray.SetIcon(IconDefault)
					cb.OnToggle(true)
				}

			case <-mUndo.ClickedCh:
				if cb.OnUndo != nil {
					cb.OnUndo()
				}

			case <-mQuit.ClickedCh:
				systray.Quit()
			}
		}
	}()

	// Goroutine to update undo menu item state (exits when tray quits)
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			if cb.HasUndo != nil && cb.HasUndo() {
				mUndo.Enable()
			} else {
				mUndo.Disable()
			}
		}
	}()
}

func onExit() {
	// Cleanup if needed
}

// FlashIcon briefly shows the active icon then reverts to default.
func FlashIcon() {
	systray.SetIcon(IconActive)
	go func() {
		time.Sleep(2 * time.Second)
		systray.SetIcon(IconDefault)
	}()
}
```

- [ ] **Step 3: Verify it compiles**

```bash
go build ./internal/tray/
```

Expected: no errors (note: systray requires CGO, ensure `CGO_ENABLED=1`).

- [ ] **Step 4: Commit**

```bash
git add internal/tray/tray.go go.mod go.sum
git commit -m "feat: add system tray with enable/disable toggle, undo, quit"
```

---

## Task 8: Wire Everything in main.go

**Files:**
- Modify: `cmd/clearpaste/main.go`

- [ ] **Step 1: Update main.go to wire all components**

Replace `cmd/clearpaste/main.go`:

```go
package main

import (
	_ "embed"

	"github.com/tonytangdev/clearpaste/internal/clipboard"
	"github.com/tonytangdev/clearpaste/internal/tray"
)

//go:embed icons/icon.png
var iconDefault []byte

//go:embed icons/icon_active.png
var iconActive []byte

//go:embed icons/icon_disabled.png
var iconDisabled []byte

func main() {
	// Inject icons into tray package
	tray.IconDefault = iconDefault
	tray.IconActive = iconActive
	tray.IconDisabled = iconDisabled

	cb := &clipboard.System{}
	monitor := clipboard.NewMonitor(cb, cb)

	stop := make(chan struct{})

	monitor.OnCleaned = func() {
		tray.FlashIcon()
	}

	// Start clipboard monitor in background
	go monitor.Start(stop)

	// Run tray (blocks until quit)
	tray.Run(tray.Callbacks{
		OnToggle: func(enabled bool) {
			monitor.SetEnabled(enabled)
		},
		OnUndo: func() bool {
			return monitor.Undo()
		},
		HasUndo: func() bool {
			return monitor.HasUndo()
		},
	})

	close(stop)
}
```

**Note:** The `go:embed` paths are relative to the file location. Since `main.go` is at `cmd/clearpaste/main.go`, the icons must be accessible from there. We need a symlink or to move icons into `cmd/clearpaste/icons/`. The simplest approach: **copy icons into `cmd/clearpaste/icons/`** during the build step.

Update the `Makefile` (from Task 1) to add this:

```makefile
.PHONY: build run test clean

build:
	mkdir -p cmd/clearpaste/icons
	cp icons/*.png cmd/clearpaste/icons/
	go build -o bin/clearpaste ./cmd/clearpaste

run:
	mkdir -p cmd/clearpaste/icons
	cp icons/*.png cmd/clearpaste/icons/
	go run ./cmd/clearpaste

test:
	go test ./... -v

clean:
	rm -rf bin/ cmd/clearpaste/icons/
```

Add `cmd/clearpaste/icons/` to `.gitignore`:

```
bin/
cmd/clearpaste/icons/
```

- [ ] **Step 2: Build and run**

```bash
make build && ./bin/clearpaste
```

Expected: tray icon appears in system tray. Copy text with box-drawing chars → text gets cleaned and icon flashes.

- [ ] **Step 3: Manual test checklist**

1. Copy `│ hello world │` → clipboard should become `hello world`, icon flashes
2. Copy `normal text` → clipboard unchanged
3. Right-click tray → uncheck "Enabled" → copy `│ test │` → clipboard unchanged
4. Right-click tray → "Undo last clean" → clipboard should restore original
5. Right-click tray → "Quit" → app exits

- [ ] **Step 4: Commit**

```bash
git add cmd/clearpaste/main.go
git commit -m "feat: wire tray, monitor, and cleaner in main"
```

---

## Task 9: GoReleaser + GitHub Actions

**Files:**
- Create: `.goreleaser.yml`
- Create: `.github/workflows/release.yml`

- [ ] **Step 1: Create GoReleaser config**

Create `.goreleaser.yml`.

Since CGO is required (for systray) and CGO cross-compilation is non-trivial, we use **separate GoReleaser config per OS** and a **matrix build** in GitHub Actions — each OS builds natively.

Create `.goreleaser.darwin.yml`:

```yaml
version: 2

builds:
  - id: clearpaste-darwin
    main: ./cmd/clearpaste
    binary: clearpaste
    env:
      - CGO_ENABLED=1
    goos:
      - darwin
    goarch:
      - amd64
      - arm64

archives:
  - id: clearpaste-darwin
    format: tar.gz
    name_template: "clearpaste_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

checksum:
  name_template: "checksums_darwin.txt"

changelog:
  sort: asc

brews:
  - repository:
      owner: tonytangdev
      name: homebrew-tap
    name: clearpaste
    homepage: "https://github.com/tonytangdev/clearpaste"
    description: "Auto-clean terminal formatting artifacts from clipboard"
```

Create `.goreleaser.linux.yml`:

```yaml
version: 2

builds:
  - id: clearpaste-linux
    main: ./cmd/clearpaste
    binary: clearpaste
    env:
      - CGO_ENABLED=1
    goos:
      - linux
    goarch:
      - amd64

archives:
  - id: clearpaste-linux
    format: tar.gz
    name_template: "clearpaste_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

checksum:
  name_template: "checksums_linux.txt"

changelog:
  disable: true
```

Create `.goreleaser.windows.yml`:

```yaml
version: 2

builds:
  - id: clearpaste-windows
    main: ./cmd/clearpaste
    binary: clearpaste
    env:
      - CGO_ENABLED=1
    goos:
      - windows
    goarch:
      - amd64

archives:
  - id: clearpaste-windows
    format: zip
    name_template: "clearpaste_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

checksum:
  name_template: "checksums_windows.txt"

changelog:
  disable: true
```

- [ ] **Step 2: Create GitHub Actions release workflow**

Create `.github/workflows/release.yml` with a matrix strategy — each OS builds natively with CGO:

```yaml
name: Release

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write

jobs:
  release:
    strategy:
      matrix:
        include:
          - os: macos-latest
            goreleaser_config: .goreleaser.darwin.yml
          - os: ubuntu-latest
            goreleaser_config: .goreleaser.linux.yml
            deps: sudo apt-get update && sudo apt-get install -y libx11-dev
          - os: windows-latest
            goreleaser_config: .goreleaser.windows.yml
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"

      - name: Install platform deps
        if: matrix.deps
        run: ${{ matrix.deps }}

      - uses: goreleaser/goreleaser-action@v6
        with:
          version: latest
          args: release --clean --config ${{ matrix.goreleaser_config }}
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

- [ ] **Step 3: Commit**

```bash
git add .goreleaser.darwin.yml .goreleaser.linux.yml .goreleaser.windows.yml .github/workflows/release.yml
git commit -m "ci: add goreleaser configs and matrix release workflow"
```

---

## Task 10: README & License

**Files:**
- Create: `README.md`
- Create: `LICENSE`

- [ ] **Step 1: Create README**

Create `README.md`:

```markdown
# ClearPaste

Auto-clean terminal formatting artifacts from your clipboard. Runs in the system tray, monitors clipboard changes, and strips box-drawing characters and fixes broken line wraps from Claude Code, Codex, and other TUI tools.

## Install

### macOS (Homebrew)

```bash
brew install tonytangdev/tap/clearpaste
```

### Download binary

Grab the latest release from [GitHub Releases](https://github.com/tonytangdev/clearpaste/releases).

## Usage

Run `clearpaste` — it appears in your system tray.

- **Right-click** the tray icon for options
- **Enabled/Disabled** — toggle clipboard monitoring
- **Undo last clean** — restore the original text
- Icon flashes green when text is cleaned

## What it cleans

- Strips Unicode box-drawing characters (U+2500–U+257F) and block elements (U+2580–U+259F)
- Rejoins broken line wraps
- Collapses excessive whitespace
- Preserves list structure, code blocks, and paragraph breaks

## Build from source

```bash
git clone https://github.com/tonytangdev/clearpaste.git
cd clearpaste
make build
./bin/clearpaste
```

## License

MIT
```

- [ ] **Step 2: Create LICENSE file**

Create `LICENSE` with MIT license text (replace year and name):

```
MIT License

Copyright (c) 2026 Tony Tang

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

- [ ] **Step 3: Commit**

```bash
git add README.md LICENSE
git commit -m "docs: add README and MIT license"
```

---

## Task Order & Dependencies

```
Task 1 (scaffold) → Task 2 (detector) → Task 3 (cleaner) → Task 4 (clipboard interface)
    → Task 5 (monitor) → Task 6 (icons) → Task 7 (tray) → Task 8 (wire main)
    → Task 9 (CI/release) → Task 10 (README/license)
```

All tasks are sequential — each builds on the previous.
