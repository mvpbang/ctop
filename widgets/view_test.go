package widgets

import "testing"

func TestSplitEmptyLine(t *testing.T) {

	result := splitLine("", 5)
	if len(result) != 0 {
		t.Errorf("expected: 0 lines, got: %d", len(result))
	}
}

func TestSplitLineShorterThanLimit(t *testing.T) {

	result := splitLine("hello", 7)
	if len(result) != 1 {
		t.Errorf("expected: 0 lines, got: %d", len(result))
	}
}

func TestSplitLineLongerThanLimit(t *testing.T) {

	result := splitLine("hello", 3)
	if len(result) != 2 {
		t.Errorf("expected: 0 lines, got: %d", len(result))
	}
}

func TestSplitLineSameAsLimit(t *testing.T) {

	result := splitLine("hello", 5)
	if len(result) != 1 {
		t.Errorf("expected: 0 lines, got: %d", len(result))
	}
}

func TestSplitLineZeroOrNegativeLimit(t *testing.T) {
	resultZero := splitLine("hello", 0)
	if len(resultZero) != 1 || resultZero[0] != "hello" {
		t.Errorf("expected ['hello'], got %v", resultZero)
	}

	resultNeg := splitLine("hello", -1)
	if len(resultNeg) != 1 || resultNeg[0] != "hello" {
		t.Errorf("expected ['hello'], got %v", resultNeg)
	}
}

type dummyToggleText string

func (d dummyToggleText) Toggle(on bool) string { return string(d) }

func TestTextViewScroll(t *testing.T) {
	tv := &TextView{
		padding: Padding{4, 2},
		render:  make(chan bool, 10),
	}
	tv.Height = 10
	tv.Width = 40
	tv.Up()
	if tv.offset != 1 {
		t.Errorf("expected offset 1, got %d", tv.offset)
	}
	tv.Down()
	if tv.offset != 0 {
		t.Errorf("expected offset 0, got %d", tv.offset)
	}
	tv.Down()
	if tv.offset != 0 {
		t.Errorf("expected offset 0 on lower bound, got %d", tv.offset)
	}
	tv.PageUp()
	if tv.offset <= 0 {
		t.Errorf("expected PageUp offset > 0, got %d", tv.offset)
	}
	tv.PageDown()
	if tv.offset != 0 {
		t.Errorf("expected PageDown offset 0, got %d", tv.offset)
	}
}
