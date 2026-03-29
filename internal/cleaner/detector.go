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
