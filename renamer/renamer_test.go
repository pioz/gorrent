package renamer

import (
	"testing"

	"github.com/therecipe/qt/core"
)

func TestRename(t *testing.T) {
	settings := core.NewQSettings("pioz", "gorrent", nil)
	Rename("/Users/pioz/Desktop/Fringe", settings.Value("tvdb/apikey", core.NewQVariant14("")).ToString(), settings.Value("tvdb/locale", core.NewQVariant14("en")).ToString())
}
