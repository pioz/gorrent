package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/pioz/gorrent/common"
	"github.com/therecipe/qt/core"
)

// Scraper type
type Scraper struct {
	core.QObject

	_ func(string)   `signal:"errorOccured"`
	_ func([][]byte) `signal:"searchCompleted"`
	_ func(string)   `signal:"downloadTorrentStarted"`
	_ func(int, int) `signal:"downloadTorrentCompleted"`
	_ func()         `signal:"downloadTorrentsCompleted"`
}

const endPointURL string = "http://www.tntvillage.scambioetico.org/src/releaselist.php"

// RetrieveTorrents func
func (s *Scraper) RetrieveTorrents(q string) {
	var (
		pages    int
		torrents [][]byte
	)
	pages = 1
	for page := 1; page <= pages && page <= 10; page++ {
		resp, err := http.PostForm(endPointURL, url.Values{"srcrel": {url.QueryEscape(q)}, "page": {strconv.Itoa(page)}})
		if err != nil {
			// return nil, err
			s.ErrorOccured(err.Error())
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			errorMessage := fmt.Sprintf("Get a response with status code %d", resp.StatusCode)
			// return nil, errors.New(errorMessage)
			s.ErrorOccured(errorMessage)
			return
		}

		doc, err := goquery.NewDocumentFromResponse(resp)
		if err != nil {
			// return nil, err
			s.ErrorOccured(err.Error())
			return
		}
		if page == 1 {
			p, _ := doc.Find(".total").First().Attr("a")
			pages, _ = strconv.Atoi(p)
		}
		doc.Find(".showrelease_tb table tr").Each(func(i int, s *goquery.Selection) {
			if i > 0 {
				// maybe check _ for some errors... :D
				cells := s.Find("td")
				link, _ := cells.Slice(0, 1).Find("a").First().Attr("href")
				magnet, _ := cells.Slice(1, 2).Find("a").First().Attr("href")
				seeds, _ := strconv.Atoi(cells.Slice(4, 5).Text())
				tmp := cells.Slice(6, 7)
				name := tmp.Find("a").First().Text()
				html, _ := tmp.Html()
				info := strings.Split(html, "</a> ")[1]
				json, _ := json.Marshal(common.Torrent{Link: link, Magnet: magnet, Name: name, Info: info, Seeds: seeds})
				torrents = append(torrents, json)
			}
		})
	}
	// return torrents, nil
	s.SearchCompleted(torrents)
}

// DownloadTorrents download the torrent files and save them in a
// file named name.torrent inside dirName
func (s *Scraper) DownloadTorrents(torrents map[int][]byte, dirName string) {
	var t common.Torrent
	i := 0
	for k := range torrents {
		i++
		json.Unmarshal(torrents[k], &t)
		s.DownloadTorrentStarted(t.Name)
		file, err := os.Create(dirName + "/" + t.Name + ".torrent")
		if err != nil {
			s.ErrorOccured(err.Error())
			return
		}
		defer file.Close()
		resp, err := http.Get(t.Link)
		if err != nil {
			s.ErrorOccured(err.Error())
			return
		}
		defer resp.Body.Close()
		_, err = io.Copy(file, resp.Body)
		if err != nil {
			s.ErrorOccured(err.Error())
			return
		}
		s.DownloadTorrentCompleted(k, 100*i/len(torrents))
	}
	time.Sleep(time.Millisecond * 500)
	s.DownloadTorrentsCompleted()
}
