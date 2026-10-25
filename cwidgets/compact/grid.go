package compact

import (
	ui "github.com/gizak/termui"
	tm "github.com/nsf/termbox-go"
)

type CompactGrid struct {
	ui.GridBufferer
	header      *CompactHeader
	cols        []CompactCol // reference columns
	Rows        []RowBufferer
	X, Y        int
	Width       int
	Height      int
	Offset      int // starting row offset
	emptyNotice *ui.Par
}

func NewCompactGrid() *CompactGrid {
	cg := &CompactGrid{header: NewCompactHeader()}
	cg.rebuildHeader()
	return cg
}

func (cg *CompactGrid) Align() {
	y := cg.Y

	if cg.Offset >= len(cg.Rows) || cg.Offset < 0 {
		cg.Offset = 0
	}

	// update row ypos, width recursively
	colWidths := cg.calcWidths()
	for _, r := range cg.pageRows() {
		r.SetY(y)
		y += r.GetHeight()
		r.SetWidths(cg.Width, colWidths)
	}
}

func (cg *CompactGrid) Clear() {
	cg.Rows = []RowBufferer{}
	cg.emptyNotice = nil
	cg.rebuildHeader()
}

func (cg *CompactGrid) SetEmptyNotice(s string) {
	if s == "" {
		cg.emptyNotice = nil
		return
	}
	p := ui.NewPar(s)
	p.X = 2
	p.Border = false
	p.Height = 1
	p.TextFgColor = ui.ThemeAttr("status.warn")
	p.TextBgColor = ui.ThemeAttr("bg")
	cg.emptyNotice = p
}

func (cg *CompactGrid) GetHeight() int { return len(cg.Rows) + cg.header.Height }
func (cg *CompactGrid) SetX(x int)     { cg.X = x }
func (cg *CompactGrid) SetY(y int)     { cg.Y = y }
func (cg *CompactGrid) SetWidth(w int) { cg.Width = w }
func (cg *CompactGrid) MaxRows() int {
	if !tm.IsInit {
		return 0
	}
	max := ui.TermHeight() - cg.header.Height - cg.Y
	if max < 0 {
		return 0
	}
	return max
}

// calculate and return per-column width
func (cg *CompactGrid) calcWidths() []int {
	var autoCols int
	width := cg.Width
	colWidths := make([]int, len(cg.cols))

	for n, w := range cg.cols {
		colWidths[n] = w.FixedWidth()
		width -= w.FixedWidth()
		if w.FixedWidth() == 0 {
			autoCols++
		}
	}

	if autoCols == 0 {
		return colWidths
	}

	spacing := colSpacing * len(cg.cols)
	remWidth := width - spacing
	if remWidth < 0 {
		remWidth = 0
	}
	autoWidth := remWidth / autoCols
	for n, val := range colWidths {
		if val == 0 {
			colWidths[n] = autoWidth
		}
	}
	return colWidths
}

func (cg *CompactGrid) pageRows() (rows []RowBufferer) {
	rows = append(rows, cg.header)
	rows = append(rows, cg.Rows[cg.Offset:]...)
	return rows
}

func (cg *CompactGrid) Buffer() ui.Buffer {
	buf := ui.NewBuffer()
	for _, r := range cg.pageRows() {
		buf.Merge(r.Buffer())
	}
	if len(cg.Rows) == 0 && cg.emptyNotice != nil {
		cg.emptyNotice.Y = cg.Y + cg.header.Height + 1
		width := cg.Width - 4
		if width < 10 {
			width = 10
		}
		cg.emptyNotice.SetWidth(width)
		buf.Merge(cg.emptyNotice.Buffer())
	}
	return buf
}

func (cg *CompactGrid) AddRows(rows ...RowBufferer) {
	cg.Rows = append(cg.Rows, rows...)
}

func (cg *CompactGrid) rebuildHeader() {
	cg.cols = newRowWidgets()
	cg.header.clearFieldPars()
	for _, col := range cg.cols {
		cg.header.addFieldPar(col.Header())
	}
}
