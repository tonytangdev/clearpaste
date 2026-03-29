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
