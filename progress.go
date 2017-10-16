package main

import (
	"time"

	"github.com/therecipe/qt/core"
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
	p.SetMaximum(100)
	return p
}

// Start progress
func (p *Progress) Start() {
	p.Show()
	go func() {
		var amount int
		for p.IsVisible() {
			time.Sleep(time.Millisecond * 10)
			if p.Value() == 0 {
				p.SetLayoutDirection(core.Qt__LeftToRight)
			} else if p.Value() == 100 {
				p.SetLayoutDirection(core.Qt__RightToLeft)
			}
			if p.LayoutDirection() == core.Qt__LeftToRight {
				amount = 1
			} else {
				amount = -1
			}
			p.SetValue(p.Value() + amount)
		}
		p.SetValue(0)
	}()
}

// Stop progress
func (p *Progress) Stop() {
	p.Hide()
	p.SetLayoutDirection(core.Qt__LeftToRight)
	p.SetValue(0)
}
