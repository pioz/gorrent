package scraper

import (
	"testing"
)

func TestGetBody(t *testing.T) {
	torrents, err := RetrieveTorrents("Trono di spade s01")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(torrents)
	if len(torrents) != 25 {
		t.Fatalf("Get a invalid number of torrents: expected 25, have %d", len(torrents))
	}
}
