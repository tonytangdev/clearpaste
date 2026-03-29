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
