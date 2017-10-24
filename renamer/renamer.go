package renamer

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/pioz/tvdb"
	"github.com/therecipe/qt/core"
)

var regexp1 = regexp.MustCompile(`(?i)(\d{1,2})x(\d{1,2})`)
var regexp2 = regexp.MustCompile(`(?i)s(\d{1,2})\w?_?e(\d{1,2})`)
var regexp3 = regexp.MustCompile(`(?i)(\d{1,2})\/(\d{1,2})\s`)

// Renamer struct
type Renamer struct {
	core.QObject
	settings *core.QSettings
	client   tvdb.Client

	_ func(string) `signal:"errorOccured"`
	_ func()       `signal:"renameSeriesCompleted"`
}

// MakeRenamer returns a new Renamer struct
func MakeRenamer(settings *core.QSettings) *Renamer {
	r := NewRenamer(nil)
	r.settings = settings
	r.client = tvdb.Client{}
	r.UpdateSettings()
	return r
}

// UpdateSettings slot
func (r *Renamer) UpdateSettings() {
	r.client.Apikey = r.settings.Value("tvdb/apikey", core.NewQVariant14("")).ToString()
	r.client.Language = r.settings.Value("tvdb/locale", core.NewQVariant14("en")).ToString()
	if r.client.Language == "" {
		r.client.Language = "en"
	}
}

// Rename function rename all series files in the form
// "seasonNumberxepisodeNumber - episodeTitle". The titles are retrieved by TVDB
// api.
func (r *Renamer) Rename(dirPath string) {
	err := r.client.Login()
	if err != nil {
		r.ErrorOccured(err.Error())
		return
	}
	seriesName := filepath.Base(dirPath)
	series, err := r.client.BestSearch(seriesName)
	if err != nil {
		if tvdb.HaveCodeError(404, err) {
			r.ErrorOccured("\"" + seriesName + "\" series not found")
			return
		}
		r.ErrorOccured(err.Error())
		return
	}
	err = r.client.GetSeriesEpisodes(&series, nil)
	if err != nil {
		r.ErrorOccured(err.Error())
		return
	}
	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			season, episode := titleToCode(filepath.Base(path))
			if season != 0 && episode != 0 {
				os.Rename(path, fmt.Sprintf("%s/%dx%02d - %s.avi", filepath.Dir(path), season, episode, series.GetEpisode(season, episode).EpisodeName))
			}
		}
		return nil
	})
	r.RenameSeriesCompleted()
}

func titleToCode(title string) (int, int) {
	m := regexp1.FindStringSubmatch(title)
	if m != nil {
		x, _ := strconv.Atoi(m[1])
		y, _ := strconv.Atoi(m[2])
		return x, y
	}
	m = regexp2.FindStringSubmatch(title)
	if m != nil {
		x, _ := strconv.Atoi(m[1])
		y, _ := strconv.Atoi(m[2])
		return x, y
	}
	m = regexp3.FindStringSubmatch(title)
	if m != nil {
		x, _ := strconv.Atoi(m[1])
		y, _ := strconv.Atoi(m[2])
		return x, y
	}
	return 0, 0
}
