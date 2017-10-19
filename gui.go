package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/pioz/gorrent/renamer"
	"github.com/pioz/gorrent/scraper"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

// Gui struct
type Gui struct {
	core.QObject

	stopChannel chan bool
	freeze      bool
	settings    *core.QSettings

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

	_ func(err string)           `signal:"errorOccured"`
	_ func()                     `signal:"activityInterrupted"`
	_ func(torrents [][]byte)    `signal:"searchCompleted"`
	_ func()                     `signal:"downloadCompleted"`
	_ func(row int, name string) `signal:"downloadTorrentStarted"`
	_ func(row, percent int)     `signal:"downloadTorrentCompleted"`
	_ func()                     `signal:"renameCompleted"`
	_ func(settingKey string)    `signal:"editSettingsRequested"`
}

// MakeGui returns new Gui struct
func MakeGui() *Gui {
	g := NewGui(nil)
	g.stopChannel = make(chan bool)
	g.freeze = false
	g.settings = core.NewQSettings("pioz", "gorrent", nil)
	// Constructors
	g.window = widgets.NewQMainWindow(nil, 0)
	g.menuBar = widgets.NewQMenuBar(nil)
	g.settingsDialog = MakeSettingsDialog(g.window, g.settings)
	g.toolBar = widgets.NewQToolBar2(g.window)
	g.searchInput = widgets.NewQLineEdit2("Game of Thrones s01", nil)
	g.searchButton = widgets.NewQPushButton2("Search", nil)
	g.downloadButton = widgets.NewQPushButton2("Download torrents", nil)
	g.filterCheckBox = widgets.NewQCheckBox(nil)
	g.filterInput = widgets.NewQLineEdit2(g.settings.Value("search/default_regexp", core.NewQVariant14("")).ToString(), nil)
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
			<p>Developed by <a href="https://github.com/pioz">Pioz</a> in an attempt to learn Go and revise QT</p>`)
	})

	g.window.SetMenuBar(g.menuBar)
}

func (g *Gui) connectEvents() {
	g.settingsDialog.ConnectSettingsSaved(func() {
		g.filterInput.SetText(g.settings.Value("search/default_regexp", core.NewQVariant14("")).ToString())
	})
	g.searchAction.ConnectTriggered(func(bool) { g.searchButton.Click() })
	g.downloadAction.ConnectTriggered(func(bool) { g.downloadButton.Click() })
	g.renameAction.ConnectTriggered(func(checked bool) {
		dirName := openFileDialog(g.settings, "paths/rename", core.QDir_HomePath(), "Select directory with files to rename", true)
		if dirName != "" {
			g.working(true)
			g.statusBar.StartProgress(true)
			g.searchAction.SetDisabled(true)
			g.searchButton.SetDisabled(true)
			g.statusBar.SetStatusMessage("Renaming...", "\uf0c5")
			go func() {
				err := renamer.Rename(dirName, g.settings.Value("tvdb/apikey", core.NewQVariant14("")).ToString(), g.settings.Value("tvdb/locale", core.NewQVariant14("en")).ToString())
				if err != nil {
					if strings.Contains(err.Error(), "401") {
						g.ErrorOccured(err.Error() + ". Is TVDB apikey valid? Try to change it on prefereces dialog.")
						g.EditSettingsRequested("tvdb/apikey")
					} else {
						g.ErrorOccured(err.Error())
					}
					return
				}
				g.RenameCompleted()
			}()
		}
	})

	g.searchInput.ConnectReturnPressed(func() { g.searchButton.Click() })

	g.searchButton.ConnectClicked(func(checked bool) {
		if !g.freeze {
			if g.searchInput.Text() != "" {
				g.working(true)
				g.statusBar.StartProgress(true)
				g.statusBar.SetStatusMessage("Searching '"+g.searchInput.Text()+"'...", "\uf002")
				go func() {
					torrents, err := scraper.RetrieveTorrents(g.searchInput.Text())
					if err != nil {
						g.ErrorOccured(err.Error())
						return
					}
					g.SearchCompleted(torrents)
				}()
			}
		} else {
			g.searchAction.SetDisabled(true)
			g.searchButton.SetDisabled(true)
			go func() {
				g.stopChannel <- true
			}()
		}
	})

	g.downloadButton.ConnectClicked(func(checked bool) {
		dirName := openFileDialog(g.settings, "paths/download", core.QDir_HomePath(), "Save torrent files in", false)
		if dirName != "" {
			g.working(true)
			g.statusBar.StartProgress(false)
			selected := g.list.RowsSelected()
			go func() {
				counter := 0
				for row := 0; row < g.list.RowCount(); row++ {
					select {
					case <-g.stopChannel:
						g.ActivityInterrupted()
						return
					default:
						if g.list.RowSelected(row) {
							counter++
							link, _, name, _, _ := g.list.RowData(row)
							g.DownloadTorrentStarted(row, name)
							g.downloadTorrent(link, name, dirName)
							g.DownloadTorrentCompleted(row, 100*counter/selected)
						}
					}
				}
				time.Sleep(time.Millisecond * 500)
				g.DownloadCompleted()
			}()
		}
	})

	g.filterCheckBox.ConnectClicked(func(checked bool) {
		if checked {
			g.applyFilter(g.filterInput.Text())
		} else {
			for row := 0; row < g.list.RowCount(); row++ {
				g.list.SetRowHidden(row, core.NewQModelIndex(), false)
			}
		}
	})

	g.filterInput.ConnectReturnPressed(func() {
		if g.filterInput.Text() == "" {
			g.filterInput.SetText(g.settings.Value("search/default_regexp", core.NewQVariant14("")).ToString())
		}
		if g.filterCheckBox.IsChecked() {
			g.applyFilter(g.filterInput.Text())
		}
	})

	g.list.ConnectChecked(func() {
		g.downloadButton.SetDisabled(false)
		g.downloadAction.SetDisabled(false)
	})
	g.list.ConnectUnchecked(func() {
		g.downloadButton.SetDisabled(true)
		g.downloadAction.SetDisabled(true)
	})

	g.ConnectErrorOccured(func(err string) {
		g.working(false)
		g.statusBar.ProgressBar.Hide()
		g.statusBar.SetErrorStatusMessage(err)
	})
	g.ConnectActivityInterrupted(func() {
		g.working(false)
		g.statusBar.ProgressBar.Hide()
		g.statusBar.SetStatusMessage("Interrupted.", "\uf05e")
	})
	g.ConnectSearchCompleted(func(torrents [][]byte) {
		select {
		case <-g.stopChannel:
			g.ActivityInterrupted()
		default:
			var t scraper.Torrent
			g.list.RemoveAllRows()
			for i := 0; i < len(torrents); i++ {
				json.Unmarshal(torrents[i], &t)
				g.list.AddRow(t.Link, t.Magnet, t.Name, t.Info, t.Seeds)
			}
			g.list.ResizeAllColumnToContents()
			g.working(false)
			g.statusBar.ProgressBar.Hide()
			if len(torrents) == 0 {
				g.statusBar.SetStatusMessage("No torrents found!", "\uf119")
			}
		}
	})
	g.ConnectDownloadTorrentStarted(func(row int, name string) {
		g.statusBar.SetStatusMessage("Downloading torrent '"+name+"'", "\uf019")
	})
	g.ConnectDownloadTorrentCompleted(func(row, percent int) {
		g.list.UncheckRow(row)
		g.statusBar.ProgressBar.SetValue(percent)
	})
	g.ConnectDownloadCompleted(func() {
		g.working(false)
		g.statusBar.ProgressBar.Hide()
		g.statusBar.SetOkStatusMessage("All torrent files downloaded!")
	})
	g.ConnectRenameCompleted(func() {
		g.working(false)
		g.statusBar.ProgressBar.Hide()
		g.statusBar.SetOkStatusMessage("Files have been renamed successfully!")
	})
	g.ConnectEditSettingsRequested(func(settingKey string) {
		g.settingsDialog.Exec(settingKey)
		g.statusBar.ClearStatusMessage()
	})
}

func (g *Gui) working(freeze bool) {
	g.freeze = freeze
	g.searchButton.SetChecked(freeze)
	g.searchInput.SetDisabled(freeze)
	g.downloadButton.SetDisabled(freeze)
	g.downloadAction.SetDisabled(freeze)
	g.renameAction.SetDisabled(freeze)
	g.filterCheckBox.SetDisabled(freeze)
	g.filterInput.SetDisabled(freeze)
	g.list.SetDisabled(freeze)
	if freeze {
		g.filterCheckBox.SetChecked(false)
		g.searchButton.SetText("Stop")
		g.searchAction.SetText("&Stop")
	} else {
		g.searchButton.SetDisabled(false)
		g.searchAction.SetDisabled(false)
		g.searchButton.SetText("Search")
		g.searchAction.SetText("&Search")
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

func (g *Gui) downloadTorrent(link, name, destDir string) {
	file, err := os.Create(destDir + "/" + name + ".torrent")
	if err != nil {
		g.ErrorOccured(err.Error())
		return
	}
	defer file.Close()
	resp, err := http.Get(link)
	if err != nil {
		g.ErrorOccured(err.Error())
		return
	}
	defer resp.Body.Close()
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		g.ErrorOccured(err.Error())
		return
	}
}

func (g *Gui) applyFilter(filter string) {
	regexp, err := regexp.Compile("(?i)" + g.filterInput.Text())
	if err != nil {
		g.statusBar.SetErrorStatusMessage("Invalid filter regexp")
	} else {
		for row := 0; row < g.list.RowCount(); row++ {
			_, _, _, info, _ := g.list.RowData(row)
			g.list.SetRowHidden(row, core.NewQModelIndex(), !regexp.MatchString(info))
		}
	}
}

// SyncSettings sync the app settings
func (g *Gui) SyncSettings() {
	g.settings.SetValue("window/size", core.NewQVariant27(g.window.Size()))
	g.settings.SetValue("window/position", core.NewQVariant29(g.window.Pos()))
	g.settings.Sync()
}

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
