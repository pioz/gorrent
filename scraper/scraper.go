package scraper

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Torrent type
type Torrent struct {
	Link   string
	Magnet string
	Seeds  int
	Name   string
	Info   string
}

const endPointURL string = "http://www.tntvillage.scambioetico.org/src/releaselist.php"

// RetrieveTorrents func
func RetrieveTorrents(q string) ([]Torrent, error) {
	var (
		pages    int
		torrents []Torrent
	)
	pages = 1
	for page := 1; page <= pages && page <= 10; page++ {
		resp, err := http.PostForm(endPointURL, url.Values{"srcrel": {url.QueryEscape(q)}, "page": {strconv.Itoa(page)}})
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			errorMessage := fmt.Sprintf("Get a response with status code %d", resp.StatusCode)
			return nil, errors.New(errorMessage)
		}

		doc, err := goquery.NewDocumentFromResponse(resp)
		if err != nil {
			return nil, err
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
				torrents = append(torrents, Torrent{link, magnet, seeds, name, info})
			}
		})
	}

	return torrents, nil
}
