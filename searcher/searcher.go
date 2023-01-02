package searcher

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
)

const (
	PATH_ITEM_LOOK_UP   = "/ttb/api/ItemLookUp.aspx"
	PATH_USED_ITEM_MALL = "/shop/UsedShop/wuseditemall.aspx"
)

var regex_only_num = regexp.MustCompile("[^0-9]")

type ItemLookUpList struct {
	Item []ItemLookUpResult `json:"item"`
}

type ItemLookUpResult struct {
	ItemId  uint `json:"itemId"`
	SubInfo struct {
		UsedList struct {
			AladinUsed UsedInfo `json:"aladinUsed"`
			UserUsed   UsedInfo `json:"userUsed"`
			SpaceUsed  UsedInfo `json:"spaceUsed"`
		} `json:"usedList"`
	} `json:"subInfo"`
}

type UsedInfo struct {
	ItemCount int    `json:"itemCount"`
	MinPrice  int    `json:"minPrice"`
	Link      string `json:"link"`
}

type SellerName string
type Bidding map[SellerName]Seller

type Seller struct {
	Link        string
	DeliveryFee string
	Proposal    []Book
}

type Book struct {
	ItemId string
	Price  uint
	Status string
	Link   string
}

type Searcher struct {
	apiHost string
	ttbkey  string
}

func NewSearcher(host string) Searcher {
	ttbkey := os.Getenv("ttbkey")
	return Searcher{
		apiHost: host,
		ttbkey:  ttbkey,
	}
}

func (s *Searcher) GetProposals(isbns []string) Bidding {
	sellers := Bidding{}
	chIsbn := make(chan Bidding)
	for _, isbn := range isbns {
		go s.crawlProposals(isbn, chIsbn)
	}

	for i := 0; i < len(isbns); i++ {
		proposals := <-chIsbn
		for sName, s := range proposals {
			if seller, isExist := sellers[sName]; isExist {
				seller.Proposal = append(seller.Proposal, s.Proposal...)
				sellers[sName] = seller
			} else {
				sellers[sName] = s
			}
		}
	}

	// TODO 가격순 정렬 필요
	// TODO 현재는 다 가진 셀러만 보여줌, 일부도 보여주려면 기준 필요, 가중치?
	result := Bidding{}
	for sName, seller := range sellers {
		if len(seller.Proposal) >= len(isbns) {
			result[sName] = seller
		}
	}
	return result
}

func (s *Searcher) getIdByIsbn(isbn string) string {
	itemInfo, err := s.firstItemLookUp(isbn)
	checkErr(err)
	itemId := strconv.Itoa(int(itemInfo.ItemId))
	return itemId
}

func (s *Searcher) firstItemLookUp(isbn string) (*ItemLookUpResult, error) {
	url := s.apiHost + PATH_ITEM_LOOK_UP
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	qry := req.URL.Query()
	qry.Add("ttbkey", s.ttbkey)
	qry.Add("itemId", isbn)
	qry.Add("output", "js")
	qry.Add("OptResult", "usedList")
	qry.Add("version", "20131101")
	// qry.Add("itemIdType", "ISBN") // (default) 이걸로 해도 ISBN13 처리 가능, 반대는 안됨
	req.URL.RawQuery = qry.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	checkErr(err)
	checkCode(resp)
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	respInfo := &ItemLookUpList{}
	if err := json.Unmarshal(respBytes, respInfo); err != nil {
		return nil, err
	}
	if len(respInfo.Item) == 0 {
		return nil, errors.New("fail to find item")
	}
	return &respInfo.Item[0], nil
}

func (s *Searcher) crawlProposals(isbn string, chIsbn chan<- Bidding) {
	itemId := s.getIdByIsbn(isbn)
	// TabType=0: 전체 목록 / SortOrder=9: 저가격순
	baseUrl := fmt.Sprintf("%s%s?TabType=0&SortOrder=9&ItemId=%s", s.apiHost, PATH_USED_ITEM_MALL, itemId)
	totalBidding := Bidding{}
	totalPage := getPages(baseUrl)
	chPage := make(chan Bidding)
	for page := 1; page <= totalPage; page++ {
		go s.extractFromPage(itemId, page, chPage)
	}
	for page := 1; page <= totalPage; page++ {
		bidding := <-chPage
		for sName, seller := range bidding {
			if s, isExist := totalBidding[sName]; isExist {
				s.Proposal = append(s.Proposal, seller.Proposal...)
			} else {
				totalBidding[sName] = seller
			}
		}
	}
	chIsbn <- totalBidding
}

func (s *Searcher) extractFromPage(itemId string, page int, chPage chan<- Bidding) {
	pageUrl := fmt.Sprintf("%s%s?TabType=0&SortOrder=9&ItemId=%s&page=%d", s.apiHost, PATH_USED_ITEM_MALL, itemId, page)
	pageBidding := Bidding{}
	chTr := make(chan Bidding)

	log.Println("Requesting... ", pageUrl)
	resp, err := http.Get(pageUrl)
	checkErr(err)
	checkCode(resp)
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	checkErr(err)
	trTag := doc.Find(".Ere_usedsell_table > table > tbody > tr")
	trTag.Each(func(i int, tr *goquery.Selection) {
		if i == 0 {
			return
		}
		go extractFromTr(itemId, tr, chTr)
	})

	for i := 0; i < trTag.Length()-1; i++ {
		bidding := <-chTr
		for sName, seller := range bidding {
			if s, isExist := pageBidding[sName]; isExist {
				s.Proposal = append(s.Proposal, seller.Proposal...)
			} else {
				pageBidding[sName] = seller
			}
		}
	}
	chPage <- pageBidding
}

func extractFromTr(itemId string, tr *goquery.Selection, ch chan<- Bidding) {
	var sName SellerName
	seller := Seller{}
	book := Book{ItemId: itemId}
	tr.Find("td").Each(func(j int, td *goquery.Selection) {
		tdTag := td.Find("div")
		if tdTag.HasClass("seller") {
			aTag := tdTag.Find("ul > li > a")
			sLink, _ := aTag.Attr("href")
			seller.Link = sLink
			sName = SellerName(aTag.Text())
		} else if tdTag.HasClass("price") {
			strPrice := tdTag.Find("ul > li:nth-child(1)").Text()
			book.Price = parsePrice(strPrice)
			seller.DeliveryFee = tdTag.Find("ul > li:nth-child(3)").Text()
		} else if tdTag.HasClass("info") {
			link, _ := tdTag.Find("ul > li:first-child > a").Attr("href")
			book.Link = link
		} else {
			td.Find("span > span").Each(func(k int, sss *goquery.Selection) {
				book.Status = sss.Text()
			})
		}
	})

	seller.Proposal = []Book{book}
	ch <- Bidding{
		sName: seller,
	}
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
