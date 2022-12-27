package service

import (
	"os"

	"github.com/lelemita/who_sells_all/searcher"
)

type Searcher struct {
	apiHost string
	ttbkey  string
}

func NewSearcher() Searcher {
	ttbkey := os.Getenv("ttbkey")
	return Searcher{
		apiHost: "http://www.aladin.co.kr",
		ttbkey:  ttbkey,
	}
}

func (s *Searcher) GetByIsbn(isbn string) (searcher.OneResult, error) {
	return searcher.OneResult{}, nil
}
