package gui

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/pioz/gorrent/common"
	t "github.com/pioz/gorrent/i18n"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

// Gui struct
type Gui struct {
	core.QObject

	freeze   bool
	settings *core.QSettings

	window         *widgets.QMainWindow
	menuBar        *widgets.QMenuBar
	searchAction   *widgets.QAction
	downloadAction *widgets.QAction
	renameAction   *widgets.QAction
	settingsDialog *SettingsDialog
	toolBar        *widgets.QToolBar
	searchInput    *widgets.QLineEdit
	searchButton   *widgets.QPushButton
	downloadButton *widgets.QPushButton
	filterCheckBox *widgets.QCheckBox
	filterInput    *widgets.QLineEdit
	list           *List
	statusBar      *StatusBar

	_ func(string)                 `signal:"searchRequested"`
	_ func(map[int][]byte, string) `signal:"downloadTorrentRequested"`
	_ func(string)                 `signal:"renameSeriesRequested"`
	_ func()                       `signal:"settingsSaved"`
}

// MakeGui returns new Gui struct
func MakeGui(settings *core.QSettings) *Gui {
	g := NewGui(nil)
	g.freeze = false
	g.settings = settings
	// Constructors
	g.window = widgets.NewQMainWindow(nil, 0)
	g.menuBar = widgets.NewQMenuBar(nil)
	g.settingsDialog = MakeSettingsDialog(g.window, g.settings)
	g.toolBar = widgets.NewQToolBar2(g.window)
	g.searchInput = widgets.NewQLineEdit2("Game of Thrones s01", nil)
	g.searchButton = widgets.NewQPushButton2(t.T("search"), nil)
	g.downloadButton = widgets.NewQPushButton2("Download torrents", nil)
	g.filterCheckBox = widgets.NewQCheckBox(nil)
	g.filterInput = widgets.NewQLineEdit2(g.settings.Value("gorrent/default_regexp", core.NewQVariant14("")).ToString(), nil)
	g.list = MakeList()
	g.statusBar = MakeStatusBar()
	// Custom init
	g.initMenuBar()
	g.connectEvents()
	// Setup toolBar widgets
	g.searchInput.SetFocus2()
	g.toolBar.AddWidget(g.searchInput)
	g.toolBar.AddWidget(g.searchButton)
	g.downloadButton.SetDisabled(true)
	g.downloadButton.SetDisabled(true)
	g.toolBar.AddWidget(g.downloadButton)
	g.filterCheckBox.SetDisabled(true)
	g.filterCheckBox.SetFont(gui.NewQFont2("FontAwesome", 14, 0, false))
	g.filterCheckBox.SetText("\uf0b0")
	g.filterCheckBox.SetLayoutDirection(core.Qt__RightToLeft)
	g.toolBar.AddWidget(g.filterCheckBox)
	g.filterInput.SetDisabled(true)
	g.filterInput.SetFixedWidth(150)
	g.filterInput.SetSizePolicy2(widgets.QSizePolicy__Fixed, widgets.QSizePolicy__Fixed)
	g.toolBar.AddWidget(g.filterInput)
	g.window.AddToolBar2(g.toolBar)
	// Setup central widgets
	scrollArea := widgets.NewQScrollArea(g.window)
	scrollArea.SetWidget(g.list)
	scrollArea.SetWidgetResizable(true)
	g.window.SetCentralWidget(scrollArea)
	// Setup statusBar widgets
	g.window.SetStatusBar(g.statusBar.QStatusBar)
	// Setup main window
	g.window.SetWindowTitle("Gorrent")
	g.window.SetMinimumSize2(800, 600)
	size := g.settings.Value("window/size", core.NewQVariant27(core.NewQSize2(800, 600))).ToSize()
	g.window.Resize(size)
	screen := gui.QGuiApplication_PrimaryScreen()
	w := screen.Geometry().Width()
	h := screen.Geometry().Height()
	x := w/2 - size.Rwidth()/2
	y := h/2 - size.Rheight()/2
	pos := g.settings.Value("window/position", core.NewQVariant29(core.NewQPoint2(x, y))).ToPoint()
	g.window.Move(pos)
	g.window.SetWindowIcon(gui.NewQIcon5(":/donkey.png"))
	g.window.Show()

	return g
}

func (g *Gui) initMenuBar() {
	g.searchAction = widgets.NewQAction2("&Search", g.menuBar)
	g.downloadAction = widgets.NewQAction2("&Download torrents", g.menuBar)
	g.downloadAction.SetDisabled(true)
	g.renameAction = widgets.NewQAction2("&Rename series files", g.menuBar)
	settingsAction := widgets.NewQAction2("&Settings", g.menuBar)
	quitAction := widgets.NewQAction2("&Quit", g.menuBar)
	aboutAction := widgets.NewQAction2("&About", g.menuBar)

	g.searchAction.SetShortcut(gui.QKeySequence_FromString("Ctrl+S", 0))
	g.downloadAction.SetShortcut(gui.QKeySequence_FromString("Ctrl+D", 0))
	g.renameAction.SetShortcut(gui.QKeySequence_FromString("Ctrl+R", 0))
	settingsAction.SetShortcuts2(gui.QKeySequence__Preferences)
	quitAction.SetShortcuts2(gui.QKeySequence__Quit)

	fileMenu := g.menuBar.AddMenu2("&File")
	fileMenu.AddActions([]*widgets.QAction{g.searchAction, g.downloadAction})
	fileMenu.AddSeparator()
	fileMenu.AddActions([]*widgets.QAction{quitAction})
	settingsMenu := g.menuBar.AddMenu2("&Tools")
	settingsMenu.AddActions([]*widgets.QAction{g.renameAction})
	settingsMenu.AddSeparator()
	settingsMenu.AddActions([]*widgets.QAction{settingsAction})
	helpMenu := g.menuBar.AddMenu2("&Help")
	helpMenu.AddActions([]*widgets.QAction{aboutAction})

	settingsAction.ConnectTriggered(func(bool) { g.settingsDialog.Exec("") })
	quitAction.ConnectTriggered(func(bool) { g.window.Close() })
	aboutAction.ConnectTriggered(func(bool) {
		widgets.QMessageBox_About(g.window, "Gorrent",
			`<p><h1>Gorrent</h1> version 0.1.0</p>
			<p>Developed by <a href="https://github.com/pioz">Pioz</a> in an attempt to learn Go and brush up QT</p>`)
	})

	g.window.SetMenuBar(g.menuBar)
}

func (g *Gui) connectEvents() {
	g.settingsDialog.ConnectSettingsSaved(func() {
		g.filterInput.SetText(g.settings.Value("gorrent/default_regexp", core.NewQVariant14("")).ToString())
		g.SettingsSaved()
	})

	g.searchAction.ConnectTriggered(func(bool) { g.searchButton.Click() })
	g.downloadAction.ConnectTriggered(func(bool) { g.downloadButton.Click() })
	g.renameAction.ConnectTriggered(g.onRenameSeries)
	g.searchInput.ConnectReturnPressed(func() { g.searchButton.Click() })
	g.searchButton.ConnectClicked(g.search)
	g.downloadButton.ConnectClicked(g.downloadTorrents)
	g.filterCheckBox.ConnectClicked(g.filter)
	g.filterInput.ConnectReturnPressed(g.filterReturnPressed)

	g.list.ConnectChecked(func() {
		g.downloadButton.SetDisabled(false)
		g.downloadAction.SetDisabled(false)
	})
	g.list.ConnectUnchecked(func() {
		g.downloadButton.SetDisabled(true)
		g.downloadAction.SetDisabled(true)
	})
}

func (g *Gui) working(freeze bool) {
	g.freeze = freeze
	g.searchButton.SetDisabled(freeze)
	g.searchAction.SetDisabled(freeze)
	g.searchInput.SetDisabled(freeze)
	g.downloadButton.SetDisabled(freeze)
	g.downloadAction.SetDisabled(freeze)
	g.renameAction.SetDisabled(freeze)
	g.filterCheckBox.SetDisabled(freeze)
	g.filterInput.SetDisabled(freeze)
	g.list.SetDisabled(freeze)
	if freeze {
		g.filterCheckBox.SetChecked(false)
	} else {
		downloadButtonState := g.list.RowsSelected() > 0
		g.downloadButton.SetEnabled(downloadButtonState)
		g.downloadAction.SetEnabled(downloadButtonState)
		if g.list.RowCount() == 0 {
			g.filterCheckBox.SetDisabled(true)
			g.filterInput.SetDisabled(true)
		}
		g.statusBar.ClearStatusMessage()
		g.searchInput.SetFocus2()
		g.searchInput.SetSelection(0, len(g.searchInput.Text()))
	}
}

func (g *Gui) applyFilter(filter string) {
	regexp, err := regexp.Compile("(?i)" + g.filterInput.Text())
	if err != nil {
		g.statusBar.SetErrorStatusMessage("Invalid filter regexp")
	} else {
		for row := 0; row < g.list.RowCount(); row++ {
			t := g.list.RowData(row)
			g.list.SetRowHidden(row, core.NewQModelIndex(), !regexp.MatchString(t.Info))
		}
	}
}

// SyncSettings sync the app settings
func (g *Gui) SyncSettings() {
	g.settings.SetValue("window/size", core.NewQVariant27(g.window.Size()))
	g.settings.SetValue("window/position", core.NewQVariant29(g.window.Pos()))
	g.settings.Sync()
}

// Slots

// ErrorOccured slot
func (g *Gui) ErrorOccured(err string) {
	g.working(false)
	g.statusBar.ProgressBar.Hide()
	if strings.Contains(err, "401") {
		g.statusBar.SetErrorStatusMessage(err + ": is your TVDB apikey valid?")
		g.settingsDialog.Exec("tvdb/apikey")
	} else {
		g.statusBar.SetErrorStatusMessage(err)
	}
}

func (g *Gui) onRenameSeries(checked bool) {
	dirName := openFileDialog(g.settings, "paths/rename", core.QDir_HomePath(), "Select directory with files to rename", true)
	if dirName != "" {
		g.working(true)
		g.statusBar.StartProgress(true)
		g.statusBar.SetStatusMessage("Renaming...", "\uf0c5")
		g.RenameSeriesRequested(dirName)
	}
}

// RenameSeriesCompleted slot
func (g *Gui) RenameSeriesCompleted() {
	g.working(false)
	g.statusBar.ProgressBar.Hide()
	g.statusBar.SetOkStatusMessage("Files have been renamed successfully!")
}

func (g *Gui) search(checked bool) {
	if g.searchInput.Text() != "" {
		g.working(true)
		g.statusBar.StartProgress(true)
		g.statusBar.SetStatusMessage("Searching '"+g.searchInput.Text()+"'...", "\uf002")
		g.SearchRequested(g.searchInput.Text())
	}
}

// SearchCompleted slot
func (g *Gui) SearchCompleted(torrents [][]byte) {
	var t common.Torrent
	g.list.RemoveAllRows()
	for i := 0; i < len(torrents); i++ {
		json.Unmarshal(torrents[i], &t)
		g.list.AddRow(t)
	}
	g.list.ResizeAllColumnToContents()
	g.working(false)
	g.statusBar.ProgressBar.Hide()
	if len(torrents) == 0 {
		g.statusBar.SetStatusMessage("No torrents found!", "\uf119")
	}
}

func (g *Gui) downloadTorrents(checked bool) {
	dirName := openFileDialog(g.settings, "paths/download", core.QDir_HomePath(), "Save torrent files in", false)
	if dirName != "" {
		g.working(true)
		g.statusBar.StartProgress(false)
		torrents := make(map[int][]byte)
		for row := 0; row < g.list.RowCount(); row++ {
			if g.list.RowSelected(row) {
				json, _ := json.Marshal(g.list.RowData(row))
				torrents[row] = json
			}
		}
		g.DownloadTorrentRequested(torrents, dirName)
	}
}

// DownloadTorrentStarted slot
func (g *Gui) DownloadTorrentStarted(name string) {
	g.statusBar.SetStatusMessage("Downloading torrent '"+name+"'", "\uf019")
}

// DownloadTorrentCompleted slot
func (g *Gui) DownloadTorrentCompleted(row int, percent int) {
	g.list.UncheckRow(row)
	g.statusBar.ProgressBar.SetValue(percent)
}

// DownloadTorrentsCompleted slot
func (g *Gui) DownloadTorrentsCompleted() {
	g.working(false)
	g.statusBar.ProgressBar.Hide()
	g.statusBar.SetOkStatusMessage("All torrent files downloaded!")
}

func (g *Gui) filter(checked bool) {
	if checked {
		g.applyFilter(g.filterInput.Text())
	} else {
		for row := 0; row < g.list.RowCount(); row++ {
			g.list.SetRowHidden(row, core.NewQModelIndex(), false)
		}
	}
}

func (g *Gui) filterReturnPressed() {
	if g.filterInput.Text() == "" {
		g.filterInput.SetText(g.settings.Value("gorrent/default_regexp", core.NewQVariant14("")).ToString())
	}
	if g.filterCheckBox.IsChecked() {
		g.applyFilter(g.filterInput.Text())
	}
}

// utils

func openFileDialog(settings *core.QSettings, settingKey, defaultDir, title string, cdUp bool) string {
	openDir := core.NewQDir2(settings.Value(settingKey, core.NewQVariant14(defaultDir)).ToString())
	if !openDir.Exists2() {
		openDir = core.QDir_Home()
	}
	dirName := widgets.QFileDialog_GetExistingDirectory(nil, title, openDir.AbsolutePath(), widgets.QFileDialog__ReadOnly)
	if dirName == "" {
		return ""
	}
	path := core.NewQDir2(dirName)
	if cdUp {
		path.CdUp()
	}
	if path.AbsolutePath() != openDir.AbsolutePath() {
		settings.SetValue(settingKey, core.NewQVariant14(path.AbsolutePath()))
	}
	return dirName
}
