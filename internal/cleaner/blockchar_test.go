package cleaner

import "testing"

func TestClean_BlockElementQuarterBar(t *testing.T) {
	input := "  \u258e every team i've been on has the same problem\n\n  \u258e 6 months in, nobody remembers why we picked Postgres over Mongo. or why we went with REST instead of GraphQL.\n\n  \u258e the decision made total sense at the time. but the context is gone. the person who made it left. the Slack thread is buried.\n\n  \u258e so the next dev either re-debates it from scratch or just accepts it blindly.\n\n  \u258e i started writing down every architecture decision with:\n  \u258e - why we chose it\n  \u258e - what we considered instead\n  \u258e - what would make us revisit it\n\n  \u258e took maybe 2 min per decision. saved hours of \"wait why did we do this\" conversations.\n\n  \u258e the unsexy stuff compounds."

	if !NeedsCleaning(input) {
		t.Fatal("NeedsCleaning returned false for \u258e input")
	}

	want := "every team i've been on has the same problem\n\n6 months in, nobody remembers why we picked Postgres over Mongo. or why we went with REST instead of GraphQL.\n\nthe decision made total sense at the time. but the context is gone. the person who made it left. the Slack thread is buried.\n\nso the next dev either re-debates it from scratch or just accepts it blindly.\n\ni started writing down every architecture decision with:\n- why we chose it\n- what we considered instead\n- what would make us revisit it\n\ntook maybe 2 min per decision. saved hours of \"wait why did we do this\" conversations.\n\nthe unsexy stuff compounds."

	got := Clean(input)
	if got != want {
		t.Errorf("Clean() =\n%q\nwant:\n%q", got, want)
	}
}

func TestClean_IndentedProseWithoutBlockChars(t *testing.T) {
	input := "  every team i've been on has the same problem\n\n  6 months in, nobody remembers why we picked Postgres over Mongo. or why we went with REST instead of GraphQL.\n\n  the decision made total sense at the time. but the context is gone. the person who made it left. the Slack thread is buried.\n\n  so the next dev either re-debates it from scratch or just accepts it blindly.\n\n  i started writing down every architecture decision with:\n  - why we chose it\n  - what we considered instead\n  - what would make us revisit it\n\n  took maybe 2 min per decision. saved hours of \"wait why did we do this\" conversations.\n\n  the unsexy stuff compounds."

	if !NeedsCleaning(input) {
		t.Fatal("NeedsCleaning returned false for indented prose without block chars")
	}

	want := "every team i've been on has the same problem\n\n6 months in, nobody remembers why we picked Postgres over Mongo. or why we went with REST instead of GraphQL.\n\nthe decision made total sense at the time. but the context is gone. the person who made it left. the Slack thread is buried.\n\nso the next dev either re-debates it from scratch or just accepts it blindly.\n\ni started writing down every architecture decision with:\n- why we chose it\n- what we considered instead\n- what would make us revisit it\n\ntook maybe 2 min per decision. saved hours of \"wait why did we do this\" conversations.\n\nthe unsexy stuff compounds."

	got := Clean(input)
	if got != want {
		t.Errorf("Clean() =\n%q\nwant:\n%q", got, want)
	}
}

func TestHasUniformIndent_NoIndent(t *testing.T) {
	if hasUniformIndent("First sentence.\nSecond sentence.\nThird sentence.") {
		t.Error("should not trigger for unindented text")
	}
}

func TestHasUniformIndent_TooFewLines(t *testing.T) {
	if hasUniformIndent("  hello world foo bar\n  another line here too") {
		t.Error("should not trigger for only 2 non-empty lines")
	}
}

func TestHasUniformIndent_NoProse(t *testing.T) {
	if hasUniformIndent("  a = 1\n  b = 2\n  c = 3") {
		t.Error("should not trigger for short code-like lines")
	}
}
