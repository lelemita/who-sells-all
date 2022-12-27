package mock

import (
	"github.com/lelemita/who_sells_all/mock"
	"github.com/lelemita/who_sells_all/searcher"
)

type Searcher struct {
	apiHost string
}

func NewSearcher() Searcher {
	go mock.RunAladdinApiMock()
	return Searcher{
		apiHost: "http://localhost:8081",
	}
}

func (s *Searcher) GetByIsbn(isbn string) (searcher.OneResult, error) {
	return searcher.OneResult{}, nil
}
