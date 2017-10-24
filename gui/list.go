package gui

import (
	"C"
	"strconv"

	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)
import "github.com/pioz/gorrent/common"

const torrentDownloadURL = int(core.Qt__UserRole + 1)
const torrentMagnetURL = int(core.Qt__UserRole + 2)

// List struct
type List struct {
	core.QObject
	*widgets.QTreeView

	model *gui.QStandardItemModel

	_ func() `signal:"checked"`
	_ func() `signal:"unchecked"`
}

// MakeList new
func MakeList() *List {
	l := NewList(nil)
	l.QTreeView = widgets.NewQTreeView(nil)
	l.init()
	return l
}

func (list *List) init() {
	list.model = gui.NewQStandardItemModel2(0, 4, nil)
	list.model.SetHorizontalHeaderLabels([]string{" ", "Name", "Info", "Seeds"})
	list.SetEditTriggers(widgets.QAbstractItemView__NoEditTriggers)
	list.Header().ConnectSectionClicked(func(index int) {
		checked := 0
		rowCount := list.RowCount()
		if index == 0 {
			for row := 0; row < rowCount; row++ {
				if list.model.Item(row, 0).CheckState() == core.Qt__Checked {
					checked++
				}
			}
			checkState := core.Qt__Checked
			if checked > rowCount/2 {
				checkState = core.Qt__Unchecked
			}
			for row := 0; row < rowCount; row++ {
				list.model.Item(row, 0).SetCheckState(checkState)
			}
		}
	})
	list.model.ConnectItemChanged(func(item *gui.QStandardItem) {
		if item.CheckState() == core.Qt__Checked {
			list.Checked()
		}
		checked := false
		for row := 0; row < list.RowCount(); row++ {
			if list.model.Item(row, 0).CheckState() == core.Qt__Checked {
				checked = true
				break
			}
		}
		if !checked {
			list.Unchecked()
		}
	})
	list.ConnectDoubleClicked(func(index *core.QModelIndex) {
		row := index.Row()
		torrent := list.model.Item(row, 0).Data(torrentMagnetURL).ToString()
		gui.QDesktopServices_OpenUrl(core.QUrl_FromUserInput(torrent))
	})

	list.SetModel(list.model)
	list.SetAlternatingRowColors(true)
	list.Header().SetStretchLastSection(false)
	list.ResizeAllColumnToContents()
	list.Header().SetSectionResizeMode2(2, widgets.QHeaderView__Stretch)
	list.Header().SetSectionsClickable(true)
}

// AddRow method
func (list *List) AddRow(torrent common.Torrent) {
	item1 := gui.NewQStandardItem2(" ")
	item1.SetData(core.NewQVariant14(torrent.Link), torrentDownloadURL)
	item1.SetData(core.NewQVariant14(torrent.Magnet), torrentMagnetURL)
	item1.SetCheckable(true)
	item2 := gui.NewQStandardItem2(torrent.Name)
	item3 := gui.NewQStandardItem2(torrent.Info)
	item4 := gui.NewQStandardItem2(strconv.Itoa(torrent.Seeds))
	i := list.model.RowCount(core.NewQModelIndex())
	list.model.SetItem(i, 0, item1)
	list.model.SetItem(i, 1, item2)
	list.model.SetItem(i, 2, item3)
	list.model.SetItem(i, 3, item4)
}

// RowCount returns the number of rows
func (list *List) RowCount() int {
	return list.model.RowCount(core.NewQModelIndex())
}

// ResizeAllColumnToContents resize all columns to contents
func (list *List) ResizeAllColumnToContents() {
	list.ResizeColumnToContents(0)
	list.ResizeColumnToContents(1)
	list.ResizeColumnToContents(3)
	// for i := 0; i < list.model.ColumnCount(core.NewQModelIndex()); i++ {
	// 	list.ResizeColumnToContents(i)
	// }
}

// RowData returns data for row at index
func (list *List) RowData(index int) common.Torrent {
	link := list.model.Item(index, 0).Data(torrentDownloadURL).ToString()
	magnet := list.model.Item(index, 0).Data(torrentMagnetURL).ToString()
	name := list.model.Item(index, 1).Text()
	info := list.model.Item(index, 2).Text()
	seeds, _ := strconv.Atoi(list.model.Item(index, 3).Text())
	return common.Torrent{link, magnet, name, info, seeds}
}

// RowSelected returns true if the row at index is selected
func (list *List) RowSelected(index int) bool {
	return list.model.Item(index, 0).CheckState() == core.Qt__Checked
}

// RowsSelected returns the number of rows selected
func (list *List) RowsSelected() int {
	counter := 0
	for row := 0; row < list.RowCount(); row++ {
		if list.model.Item(row, 0).CheckState() == core.Qt__Checked {
			counter++
		}
	}
	return counter
}

// UncheckRow uncheck a row at index
func (list *List) UncheckRow(index int) {
	list.model.Item(index, 0).SetCheckState(core.Qt__Unchecked)
}

// RemoveAllRows remove all rows
func (list *List) RemoveAllRows() {
	list.model.RemoveRows(0, list.RowCount(), core.NewQModelIndex())
}
