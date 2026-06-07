package searcher

import (
	"context"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/lelemita/who_sells_all/searcher/mock"
	"github.com/stretchr/testify/assert"
)

var ttbkey string

func TestMain(m *testing.M) {

	ttbkey = os.Getenv("ttbkey")
	if len(ttbkey) == 0 {
		log.Fatal("ttbkey value is required")
	}

	var wg sync.WaitGroup
	wg.Add(1)
	mock.RunAladinWeb(&wg)
	wg.Wait()
	exitVal := m.Run()

	os.Exit(exitVal)
}

func TestSearch(t *testing.T) {
	target := NewSearcher("https://www.aladin.co.kr", ttbkey)
	results, err := target.Search(context.Background(), "김초엽")
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

func TestExtractFromPage(t *testing.T) {
	target := NewSearcher("http://localhost:8082", "ttbkey")
	chPage := make(chan Proposals)
	ctx := context.WithValue(context.Background(), "rid", "test-extractFromPage")
	ctx, cancelFn := context.WithTimeout(ctx, time.Second)
	defer cancelFn()
	go target.extractFromPage(ctx, "ISBN", "itemID", 1, chPage)

	pageResult := <-chPage

	tests := []struct {
		seller string
		price  uint
		status string
	}{
		{"중고매장전주점", 7300, "상"},
		{"긴급세일책방", 7200, "상"},
		{"참좋은이야기", 1000, "중"},
		{"지식의선율", 3000, "상"},
	}

	for _, tc := range tests {
		assert := assert.New(t)
		t.Run(tc.seller, func(t *testing.T) {
			books := pageResult[SellerName(tc.seller)].Books
			assert.Len(books, 1)
			assert.Equal(tc.price, books["ISBN"].Price)
			assert.Equal(tc.status, books["ISBN"].Status)
		})
	}
}
