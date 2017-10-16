package renamer

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/pioz/tvdb"
)

var regexp1 = regexp.MustCompile(`(?i)(\d{1,2})x(\d{1,2})`)
var regexp2 = regexp.MustCompile(`(?i)s(\d{1,2})\w?_?e(\d{1,2})`)
var regexp3 = regexp.MustCompile(`(?i)(\d{1,2})\/(\d{1,2})\s`)

// Rename function rename all series files in the form
// "seasonNumberxepisodeNumber - episodeTitle". The titles are retrieved by TVDB
// api.
func Rename(dirPath string) error {
	c := tvdb.Client{Apikey: "YOUR API KEY"}
	err := c.Login()
	if err != nil {
		fmt.Println("LOGIN ERROR")
		return err
	}
	seriesName := filepath.Base(dirPath)
	series, err := c.BestSearch(seriesName)
	if err != nil {
		if tvdb.HaveCodeError(404, err) {
			return errors.New("\"" + seriesName + "\" series not found")
		}
		return err
	}
	err = c.GetSeriesEpisodes(&series, nil)
	if err != nil {
		return err
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
	return nil
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
