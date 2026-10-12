package widgets

import (
	"strings"

	ui "github.com/gizak/termui"
	tm "github.com/nsf/termbox-go"
)

var (
	input_chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-_.:/ "
)

type Padding [2]int // x,y padding

type Input struct {
	ui.Block
	Label       string
	Data        string
	MaxLen      int
	TextFgColor ui.Attribute
	TextBgColor ui.Attribute
	stream      chan string // stream text as it changes
	padding     Padding
}

func NewInput() *Input {
	i := &Input{
		Block:       *ui.NewBlock(),
		Label:       "input",
		MaxLen:      50,
		TextFgColor: ui.ThemeAttr("menu.text.fg"),
		TextBgColor: ui.ThemeAttr("menu.text.bg"),
		padding:     Padding{4, 2},
	}
	i.BorderFg = ui.ThemeAttr("menu.border.fg")
	i.BorderLabelFg = ui.ThemeAttr("menu.label.fg")
	i.calcSize()
	return i
}

func (i *Input) calcSize() {
	i.Height = 3 // minimum height
	i.Width = i.MaxLen + (i.padding[0] * 2)
}

func (i *Input) Buffer() ui.Buffer {
	var cell ui.Cell
	buf := i.Block.Buffer()

	x := i.Block.X + i.padding[0]
	y := i.Block.Y + 1
	for _, ch := range i.Data {
		cell = ui.Cell{Ch: ch, Fg: i.TextFgColor, Bg: i.TextBgColor}
		buf.Set(x, y, cell)
		x++
	}

	return buf
}

func (i *Input) Stream() chan string {
	i.stream = make(chan string, 10)
	return i.stream
}

func (i *Input) KeyPress(e ui.Event) {
	ch := strings.Replace(e.Path, "/sys/kbd/", "", -1)
	if ch == "<space>" {
		ch = " "
	}
	if ch == "C-8" || ch == "<backspace>" || ch == "<delete>" || ch == "C-h" {
		idx := len(i.Data) - 1
		if idx > -1 {
			i.Data = i.Data[0:idx]
			if i.stream != nil {
				select {
				case i.stream <- i.Data:
				default:
				}
			}
		}
		i.render()
		return
	}
	if len(i.Data) >= i.MaxLen {
		return
	}
	if strings.Contains(input_chars, ch) {
		i.Data += ch
		if i.stream != nil {
			select {
			case i.stream <- i.Data:
			default:
			}
		}
		i.render()
	}
}

func (i *Input) render() {
	if tm.IsInit {
		ui.Render(i)
	}
}

// Setup some default handlers for menu navigation
func (i *Input) InputHandlers() {
	ui.Handle("/sys/kbd/", i.KeyPress)
}
