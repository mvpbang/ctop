//go:build !release
// +build !release

package collector

import (
	"testing"
	"time"
)

func TestMockLogsStop(t *testing.T) {
	ml := &MockLogs{make(chan bool)}
	stream := ml.Stream()

	select {
	case <-stream:
	case <-time.After(1 * time.Second):
		t.Fatalf("timed out waiting for first log")
	}

	ml.Stop()

	// 在 Stop 之后，stream 必须被关闭且不能继续无限生产
	done := make(chan bool)
	go func() {
		for range stream {
		}
		done <- true
	}()

	select {
	case <-done:
		// 成功退出
	case <-time.After(800 * time.Millisecond):
		t.Fatalf("MockLogs stream did not stop after Stop() called")
	}
}
