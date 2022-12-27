package mock

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/lelemita/who_sells_all/searcher"
)

type Searcher struct {
	apiHost string
}

func NewSearcher() Searcher {
	return Searcher{
		apiHost: "http://localhost:8081",
	}
}

func (s *Searcher) GetByIsbn(isbn string) (*searcher.ItemLookUpResult, error) {
	url := s.apiHost + searcher.PATH_ITEM_LOOK_UP
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	respInfo := &searcher.ItemLookUpResult{}
	if err := json.Unmarshal(respBytes, respInfo); err != nil {
		return nil, err
	}

	return respInfo, nil
}

func (s *Searcher) CrawlProposals(itemInfo searcher.ItemLookUpResult) (*searcher.OneResult, error) {
	return nil, nil
}
