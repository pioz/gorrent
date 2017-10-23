package i18n

import (
	"github.com/nicksnyder/go-i18n/i18n"
	"github.com/therecipe/qt/core"
)

// T is the translate function
var T i18n.TranslateFunc

// LoadI18nFile load translation file
func LoadI18nFile(locale string) {
	file := core.NewQFile2(":/" + locale + ".json")
	if file.Open(core.QIODevice__ReadOnly | core.QIODevice__Text) {
		in := core.NewQTextStream2(file)
		i18n.ParseTranslationFileBytes(locale+".json", []byte(in.ReadAll()))
		file.Close()
	}
}
