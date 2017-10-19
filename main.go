package main

import (
	"fmt"
	"os"

	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

func main() {
	app := widgets.NewQApplication(len(os.Args), os.Args)
	app.SetApplicationName("gorrent")
	app.SetApplicationDisplayName("Gorrent")
	app.SetAttribute(core.Qt__AA_UseHighDpiPixmaps, true)
	if gui.QFontDatabase_AddApplicationFont(":/FontAwesome.otf") < 0 {
		fmt.Println("Impossible to load FontAwesome")
	}
	gui := MakeGui()
	app.ConnectLastWindowClosed(func() {
		gui.SyncSettings()
	})
	app.Exec()
}
