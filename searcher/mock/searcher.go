package mock

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

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

func (s *Searcher) FirstItemLookUp(isbn string) (*searcher.ItemLookUpResult, error) {
	url := s.apiHost + searcher.PATH_ITEM_LOOK_UP
	req, err := http.NewRequest(http.MethodGet, url, nil)
	checkErr(err)

	client := &http.Client{}
	resp, err := client.Do(req)
	checkErr(err)
	checkCode(resp)
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	respInfo := &searcher.ItemLookUpList{}
	if err := json.Unmarshal(respBytes, respInfo); err != nil {
		return nil, err
	}

	return &respInfo.Item[0], nil
}

func (s *Searcher) GetIdByIsbn(isbn string) string {
	itemInfo, err := s.FirstItemLookUp(isbn)
	checkErr(err)
	itemId := strconv.Itoa(int(itemInfo.ItemId))
	return itemId
}

func (s *Searcher) CrawlProposals(itemInfo searcher.ItemLookUpResult) *searcher.Bidding {
	return nil
}

func checkCode(resp *http.Response) {
	if resp.StatusCode != http.StatusOK {
		// TODO 에러 타입 정의 필요
		log.Fatalf("Request failed with Status: %d", resp.StatusCode)
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
