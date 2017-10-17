package main

import (
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
	stopChannel    chan bool
	freeze         bool
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
	progressBar    *Progress
}

const defaultRegexp = `ita\s(eng\s)?(mp3|ac3)`

// NewGui returns new Gui struct
func NewGui() *Gui {
	g := new(Gui)
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
	g.progressBar = NewProgress()
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
	g.searchInput.ConnectReturnPressed(func() { g.search() })

	g.searchButton.ConnectClicked(func(checked bool) {
		if !g.freeze {
			g.search()
		} else {
			g.stopSearch()
		}
	})

	g.downloadButton.ConnectClicked(func(checked bool) {
		path := core.QDir_HomePath() + "/Desktop"
		if !core.NewQDir2(path).Exists2() {
			path = core.QDir_HomePath()
		}
		dirName := widgets.QFileDialog_GetExistingDirectory(nil, "Save torrent files", path, widgets.QFileDialog__ReadOnly)
		g.working(true)
		g.progressBar.SetMaximum(100)
		g.progressBar.Show()
		selected := g.list.RowsSelected()
		go func() {
			counter := 0
			for row := 0; row < g.list.RowCount(); row++ {
				if g.list.RowSelected(row) {
					counter++
					link, _, name, _, _ := g.list.RowData(row)
					g.downloadTorrent(link, name, dirName)
					g.list.UncheckRow(row)
					g.progressBar.SetValue(100 * counter / selected)
				}
			}
			time.Sleep(time.Millisecond * 500)
			g.working(false)
			g.progressBar.Stop()
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

	g.searchAction.ConnectTriggered(func(bool) { g.searchButton.Click() })
	g.downloadAction.ConnectTriggered(func(bool) { g.downloadButton.Click() })
	g.renameAction.ConnectTriggered(func(checked bool) {
		dirName := widgets.QFileDialog_GetExistingDirectory(nil, "Select directory with files to rename", core.QDir_HomePath(), widgets.QFileDialog__ReadOnly)
		g.working(true)
		g.setStatusMessage("Renaming...", "\uf0c5")
		g.progressBar.Show()
		go func() {
			err := renamer.Rename(dirName)
			if err != nil {
				g.working(false)
				g.progressBar.Stop()
				g.setErrorStatusMessage("Impossible rename files: " + err.Error())
				return
			}
			g.working(false)
			g.progressBar.Stop()
			g.setOkStatusMessage("Files renamed successfully!")
		}()
	})

}

func (g *Gui) setStatusMessage(message, iconUnicode string) {
	g.statusMessage.SetText(message)
	g.statusIcon.SetText(iconUnicode)
}

func (g *Gui) setErrorStatusMessage(message string) {
	g.setStatusMessage(message, "\uf071")
	go func() {
		time.Sleep(time.Second * 5)
		g.clearStatusMessage()
	}()
}

func (g *Gui) setOkStatusMessage(message string) {
	g.setStatusMessage(message, "\uf05d")
	go func() {
		time.Sleep(time.Second * 5)
		g.clearStatusMessage()
	}()
}

func (g *Gui) clearStatusMessage() {
	g.statusMessage.SetText("")
	g.statusIcon.SetText("")
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

func (g *Gui) search() {
	if g.searchInput.Text() != "" {
		g.working(true)
		g.progressBar.Show()
		g.setStatusMessage("Searching '"+g.searchInput.Text()+"'...", "\uf002")
		go func() {
			torrents, err := scraper.RetrieveTorrents(g.searchInput.Text())
			g.searchButton.SetDisabled(true)
			g.searchAction.SetDisabled(true)
			select {
			case <-g.stopChannel:
				g.working(false)
				g.progressBar.Stop()
				return
			default:
				if err != nil {
					g.working(false)
					g.progressBar.Stop()
					g.handleError(err)
					return
				}
				g.list.RemoveAllRows()
				for i := 0; i < len(torrents); i++ {
					g.list.AddRow(torrents[i].Link, torrents[i].Magnet, torrents[i].Name, torrents[i].Info, torrents[i].Seeds)
				}
				g.list.ResizeAllColumnToContents()
				g.working(false)
				g.progressBar.Stop()
			}
		}()
	}
}

func (g *Gui) stopSearch() {
	go func() {
		g.stopChannel <- true
	}()
}

func (g *Gui) downloadTorrent(link, name, destDir string) {
	g.setStatusMessage("Downloading torrent '"+name+"'", "\uf019")
	file, err := os.Create(destDir + "/" + name + ".torrent")
	if g.handleError(err) {
		defer file.Close()
		resp, err := http.Get(link)
		if g.handleError(err) {
			defer resp.Body.Close()
			_, err := io.Copy(file, resp.Body)
			if g.handleError(err) {
				g.clearStatusMessage()
			}
		}
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

func (g *Gui) handleError(err error) bool {
	if err != nil {
		g.setErrorStatusMessage(err.Error())
		return false
	}
	return true
}
