package gui

import (
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
)

// QVFormLayout struct
type QVFormLayout struct {
	*widgets.QVBoxLayout
	blocks        []*widgets.QFormLayout
	labels        []*widgets.QLabel
	maxLabelWidth int
}

// MakeQVFormLayout returns a new QVFormLayout
func MakeQVFormLayout() *QVFormLayout {
	layout := new(QVFormLayout)
	layout.QVBoxLayout = widgets.NewQVBoxLayout()
	layout.maxLabelWidth = 0
	return layout
}

// AddBlock add a group box
func (layout *QVFormLayout) AddBlock(title string) {
	gb := widgets.NewQGroupBox2(title, nil)
	fl := widgets.NewQFormLayout(nil)
	gb.SetLayout(fl)
	layout.QVBoxLayout.AddWidget(gb, 0, core.Qt__AlignCenter)
	layout.blocks = append(layout.blocks, fl)
}

// AddRow add a new row
func (layout *QVFormLayout) AddRow(block int, label string, widget widgets.QWidget_ITF) {
	if block < len(layout.blocks) {
		l := widgets.NewQLabel2(label, nil, core.Qt__Window)
		l.SetAlignment(core.Qt__AlignRight)
		layout.blocks[block].AddRow(l, widget)
		layout.labels = append(layout.labels, l)
		w := l.FontMetrics().Width(l.Text(), len(l.Text()))
		if w > layout.maxLabelWidth {
			layout.maxLabelWidth = w
		}
		layout.resizeAllLabels()
	}

}

func (layout *QVFormLayout) resizeAllLabels() {
	for i := 0; i < len(layout.labels); i++ {
		layout.labels[i].SetMinimumWidth(layout.maxLabelWidth)
	}
}
