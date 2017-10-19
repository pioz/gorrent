package main

import (
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

// StatusBar stuct
type StatusBar struct {
	*widgets.QStatusBar
	statusMessage *widgets.QLabel
	statusIcon    *widgets.QLabel
	ProgressBar   *widgets.QProgressBar
}

// MakeStatusBar returns a new StatusBar
func MakeStatusBar() *StatusBar {
	sb := new(StatusBar)
	sb.QStatusBar = widgets.NewQStatusBar(nil)
	sb.statusMessage = widgets.NewQLabel(nil, 0)
	sb.statusIcon = widgets.NewQLabel(nil, 0)
	sb.ProgressBar = widgets.NewQProgressBar(nil)

	sb.statusIcon.SetFont(gui.NewQFont2("FontAwesome", 14, 0, false))
	sb.statusIcon.SetContentsMargins(5, 0, 0, 0)
	sb.AddWidget(sb.statusIcon, 0)
	sb.AddWidget(sb.statusMessage, 0)
	sb.ProgressBar.SetMinimum(0)
	sb.ProgressBar.SetMaximum(100)
	sb.ProgressBar.SetFixedWidth(200)
	sb.ProgressBar.Hide()
	sb.AddPermanentWidget(sb.ProgressBar, 0)
	sb.SetStyleSheet("QStatusBar::item { border: 0px}")

	sb.statusMessage.ConnectMouseDoubleClickEvent(func(event *gui.QMouseEvent) {
		sb.ClearStatusMessage()
	})

	return sb
}

// SetStatusMessage set a status message and the icon (font awesome UTF code)
func (sb *StatusBar) SetStatusMessage(message, iconUnicode string) {
	sb.statusMessage.SetText(message)
	sb.statusIcon.SetText(iconUnicode)
}

// SetErrorStatusMessage set a status error
func (sb *StatusBar) SetErrorStatusMessage(message string) {
	sb.SetStatusMessage(message, "\uf071")
	// go func() {
	// 	time.Sleep(time.Second * 5)
	// 	sb.clearStatusMessage()
	// }()
}

// SetOkStatusMessage set a ok status message
func (sb *StatusBar) SetOkStatusMessage(message string) {
	sb.SetStatusMessage(message, "\uf05d")
}

// ClearStatusMessage clear the status message and the icon
func (sb *StatusBar) ClearStatusMessage() {
	sb.statusMessage.SetText("")
	sb.statusIcon.SetText("")
}

// StartProgress start pulse the progress bar
func (sb *StatusBar) StartProgress(pulse bool) {
	sb.ProgressBar.SetValue(0)
	if pulse {
		sb.ProgressBar.SetMaximum(0)
	} else {
		sb.ProgressBar.SetMaximum(100)
	}
	sb.ProgressBar.Show()
}
