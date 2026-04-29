package clifmt

import (
	"strings"
	"testing"
)

func TestLineDiff_Basic(t *testing.T) {
	old := "line1\nline2\nline3\n"
	new := "line1\nline2modified\nline3\n"

	lines := lineDiff(old, new)
	if len(lines) == 0 {
		t.Fatal("expected non-empty diff")
	}

	var hasDelete, hasInsert bool
	for _, dl := range lines {
		if dl.kind == '-' && strings.Contains(dl.text, "line2") {
			hasDelete = true
		}
		if dl.kind == '+' && strings.Contains(dl.text, "line2modified") {
			hasInsert = true
		}
	}
	if !hasDelete {
		t.Error("expected a deleted line containing 'line2'")
	}
	if !hasInsert {
		t.Error("expected an inserted line containing 'line2modified'")
	}
}

func TestLineDiff_LineNumbers(t *testing.T) {
	old := "a\nb\nc\n"
	new := "a\nB\nc\n"

	lines := lineDiff(old, new)

	for _, dl := range lines {
		if dl.kind == ' ' {
			if dl.oldNum == 0 || dl.newNum == 0 {
				t.Errorf("context line should have both old and new line numbers, got old=%d new=%d", dl.oldNum, dl.newNum)
			}
		}
		if dl.kind == '-' {
			if dl.oldNum == 0 {
				t.Error("deleted line should have old line number")
			}
			if dl.newNum != 0 {
				t.Error("deleted line should not have new line number")
			}
		}
		if dl.kind == '+' {
			if dl.newNum == 0 {
				t.Error("inserted line should have new line number")
			}
			if dl.oldNum != 0 {
				t.Error("inserted line should not have old line number")
			}
		}
	}
}

func TestFoldContext(t *testing.T) {
	lines := []diffLine{
		{kind: ' ', text: "a", oldNum: 1, newNum: 1},
		{kind: ' ', text: "b", oldNum: 2, newNum: 2},
		{kind: ' ', text: "c", oldNum: 3, newNum: 3},
		{kind: ' ', text: "d", oldNum: 4, newNum: 4},
		{kind: ' ', text: "e", oldNum: 5, newNum: 5},
		{kind: ' ', text: "f", oldNum: 6, newNum: 6},
		{kind: ' ', text: "g", oldNum: 7, newNum: 7},
		{kind: ' ', text: "h", oldNum: 8, newNum: 8},
		{kind: ' ', text: "i", oldNum: 9, newNum: 9},
		{kind: '-', text: "j", oldNum: 10, newNum: 0},
		{kind: '+', text: "J", oldNum: 0, newNum: 10},
		{kind: ' ', text: "k", oldNum: 11, newNum: 11},
		{kind: ' ', text: "l", oldNum: 12, newNum: 12},
	}

	folded := foldContext(lines, 2)

	if len(folded) == 0 {
		t.Fatal("expected non-empty folded result")
	}

	var hasFoldMarker bool
	var hasJ bool
	for _, dl := range folded {
		if dl.kind == 0 {
			hasFoldMarker = true
		}
		if dl.kind == '-' && dl.text == "j" {
			hasJ = true
		}
	}
	if !hasFoldMarker {
		t.Error("expected a fold marker for distant context lines")
	}
	if !hasJ {
		t.Error("expected the change (line 'j') to be present")
	}
}

func TestFoldContext_NoChange(t *testing.T) {
	lines := []diffLine{
		{kind: ' ', text: "a", oldNum: 1, newNum: 1},
		{kind: ' ', text: "b", oldNum: 2, newNum: 2},
	}
	folded := foldContext(lines, 2)
	if folded != nil {
		t.Error("expected nil when there are no changes to fold around")
	}
}

func TestRenderDiff_WorksWithoutColor(t *testing.T) {
	// In test environments stdout is not a terminal, so useColor() is false.
	// RenderDiff should still produce a plain-text diff (no ANSI codes).
	result := RenderDiff("hello.go", "a\nb\nc\n", "a\nB\nc\n")
	if result == "" {
		t.Fatal("expected non-empty diff even when color is disabled")
	}
	if strings.Contains(result, "\x1b[") {
		t.Error("expected plain-text diff without ANSI codes when color is disabled")
	}
	if !strings.Contains(result, "-") {
		t.Error("expected diff to contain deletion marker")
	}
	if !strings.Contains(result, "+") {
		t.Error("expected diff to contain insertion marker")
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"a\nb\nc\n", []string{"a", "b", "c"}},
		{"a\nb\nc", []string{"a", "b", "c"}},
		{"", nil},
		{"single", []string{"single"}},
	}
	for _, tt := range tests {
		got := splitLines(tt.input)
		if len(got) != len(tt.expected) {
			t.Errorf("splitLines(%q) = %v, want %v", tt.input, got, tt.expected)
			continue
		}
		for i := range got {
			if got[i] != tt.expected[i] {
				t.Errorf("splitLines(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.expected[i])
			}
		}
	}
}
