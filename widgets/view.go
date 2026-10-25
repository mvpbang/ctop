package widgets

import (
	"sync"

	ui "github.com/gizak/termui"
	"github.com/mattn/go-runewidth"
)

type ToggleText interface {
	// returns text for toggle on/off
	Toggle(on bool) string
}

type TextView struct {
	ui.Block
	inputStream <-chan ToggleText
	render      chan bool
	Text        []ToggleText // all the text
	TextOut     []string     // text to be displayed
	TextFgColor ui.Attribute
	TextBgColor ui.Attribute
	padding     Padding
	// mu guards fields touched from both the render loop and key handlers
	mu          sync.RWMutex
	toggleState bool
	offset      int
}

func (t *TextView) triggerRender() {
	select {
	case t.render <- true:
	default:
	}
}

func (t *TextView) Up() {
	t.mu.Lock()
	t.offset++
	t.mu.Unlock()
	t.triggerRender()
}

func (t *TextView) Down() {
	t.mu.Lock()
	if t.offset > 0 {
		t.offset--
	}
	t.mu.Unlock()
	t.triggerRender()
}

func (t *TextView) PageUp() {
	t.mu.Lock()
	h := t.Height - (t.padding[1] * 2)
	if h <= 0 {
		h = 5
	}
	t.offset += h
	t.mu.Unlock()
	t.triggerRender()
}

func (t *TextView) PageDown() {
	t.mu.Lock()
	h := t.Height - (t.padding[1] * 2)
	if h <= 0 {
		h = 5
	}
	t.offset -= h
	if t.offset < 0 {
		t.offset = 0
	}
	t.mu.Unlock()
	t.triggerRender()
}

func NewTextView(lines <-chan ToggleText) *TextView {
	t := &TextView{
		Block:       *ui.NewBlock(),
		inputStream: lines,
		render:      make(chan bool),
		Text:        []ToggleText{},
		TextOut:     []string{},
		TextFgColor: ui.ThemeAttr("menu.text.fg"),
		TextBgColor: ui.ThemeAttr("menu.text.bg"),
		padding:     Padding{4, 2},
		offset:      0,
	}

	t.BorderFg = ui.ThemeAttr("menu.border.fg")
	t.BorderLabelFg = ui.ThemeAttr("menu.label.fg")
	t.Height = ui.TermHeight()
	t.Width = ui.TermWidth()

	t.readInputLoop()
	t.renderLoop()
	return t
}

// Adjusts text inside this view according to the window size. No need to call ui.Render(...)
// after calling this method, it is called automatically
func (t *TextView) Resize() {
	ui.Clear()
	t.mu.Lock()
	t.Height = ui.TermHeight()
	t.Width = ui.TermWidth()
	t.mu.Unlock()
	t.triggerRender()
}

// Toggles text inside this view. No need to call ui.Render(...) after calling this method,
// it is called automatically
func (t *TextView) Toggle() {
	t.mu.Lock()
	t.toggleState = !t.toggleState
	t.mu.Unlock()
	t.triggerRender()
}

func (t *TextView) Buffer() ui.Buffer {
	var cell ui.Cell
	buf := t.Block.Buffer()

	t.mu.RLock()
	lines := t.TextOut
	t.mu.RUnlock()

	x := t.Block.X + t.padding[0]
	y := t.Block.Y + t.padding[1]

	for _, line := range lines {
		for _, ch := range line {
			cell = ui.Cell{Ch: ch, Fg: t.TextFgColor, Bg: t.TextBgColor}
			buf.Set(x, y, cell)
			x = x + runewidth.RuneWidth(ch)
		}
		x = t.Block.X + t.padding[0]
		y++
	}
	return buf
}

func (t *TextView) renderLoop() {
	go func() {
		for range t.render {
			t.mu.RLock()
			maxWidth := t.Width - (t.padding[0] * 2)
			height := t.Height - (t.padding[1] * 2)
			toggle := t.toggleState
			offset := t.offset
			text := t.Text
			t.mu.RUnlock()

			if maxWidth <= 0 || height <= 0 {
				continue
			}

			var allLines []string
			for i := 0; i < len(text); i++ {
				lines := splitLine(text[i].Toggle(toggle), maxWidth)
				allLines = append(allLines, lines...)
			}

			total := len(allLines)
			if total <= height {
				offset = 0
			} else {
				maxOffset := total - height
				if offset > maxOffset {
					offset = maxOffset
				}
				if offset < 0 {
					offset = 0
				}
				end := total - offset
				allLines = allLines[end-height : end]
			}

			t.mu.Lock()
			t.offset = offset
			t.TextOut = allLines
			t.mu.Unlock()
			ui.Render(t)
		}
	}()
}

func (t *TextView) readInputLoop() {
	go func() {
		for line := range t.inputStream {
			t.mu.Lock()
			t.Text = append(t.Text, line)
			t.mu.Unlock()
			t.render <- true
		}
		close(t.render)
	}()
}

func splitLine(line string, lineSize int) []string {
	if line == "" {
		return []string{}
	}
	if lineSize <= 0 {
		return []string{line}
	}

	var lines []string
	for {
		if len(line) <= lineSize {
			lines = append(lines, line)
			return lines
		}
		lines = append(lines, line[:lineSize])
		line = line[lineSize:]
	}
}
