package main

import (
	"github.com/therecipe/qt/widgets"
)

// Progress struct
type Progress struct {
	*widgets.QProgressBar
}

// NewProgress create a new ProgressBar
func NewProgress() *Progress {
	p := new(Progress)
	p.QProgressBar = widgets.NewQProgressBar(nil)
	p.SetMinimum(0)
	p.SetMaximum(0)
	return p
}

// Stop progress
func (p *Progress) Stop() {
	p.Hide()
	p.SetValue(0)
	p.SetMaximum(0)
}
