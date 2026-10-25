package main

import (
	"testing"

	"github.com/bcicen/ctop/cwidgets/compact"
	ui "github.com/gizak/termui"
)

func TestSlashKeyBinding(t *testing.T) {
	ui.DefaultEvtStream.ResetHandlers()

	called := false
	HandleKey("/sys/kbd//", func(e ui.Event) {
		called = true
	})

	if _, ok := ui.DefaultEvtStream.Handlers["/sys/kbd//"]; !ok {
		t.Fatalf("expected /sys/kbd// handler to exist")
	}

	if _, ok := ui.DefaultEvtStream.Handlers["/sys/kbd"]; ok {
		t.Fatalf("/sys/kbd must not be registered directly")
	}

	fn := ui.DefaultEvtStream.Handlers["/sys/kbd//"]
	fn(ui.Event{Path: "/sys/kbd//"})
	if !called {
		t.Fatalf("expected handler to be invoked")
	}
}

func TestPgCountZeroMaxRows(t *testing.T) {
	cGrid = compact.NewCompactGrid()
	gc := &GridCursor{}
	pages := gc.pgCount()
	if pages != 0 {
		t.Errorf("expected 0 pages, got %d", pages)
	}
}
