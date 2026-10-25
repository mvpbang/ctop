package compact

import (
	"testing"
)

type dummyCol struct {
	CompactCol
	width int
}

func (d *dummyCol) FixedWidth() int { return d.width }

func TestCalcWidthsNoAutoCols(t *testing.T) {
	cg := &CompactGrid{
		Width: 100,
		cols: []CompactCol{
			&dummyCol{width: 30},
			&dummyCol{width: 40},
		},
	}

	widths := cg.calcWidths()
	if len(widths) != 2 || widths[0] != 30 || widths[1] != 40 {
		t.Errorf("unexpected widths: %v", widths)
	}
}

func TestCompactGridEmptyNotice(t *testing.T) {
	cg := NewCompactGrid()
	cg.SetEmptyNotice("No containers match filter 'foo'")
	buf := cg.Buffer()
	found := false
	for _, cell := range buf.CellMap {
		if cell.Ch == 'N' {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected empty notice to be rendered in buffer")
	}
}
