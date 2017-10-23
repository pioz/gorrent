package main

import (
	"fmt"
	"os"

	"github.com/nicksnyder/go-i18n/i18n"
	g "github.com/pioz/gorrent/gui"
	t "github.com/pioz/gorrent/i18n"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

func main() {
	settings := core.NewQSettings("pioz", "gorrent", nil)
	locale := settings.Value("gorrent/locale", core.NewQVariant14("en")).ToString()
	t.LoadI18nFile(locale)
	t.T, _ = i18n.Tfunc(locale)

	app := widgets.NewQApplication(len(os.Args), os.Args)
	app.SetApplicationName("gorrent")
	app.SetApplicationDisplayName("Gorrent")
	app.SetAttribute(core.Qt__AA_UseHighDpiPixmaps, true)
	if gui.QFontDatabase_AddApplicationFont(":/FontAwesome.otf") < 0 {
		fmt.Println("Impossible to load FontAwesome")
	}

	// translator1 := core.NewQTranslator(nil)
	// translator1.Load2(core.QLocale_System(), "qt_it", "",
	// 	core.QLibraryInfo_Location(core.QLibraryInfo__TranslationsPath), ".qm")
	// app.InstallTranslator(translator1)

	gui := g.MakeGui(settings)
	app.ConnectLastWindowClosed(func() {
		gui.SyncSettings()
	})
	app.Exec()
}
