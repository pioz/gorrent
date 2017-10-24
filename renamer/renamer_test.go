package renamer

import (
	"testing"

	"github.com/therecipe/qt/core"
)

func TestRename(t *testing.T) {
	settings := core.NewQSettings("pioz", "gorrent", nil)
	r := MakeRenamer(settings)
	r.Rename("/Users/pioz/Desktop/Fringe")
}
