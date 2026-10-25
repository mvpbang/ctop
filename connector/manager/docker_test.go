package manager

import (
	"bytes"
	"testing"
)

func TestFrameWriter(t *testing.T) {
	var stdout, stderr, stdin bytes.Buffer
	fw := &frameWriter{
		stdout: &stdout,
		stderr: &stderr,
		stdin:  &stdin,
	}

	// 1. Docker multiplexed STDOUT frame [1, 0, 0, 0, 0, 0, 0, 4, 't', 'e', 's', 't']
	frame := []byte{STDOUT, 0, 0, 0, 0, 0, 0, 4, 't', 'e', 's', 't'}
	n, err := fw.Write(frame)
	if err != nil {
		t.Fatalf("unexpected error on multiplexed frame: %v", err)
	}
	if n != len(frame) {
		t.Errorf("expected n=%d, got %d", len(frame), n)
	}
	if stdout.String() != "test" {
		t.Errorf("expected stdout 'test', got %q", stdout.String())
	}

	stdout.Reset()

	// 2. Raw terminal bytes shorter than 8 bytes (e.g. prompt or short string)
	shortBytes := []byte("hi")
	n, err = fw.Write(shortBytes)
	if err != nil {
		t.Fatalf("unexpected error on short raw bytes: %v", err)
	}
	if n != len(shortBytes) {
		t.Errorf("expected n=%d, got %d", len(shortBytes), n)
	}
	if stdout.String() != "hi" {
		t.Errorf("expected stdout 'hi', got %q", stdout.String())
	}

	stdout.Reset()

	// 3. Raw terminal bytes starting with escape sequence (\033, 0x1b)
	escBytes := []byte("\x1b[0m\x1b[?25hhello world")
	n, err = fw.Write(escBytes)
	if err != nil {
		t.Fatalf("unexpected error on raw escape sequence: %v", err)
	}
	if stdout.String() != string(escBytes) {
		t.Errorf("expected stdout %q, got %q", string(escBytes), stdout.String())
	}
}

func TestManagerKill(t *testing.T) {
	var mgr Manager = NewMock()
	err := mgr.Kill()
	if err == nil {
		t.Fatalf("expected ActionNotImplErr for Mock.Kill(), got nil")
	}
}
