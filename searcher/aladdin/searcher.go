package service

import (
	"encoding/json"
	"io"
	"net/http"
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

func (s *Searcher) GetByIsbn(isbn string) (*searcher.ItemLookUpResult, error) {
	url := s.apiHost + searcher.PATH_ITEM_LOOK_UP
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	qry := req.URL.Query()
	qry.Add("ttbkey", s.ttbkey)
	qry.Add("itemId", isbn)
	qry.Add("output", "js")
	qry.Add("OptResult", "usedList")
	qry.Add("itemIdType", "ISBN13")
	qry.Add("version", "20131101")
	req.URL.RawQuery = qry.Encode()

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

	// TODO
	// https://www.aladin.co.kr/shop/UsedShop/wuseditemall.aspx?ItemId=284863481&TabType=0&SortOrder=9&page=1
	// 여기서 page 결과 없을 때까지 page 올리면서 crawling 하면 됨.
	// TabTyle, SortOrder 상수로 빼뒀음
	// parameter: itemId만 필요한데? GetByIsbn 에서도 그것만 얻도록 고쳐야 하나.... 아깝긴 하지만.

	return nil, nil
}
