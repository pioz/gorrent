package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"regexp"
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

	window         *widgets.QMainWindow
	menuBar        *widgets.QMenuBar
	searchAction   *widgets.QAction
	downloadAction *widgets.QAction
	renameAction   *widgets.QAction
	toolBar        *widgets.QToolBar
	searchInput    *widgets.QLineEdit
	searchButton   *widgets.QPushButton
	downloadButton *widgets.QPushButton
	filterCheckBox *widgets.QCheckBox
	filterInput    *widgets.QLineEdit
	list           *List
	statusBar      *widgets.QStatusBar
	statusMessage  *widgets.QLabel
	statusIcon     *widgets.QLabel
	progressBar    *widgets.QProgressBar

	_ func(err error)            `signal:"errorOccured"`
	_ func()                     `signal:"activityInterrupted"`
	_ func(torrents [][]byte)    `signal:"searchCompleted"`
	_ func()                     `signal:"downloadCompleted"`
	_ func(row int, name string) `signal:"downloadTorrentStarted"`
	_ func(row, percent int)     `signal:"downloadTorrentCompleted"`
	_ func()                     `signal:"RenameCompleted"`
}

const defaultRegexp = `ita\s(eng\s)?(mp3|ac3)`

// MakeGui returns new Gui struct
func MakeGui() *Gui {
	g := NewGui(nil)
	g.stopChannel = make(chan bool)
	g.freeze = false
	// Constructors
	g.window = widgets.NewQMainWindow(nil, 0)
	g.menuBar = widgets.NewQMenuBar(nil)
	g.toolBar = widgets.NewQToolBar2(g.window)
	g.searchInput = widgets.NewQLineEdit2("Game of Thrones s01", nil)
	g.searchButton = widgets.NewQPushButton2("Search", nil)
	g.downloadButton = widgets.NewQPushButton2("Download torrents", nil)
	g.filterCheckBox = widgets.NewQCheckBox(nil)
	g.filterInput = widgets.NewQLineEdit2(defaultRegexp, nil)
	g.list = MakeList()
	g.statusBar = widgets.NewQStatusBar(g.window)
	g.statusMessage = widgets.NewQLabel(nil, 0)
	g.statusIcon = widgets.NewQLabel(nil, 0)
	g.progressBar = widgets.NewQProgressBar(nil)
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
	g.statusIcon.SetFont(gui.NewQFont2("FontAwesome", 14, 0, false))
	g.statusIcon.SetContentsMargins(5, 0, 0, 0)
	g.statusBar.AddWidget(g.statusIcon, 0)
	g.statusBar.AddWidget(g.statusMessage, 0)
	g.progressBar.SetMinimum(0)
	g.progressBar.SetMaximum(100)
	g.progressBar.SetFixedWidth(200)
	g.progressBar.Hide()
	g.statusBar.AddPermanentWidget(g.progressBar, 0)
	g.statusBar.SetStyleSheet("QStatusBar::item { border: 0px}")
	g.window.SetStatusBar(g.statusBar)
	// Setup main window
	g.window.SetWindowTitle("Gorrent")
	g.window.SetMinimumSize2(800, 600)
	g.window.SetWindowIcon(gui.NewQIcon5(":/donkey.png"))
	g.window.Show()

	return g
}

func (g *Gui) initMenuBar() {
	g.searchAction = widgets.NewQAction2("&Search", g.menuBar)
	g.downloadAction = widgets.NewQAction2("&Download torrents", g.menuBar)
	g.downloadAction.SetDisabled(true)
	g.renameAction = widgets.NewQAction2("&Rename series files", g.menuBar)
	quitAction := widgets.NewQAction2("&Quit", g.menuBar)
	aboutAction := widgets.NewQAction2("&About", g.menuBar)

	g.searchAction.SetShortcut(gui.QKeySequence_FromString("Ctrl+S", 0))
	g.downloadAction.SetShortcut(gui.QKeySequence_FromString("Ctrl+D", 0))
	g.renameAction.SetShortcut(gui.QKeySequence_FromString("Ctrl+R", 0))
	quitAction.SetShortcuts2(gui.QKeySequence__Quit)

	fileMenu := g.menuBar.AddMenu2("&File")
	fileMenu.AddActions([]*widgets.QAction{g.searchAction, g.downloadAction})
	fileMenu.AddSeparator()
	fileMenu.AddActions([]*widgets.QAction{g.renameAction})
	fileMenu.AddSeparator()
	fileMenu.AddActions([]*widgets.QAction{quitAction})
	helpMenu := g.menuBar.AddMenu2("&Help")
	helpMenu.AddActions([]*widgets.QAction{aboutAction})

	quitAction.ConnectTriggered(func(bool) { g.window.Close() })
	aboutAction.ConnectTriggered(func(bool) {
		widgets.QMessageBox_About(g.window, "Gorrent",
			`<p><h1>Gorrent</h1> version 0.1.0</p>
			<p>Developed by <a href="https://github.com/pioz">Pioz</a> in an attempt to learn Go and revise QT</p>`)
	})

	g.window.SetMenuBar(g.menuBar)
}

func (g *Gui) connectEvents() {
	g.searchAction.ConnectTriggered(func(bool) { g.searchButton.Click() })
	g.downloadAction.ConnectTriggered(func(bool) { g.downloadButton.Click() })
	g.renameAction.ConnectTriggered(func(checked bool) {
		dirName := widgets.QFileDialog_GetExistingDirectory(nil, "Select directory with files to rename", core.QDir_HomePath(), widgets.QFileDialog__ReadOnly)
		g.working(true)
		g.startProgress(true)
		g.searchAction.SetDisabled(true)
		g.searchButton.SetDisabled(true)
		g.setStatusMessage("Renaming...", "\uf0c5")
		go func() {
			err := renamer.Rename(dirName)
			if err != nil {
				g.ErrorOccured(err)
				return
			}
			g.RenameCompleted()
		}()
	})

	g.searchInput.ConnectReturnPressed(func() { g.searchButton.Click() })

	g.searchButton.ConnectClicked(func(checked bool) {
		if !g.freeze {
			if g.searchInput.Text() != "" {
				g.working(true)
				g.startProgress(true)
				g.setStatusMessage("Searching '"+g.searchInput.Text()+"'...", "\uf002")
				go func() {
					torrents, err := scraper.RetrieveTorrents(g.searchInput.Text())
					if err != nil {
						g.ErrorOccured(err)
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
		path := core.QDir_HomePath() + "/Desktop"
		if !core.NewQDir2(path).Exists2() {
			path = core.QDir_HomePath()
		}
		dirName := widgets.QFileDialog_GetExistingDirectory(nil, "Save torrent files", path, widgets.QFileDialog__ReadOnly)
		g.working(true)
		g.startProgress(false)
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
			g.filterInput.SetText(defaultRegexp)
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

	g.ConnectErrorOccured(func(err error) {
		g.working(false)
		g.progressBar.Hide()
		g.setErrorStatusMessage(err.Error())
	})
	g.ConnectActivityInterrupted(func() {
		g.working(false)
		g.progressBar.Hide()
		g.setStatusMessage("Interrupted.", "\uf05e")
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
			g.progressBar.Hide()
		}
	})
	g.ConnectDownloadTorrentStarted(func(row int, name string) {
		g.setStatusMessage("Downloading torrent '"+name+"'", "\uf019")
	})
	g.ConnectDownloadTorrentCompleted(func(row, percent int) {
		g.list.UncheckRow(row)
		g.progressBar.SetValue(percent)
	})
	g.ConnectDownloadCompleted(func() {
		g.working(false)
		g.progressBar.Hide()
		g.setOkStatusMessage("All torrent files downloaded!")
	})
	g.ConnectRenameCompleted(func() {
		g.working(false)
		g.progressBar.Hide()
		g.setOkStatusMessage("Files renamed successfully!")
	})

}

func (g *Gui) setStatusMessage(message, iconUnicode string) {
	g.statusMessage.SetText(message)
	g.statusIcon.SetText(iconUnicode)
}

func (g *Gui) setErrorStatusMessage(message string) {
	g.setStatusMessage(message, "\uf071")
	// go func() {
	// 	time.Sleep(time.Second * 5)
	// 	g.clearStatusMessage()
	// }()
}

func (g *Gui) setOkStatusMessage(message string) {
	g.setStatusMessage(message, "\uf05d")
	// go func() {
	// 	time.Sleep(time.Second * 5)
	// 	g.clearStatusMessage()
	// }()
}

func (g *Gui) clearStatusMessage() {
	g.statusMessage.SetText("")
	g.statusIcon.SetText("")
}

func (g *Gui) startProgress(pulse bool) {
	g.progressBar.SetValue(0)
	if pulse {
		g.progressBar.SetMaximum(0)
	} else {
		g.progressBar.SetMaximum(100)
	}
	g.progressBar.Show()
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
		g.clearStatusMessage()
		g.searchInput.SetFocus2()
		g.searchInput.SetSelection(0, len(g.searchInput.Text()))
	}
}

func (g *Gui) downloadTorrent(link, name, destDir string) {
	file, err := os.Create(destDir + "/" + name + ".torrent")
	if err != nil {
		g.ErrorOccured(err)
		return
	}
	defer file.Close()
	resp, err := http.Get(link)
	if err != nil {
		g.ErrorOccured(err)
		return
	}
	defer resp.Body.Close()
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		g.ErrorOccured(err)
		return
	}
}

func (g *Gui) applyFilter(filter string) {
	regexp, err := regexp.Compile("(?i)" + g.filterInput.Text())
	if err != nil {
		g.setErrorStatusMessage("Invalid filter regexp")
	} else {
		for row := 0; row < g.list.RowCount(); row++ {
			_, _, _, info, _ := g.list.RowData(row)
			g.list.SetRowHidden(row, core.NewQModelIndex(), !regexp.MatchString(info))
		}
	}
}
