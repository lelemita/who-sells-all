package searcher

import (
	"log"
	"os"
	"testing"
)

var ttbkey string
var genie Searcher

func TestMain(m *testing.M) {

	ttbkey = os.Getenv("ttbkey")
	if len(ttbkey) == 0 {
		log.Fatal("ttbkey value is required")
	}
	genie = NewSearcher("https://www.aladin.co.kr", ttbkey)
	m.Run()
}

func TestSearch(t *testing.T) {
	results, err := genie.Search("김초엽")
	if err != nil {
		t.Error(err)
	}

	if len(results.Books) == 0 {
		t.Error("fail to find item")
	}

	for _, book := range results.Books {
		t.Log(book.Title, book.Isbn13)
	}
}
