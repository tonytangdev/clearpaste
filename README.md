# ClearPaste

Auto-clean terminal formatting artifacts from your clipboard. Runs in the system tray, monitors clipboard changes, and strips box-drawing characters and fixes broken line wraps from Claude Code, Codex, and other TUI tools.

## Install

### macOS (Homebrew)

```bash
brew install tonytangdev/tap/clearpaste
brew services start clearpaste
```

`brew services start` registers ClearPaste as a Launch Agent — it starts automatically on login and restarts if it crashes. To stop it: `brew services stop clearpaste`.

### Linux / Windows

Download the binary for your platform from [GitHub Releases](https://github.com/tonytangdev/clearpaste/releases) and run it manually.

### Build from source

```bash
git clone https://github.com/tonytangdev/clearpaste.git
cd clearpaste
make build
./bin/clearpaste
```

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

## Inspiration

Inspired by [Clean-Clode](https://github.com/TheJoWo/Clean-Clode). I wanted something with less friction — no manual paste, just copy and it's clean.

## License

MIT
