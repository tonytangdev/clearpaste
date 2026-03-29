# ClearPaste — Design Spec

Cross-platform system tray app that monitors the clipboard and auto-cleans terminal formatting artifacts from copied text.

## Problem

Copying text from Claude Code, Codex, or other TUI tools produces mangled output: box-drawing characters (`│`, `┃`, `╏`), broken line wraps, excessive whitespace. Users must manually clean this before pasting elsewhere.

## Solution

A lightweight Go binary that runs in the system tray, watches the clipboard, and automatically cleans terminal-formatted text — silently replacing it with clean prose.

## Requirements

- **Detection**: only clean text containing box-drawing chars from the Unicode Box Drawing (U+2500–U+257F) and Block Elements (U+2580–U+259F) ranges — leave normal text untouched
- **Feedback**: flash tray icon for 2s after cleaning
- **Undo**: one-level undo via tray menu (restores original text)
- **Controls**: tray right-click menu — toggle on/off, undo last clean, quit
- **Platforms**: macOS (amd64 + arm64), Windows (amd64), Linux (amd64)
- **Distribution**: GitHub Releases + Homebrew tap (v1), Scoop/Snap/AUR later

## Architecture

```
clearpaste/
├── cmd/
│   └── clearpaste/
│       └── main.go              — entry point, wires everything together
├── internal/
│   ├── clipboard/
│   │   ├── monitor.go           — polls clipboard, detects changes
│   │   └── clipboard.go         — read/write clipboard (interface + impl)
│   ├── cleaner/
│   │   ├── detector.go          — decides if text needs cleaning
│   │   ├── cleaner.go           — text cleaning transforms
│   │   └── cleaner_test.go      — tests
│   └── tray/
│       ├── tray.go              — system tray menu, icon, user actions
│       └── icons.go             — embedded icon assets
├── assets/
│   ├── icon.png                 — normal tray icon
│   └── icon_active.png          — "just cleaned" flash icon
├── go.mod
├── go.sum
├── Makefile
└── .goreleaser.yml              — cross-platform build + release config
```

### Principles

- **Single responsibility**: each package does one thing
- **Dependency inversion**: monitor depends on `ClipboardReader` interface, not concrete impl
- **Domain logic isolated**: cleaner is pure functions, zero dependencies
- **Composition in main**: `main.go` wires pieces together, no business logic

### Interfaces

```go
// clipboard package
type Reader interface {
    Read() (string, error)
}
type Writer interface {
    Write(text string) error
}

// cleaner package — pure functions
func NeedsCleaning(text string) bool
func Clean(text string) string
```

## Clipboard Monitor

- Polls clipboard every 300ms using a Go ticker
- Detects changes via clipboard content hash comparison
- **Text-only**: only reads text content from clipboard. Non-text content (images, files, rich text) is ignored — `Read()` returns empty/error for non-text, which we skip silently.
- **Size guard**: skip text larger than 1MB to avoid processing huge log files
- On change: check `NeedsCleaning()` → if true, `Clean()` → write back
- **Loop prevention**: after writing cleaned text, store its hash. Next poll sees the change but hash matches → skip.
- **Undo bypass**: after undo writes original text back, set a `skipNext` flag so the monitor ignores the next clipboard change (preventing re-cleaning of the restored dirty text).
- Stores original text for one-level undo
- **Error handling**: all read/write errors are silently ignored (background daemon, no stderr to watch). No log file in v1.

### Why poll instead of `clipboard.Watch()`

`golang.design/x/clipboard` offers a `Watch()` channel API, but polling gives us more control over loop prevention and undo bypass. The 300ms poll interval is negligible CPU cost. We can switch to `Watch()` later if needed.

## Cleaning Logic

### Detection (`NeedsCleaning`)

Returns true if text contains any character in the Unicode Box Drawing range (U+2500–U+257F) or Block Elements range (U+2580–U+259F). This covers `│ ┃ ╏ ╎ ▌ ┌ ┐ └ ┘ ├ ┤ ┬ ┴ ┼` and all their variants.

Does NOT trigger on ASCII pipe `|` alone — too many false positives (markdown tables, shell commands, code).

### Cleaning Pipeline (`Clean`)

Applied in order:

1. **Strip box-drawing and block element chars** — remove all U+2500–U+257F and U+2580–U+259F characters via regex. ASCII `|` is only stripped on lines where a box-drawing char is also present (to catch `│` mixed with `|` in the same TUI frame, while preserving `|` in normal text).
2. **Collapse multiple spaces** — 2+ spaces → single space
3. **Rejoin broken lines** — merge line into previous one. A line is kept separate (NOT merged) if:
   - It is empty (paragraph break)
   - It starts with a list marker (`-`, `*`, `•`)
   - It starts with a numbered list (`1.`, `2.`, etc.)
   - It starts with an emoji (Unicode Emoji_Presentation)
   - The previous line ends with sentence-ending punctuation (`.`, `!`, `?`) AND this line starts with an uppercase letter — this is a new sentence/paragraph. Abbreviations (`e.g.`, `Dr.`, `v1.0`) are not sentence-ending because they are followed by a lowercase letter or digit on the next line.
   - It looks like code: starts with 4+ spaces, starts with a tab, or contains 3+ chars from `{}();=` on the same line
4. **Normalize whitespace** — trim each line
5. **Collapse excessive newlines** — 3+ newlines → 2

### Preserved

- Paragraph breaks (double newlines)
- List structure (bullets, numbered lists)
- Code blocks (4+ leading spaces or tab-indented lines)
- Code-like lines (3+ occurrences of `{}();=` characters on the same line)

## System Tray

### Menu (right-click)

```
ClearPaste
─────────────
✓ Enabled
  Undo last clean
─────────────
  Quit
```

### Icon States

- **Default**: clipboard outline icon
- **Cleaned flash**: accent-colored icon for 2 seconds
- **Disabled**: dimmed/grayed icon when toggled off

### Startup

Starts enabled by default. No built-in "start at login" — users configure via OS settings (Login Items / Task Scheduler / `.desktop` autostart).

### State Persistence

Toggle state (enabled/disabled) is NOT persisted across restarts in v1. App always starts enabled. This is a known limitation; config file can be added later if needed.

## Dependencies

| Dependency | Purpose | Status |
|---|---|---|
| `fyne.io/systray` | Cross-platform system tray | Actively maintained |
| `golang.design/x/clipboard` | Cross-platform clipboard read/write | Low maintenance (~12 months since last release) but 450+ importers. Fallback: `atotto/clipboard` for read/write with manual polling. |

## Build & Distribution

### GoReleaser + GitHub Actions

- Tag push (`v0.1.0`) triggers GitHub Actions
- GoReleaser builds for macOS (amd64 + arm64), Windows (amd64), Linux (amd64)
- Uploads binaries to GitHub Releases

### Package Managers

| Platform | Manager | Timeline |
|---|---|---|
| macOS | Homebrew tap (`tonytangdev/homebrew-tap`) | v1 |
| Windows | Scoop bucket | post-v1 |
| Linux | Snap or AUR | post-v1 |

## Tech Stack

- **Language**: Go
- **Tray**: `fyne.io/systray`
- **Clipboard**: `golang.design/x/clipboard` (fallback: `atotto/clipboard`)
- **Build**: GoReleaser
- **CI**: GitHub Actions
- **License**: MIT
