package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/lelemita/who_sells_all/searcher"
)

var regex_only_num = regexp.MustCompile("[^0-9]")

type Searcher struct {
	apiHost string
	ttbkey  string
}

func NewSearcher() Searcher {
	ttbkey := os.Getenv("ttbkey")
	return Searcher{
		apiHost: "https://www.aladin.co.kr",
		ttbkey:  ttbkey,
	}
}

func (s *Searcher) firstItemLookUp(isbn string) (*searcher.ItemLookUpResult, error) {
	url := s.apiHost + searcher.PATH_ITEM_LOOK_UP
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	qry := req.URL.Query()
	qry.Add("ttbkey", s.ttbkey)
	qry.Add("itemId", isbn)
	// qry.Add("itemIdType", "ISBN")
	qry.Add("output", "js")
	qry.Add("OptResult", "usedList")
	qry.Add("version", "20131101")
	req.URL.RawQuery = qry.Encode()

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
	if len(respInfo.Item) == 0 {
		return nil, errors.New("fail to find item")
	}
	return &respInfo.Item[0], nil
}

func (s *Searcher) getIdByIsbn(isbn string) string {
	itemInfo, err := s.firstItemLookUp(isbn)
	checkErr(err)
	itemId := strconv.Itoa(int(itemInfo.ItemId))
	return itemId
}

func (s *Searcher) crawlProposals(itemId string) searcher.Bidding {
	// TabType=0: 전체 목록 / SortOrder=9: 저가격순
	baseUrl := fmt.Sprintf("%s%s?TabType=0&SortOrder=9&ItemId=%s", s.apiHost, searcher.PATH_USED_ITEM_MALL, itemId)
	bidding := searcher.Bidding{}
	totalPage := getPages(baseUrl)
	for page := 1; page <= totalPage; page++ {
		pageUrl := baseUrl + "&page=" + strconv.Itoa(page)
		log.Println("Requesting... ", pageUrl)

		resp, err := http.Get(pageUrl)
		checkErr(err)
		checkCode(resp)
		defer resp.Body.Close()

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		checkErr(err)
		doc.Find(".Ere_usedsell_table > table > tbody > tr").Each(func(i int, s *goquery.Selection) {
			if i == 0 {
				return
			}
			var sName searcher.SellerName
			seller := searcher.Seller{}
			book := searcher.Book{ItemId: itemId}
			s.Find("td").Each(func(j int, ss *goquery.Selection) {
				tdTag := ss.Find("div")
				if tdTag.HasClass("seller") {
					aTag := tdTag.Find("ul > li > a")
					sLink, _ := aTag.Attr("href")
					seller.Link = sLink
					sName = searcher.SellerName(aTag.Text())
				} else if tdTag.HasClass("price") {
					strPrice := tdTag.Find("ul > li:nth-child(1)").Text()
					book.Price = parsePrice(strPrice)
					seller.DeliveryFee = tdTag.Find("ul > li:nth-child(3)").Text()
				} else if tdTag.HasClass("info") {
					link, _ := tdTag.Find("ul > li:first-child > a").Attr("href")
					book.Link = link
				} else {
					ss.Find("span > span").Each(func(k int, sss *goquery.Selection) {
						book.Status = sss.Text()
					})
				}
			})
			if s, isExist := bidding[sName]; isExist {
				s.Proposal = append(s.Proposal, book)
			} else {
				seller.Proposal = []searcher.Book{book}
				bidding[sName] = seller
			}
		})
	}
	return bidding
}

func (s *Searcher) GetProposals(isbns []string) searcher.Bidding {
	sellers := searcher.Bidding{}
	for _, isbn := range isbns {
		itemId := s.getIdByIsbn(isbn)
		proposals := s.crawlProposals(itemId)
		for sName, s := range proposals {
			if seller, isExist := sellers[sName]; isExist {
				seller.Proposal = append(seller.Proposal, s.Proposal...)
				sellers[sName] = seller
			} else {
				sellers[sName] = s
			}
		}
	}

	result := searcher.Bidding{}
	for sName, seller := range sellers {
		if len(seller.Proposal) >= len(isbns) {
			result[sName] = seller
		}
	}
	return result
}

func getPages(url string) int {
	resp, err := http.Get(url)
	checkErr(err)
	checkCode(resp)
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	checkErr(err)
	pageNum := doc.Find(".Ere_usedsell_num_box > div > div > ul > li").Length()
	return pageNum
}

func parsePrice(strPrice string) uint {
	price := regex_only_num.ReplaceAllString(strPrice, "")
	result, err := strconv.ParseUint(price, 0, 64)
	checkErr(err)
	return uint(result)
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
