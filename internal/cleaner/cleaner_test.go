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
	input := "│ hello | world │"
	want := "hello world"
	got := Clean(input)
	if got != want {
		t.Errorf("Clean(%q) = %q, want %q", input, got, want)
	}
}

func TestClean_PreservePipeOnLinesWithoutBoxChars(t *testing.T) {
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
	input := "  │ Interesting framing. If agents are making stack decisions, the question\n  │ becomes: what context are they pulling from? Right now it's mostly training\n  │ data + whatever's in CLAUDE.md. Teams that structure their decision history\n  │ so agents can query it will have a real edge here."
	want := "Interesting framing. If agents are making stack decisions, the question becomes: what context are they pulling from? Right now it's mostly training data + whatever's in CLAUDE.md. Teams that structure their decision history so agents can query it will have a real edge here."
	got := Clean(input)
	if got != want {
		t.Errorf("Clean() =\n%q\nwant:\n%q", got, want)
	}
}
