package gui

import (
	"C"
	"regexp"

	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
)

// SettingsDialog struct
type SettingsDialog struct {
	core.QObject
	*widgets.QDialog

	settings        *core.QSettings
	tvdbApikey      *widgets.QLineEdit
	tvdbLocale      *widgets.QComboBox
	locale          *widgets.QComboBox
	defaultRegexp   *widgets.QLineEdit
	dialogButtonBox *widgets.QDialogButtonBox

	_ func() `signal:"settingsSaved"`
}

var tvdbLocales = []string{"en", "cs", "da", "de", "el", "es", "fi", "fr", "he", "hr", "hu", "it", "ja", "ko", "nl", "no", "pl", "pt", "ru", "sl", "sv", "tr", "zh"}
var availableLocales = []string{"en", "it"}

// MakeSettingsDialog returns new SettingsDialog struct
func MakeSettingsDialog(window *widgets.QMainWindow, settings *core.QSettings) *SettingsDialog {
	sd := NewSettingsDialog(nil)
	sd.QDialog = widgets.NewQDialog(window, core.Qt__Dialog|core.Qt__MSWindowsFixedSizeDialogHint)
	sd.settings = settings
	sd.tvdbApikey = widgets.NewQLineEdit(nil)
	sd.tvdbApikey.SetEchoMode(widgets.QLineEdit__Password)
	sd.tvdbLocale = widgets.NewQComboBox(nil)
	sd.tvdbLocale.AddItems(tvdbLocales)
	sd.locale = widgets.NewQComboBox(nil)
	sd.locale.AddItems(availableLocales)
	sd.defaultRegexp = widgets.NewQLineEdit(nil)

	mainLayout := MakeQVFormLayout()
	mainLayout.AddBlock("Gorrent")
	mainLayout.AddRow(0, "Gorrent locale (need app restart to take effect):", sd.locale)
	mainLayout.AddRow(0, "Default regex to filter search results:", sd.defaultRegexp)
	mainLayout.AddBlock("TVDB")
	mainLayout.AddRow(1, "TVDB api key:", sd.tvdbApikey)
	mainLayout.AddRow(1, "Locale in which TVDB returns the data:", sd.tvdbLocale)

	sd.dialogButtonBox = widgets.NewQDialogButtonBox3(widgets.QDialogButtonBox__Save|widgets.QDialogButtonBox__Cancel, nil)
	mainLayout.AddWidget(sd.dialogButtonBox, 0, core.Qt__AlignRight|core.Qt__AlignBottom)
	sd.SetLayout(mainLayout.QVBoxLayout)

	sd.SetWindowTitle("Settings")
	sd.SetModal(true)
	sd.SetFixedSize(sd.SizeHint())

	sd.connectEvents()

	return sd
}

// Exec override
func (sd *SettingsDialog) Exec(settingKey string) {
	sd.setFormValuesFromSettings()
	switch settingKey {
	case "tvdb/apikey":
		sd.tvdbApikey.SetFocus2()
		sd.tvdbApikey.SetSelection(0, len(sd.tvdbApikey.Text()))
	default:
		sd.SetFocus2()
	}
	sd.QDialog.Exec()
}

func (sd *SettingsDialog) setFormValuesFromSettings() {
	sd.tvdbApikey.SetText(sd.settings.Value("tvdb/apikey", core.NewQVariant14("")).ToString())
	locale := sd.settings.Value("tvdb/locale", core.NewQVariant14("en")).ToString()
	for i := 0; i < sd.tvdbLocale.Count(); i++ {
		if sd.tvdbLocale.ItemText(i) == locale {
			sd.tvdbLocale.SetCurrentIndex(i)
			break
		}
	}
	locale = sd.settings.Value("gorrent/locale", core.NewQVariant14("en")).ToString()
	for i := 0; i < sd.locale.Count(); i++ {
		if sd.locale.ItemText(i) == locale {
			sd.locale.SetCurrentIndex(i)
			break
		}
	}
	sd.defaultRegexp.SetText(sd.settings.Value("gorrent/default_regexp", core.NewQVariant14("")).ToString())
}

func (sd *SettingsDialog) connectEvents() {
	sd.dialogButtonBox.ConnectAccepted(func() {
		sd.settings.SetValue("tvdb/apikey", core.NewQVariant14(sd.tvdbApikey.Text()))
		sd.settings.SetValue("tvdb/locale", core.NewQVariant14(sd.tvdbLocale.CurrentText()))
		sd.settings.SetValue("gorrent/locale", core.NewQVariant14(sd.locale.CurrentText()))
		if sd.defaultRegexp.Text() != "" {
			_, err := regexp.Compile("(?i)" + sd.defaultRegexp.Text())
			if err == nil {
				sd.settings.SetValue("gorrent/default_regexp", core.NewQVariant14(sd.defaultRegexp.Text()))
			}
		}
		sd.settings.Sync()
		sd.SettingsSaved()
		sd.Close()
	})
	sd.dialogButtonBox.ConnectRejected(func() {
		sd.Close()
	})
}
