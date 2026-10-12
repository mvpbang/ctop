package widgets

import (
	"testing"

	ui "github.com/gizak/termui"
)

func TestInputKeyPress(t *testing.T) {
	input := NewInput()
	stream := input.Stream()
	go func() {
		for range stream {
		}
	}()

	// 输入常规字母
	input.KeyPress(ui.Event{Path: "/sys/kbd/a"})
	input.KeyPress(ui.Event{Path: "/sys/kbd/b"})
	if input.Data != "ab" {
		t.Fatalf("expected 'ab', got '%s'", input.Data)
	}

	// 通过 <backspace> 退格
	input.KeyPress(ui.Event{Path: "/sys/kbd/<backspace>"})
	if input.Data != "a" {
		t.Fatalf("expected 'a' after <backspace>, got '%s'", input.Data)
	}

	// 通过 <delete> 退格
	input.KeyPress(ui.Event{Path: "/sys/kbd/<delete>"})
	if input.Data != "" {
		t.Fatalf("expected '' after <delete>, got '%s'", input.Data)
	}

	// 特殊符号及空格输入
	input.KeyPress(ui.Event{Path: "/sys/kbd/c"})
	input.KeyPress(ui.Event{Path: "/sys/kbd/:"})
	input.KeyPress(ui.Event{Path: "/sys/kbd//"})
	input.KeyPress(ui.Event{Path: "/sys/kbd/<space>"})
	input.KeyPress(ui.Event{Path: "/sys/kbd/d"})
	if input.Data != "c:/ d" {
		t.Fatalf("expected 'c:/ d', got '%s'", input.Data)
	}
}
