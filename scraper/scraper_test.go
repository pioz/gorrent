package scraper

import (
	"testing"
	"time"
)

func TestGetBody(t *testing.T) {
	s := NewScraper(nil)
	s.ConnectSearchCompleted(func(torrents [][]byte) {
		if len(torrents) != 25 {
			t.Fatalf("Get a invalid number of torrents: expected 25, have %d", len(torrents))
		}
	})
	s.ConnectErrorOccured(func(err string) {
		t.Fatal(err)
	})
	s.RetrieveTorrents("Trono di spade s01")
	time.Sleep(time.Second)
}
