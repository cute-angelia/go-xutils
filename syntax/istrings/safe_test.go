package istrings

import "testing"

func TestSanitizeNameFiltersUnusualUnicode(t *testing.T) {
	got := SanitizeName("报告⁸ɞ⭐.txt", "_")
	want := "报告___.txt"

	if got != want {
		t.Fatalf("SanitizeName() = %q, want %q", got, want)
	}
}

func TestSanitizeNameKeepsCommonFilenameCharacters(t *testing.T) {
	got := SanitizeName("Report 2026_04-12(草稿)[v2].txt", "_")
	want := "Report 2026_04-12(草稿)[v2].txt"

	if got != want {
		t.Fatalf("SanitizeName() = %q, want %q", got, want)
	}
}
