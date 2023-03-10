package searcher

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
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

type Proposals map[SellerName]Seller

type Seller struct {
	Link        string
	DeliveryFee string
	Books       []Book
}

type Book struct {
	ItemId string
	Price  uint
	Status string
	Link   string
}

type Shop struct {
	Name        SellerName `json:"name"`
	Link        string     `json:"link"`
	TotalPrice  uint       `json:"totalPrice"`
	DeliveryFee string     `json:"deliveryFee"`
}

type ShopList []Shop

func (s ShopList) Len() int {
	return len(s)
}

func (s ShopList) Less(i, j int) bool {
	return s[i].TotalPrice < s[j].TotalPrice
}

func (s ShopList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type Searcher struct {
	apiHost string
	ttbkey  string
}

func NewSearcher(host string, ttbkey string) Searcher {
	return Searcher{
		apiHost: host,
		ttbkey:  ttbkey,
	}
}

// Scrape Aladin used bookstore by isbn list
func (s *Searcher) GetOrderedList(isbns []string) ShopList {
	output := ShopList{}
	proposals := s.getProposals(isbns)
	for sName, seller := range proposals {
		shop := Shop{
			Name:        sName,
			Link:        s.apiHost + seller.Link,
			DeliveryFee: seller.DeliveryFee,
		}
		for _, book := range seller.Books {
			shop.TotalPrice += book.Price
		}
		// Think 배송비 반영 해야 하려나? 얼마이상 무료배송 이런거 까지 고려해야 할지도..
		output = append(output, shop)
	}

	sort.Sort(output)
	return output
}

func (s *Searcher) getProposals(isbns []string) Proposals {
	sellers := Proposals{}
	chIsbn := make(chan Proposals)
	for _, isbn := range isbns {
		go s.getForOneIsbn(isbn, chIsbn)
	}

	for i := 0; i < len(isbns); i++ {
		proposals := <-chIsbn
		for sName, s := range proposals {
			if seller, isExist := sellers[sName]; isExist {
				seller.Books = append(seller.Books, s.Books...)
				sellers[sName] = seller
			} else {
				sellers[sName] = s
			}
		}
	}

	// TODO 현재는 다 가진 셀러만 보여줌, 일부도 보여주려면 기준 필요, 가중치?
	result := Proposals{}
	for sName, seller := range sellers {
		if len(seller.Books) >= len(isbns) {
			result[sName] = seller
		}
	}
	return result
}

func (s *Searcher) getForOneIsbn(isbn string, chIsbn chan<- Proposals) {
	totalResult := Proposals{}
	itemId, err := s.getItemIdByIsbn(isbn)
	if err != nil {
		chIsbn <- totalResult
		return
	}
	// TabType=0: 전체 목록 / SortOrder=9: 저가격순
	baseUrl := fmt.Sprintf("%s%s?TabType=0&SortOrder=9&ItemId=%s", s.apiHost, PATH_USED_ITEM_MALL, itemId)
	totalPage := getPages(baseUrl)
	chPage := make(chan Proposals)
	for page := 1; page <= totalPage; page++ {
		go s.extractFromPage(itemId, page, chPage)
	}
	for page := 1; page <= totalPage; page++ {
		pageResult := <-chPage
		for sName, seller := range pageResult {
			if s, isExist := totalResult[sName]; isExist {
				s.Books = append(s.Books, seller.Books...)
			} else {
				totalResult[sName] = seller
			}
		}
	}
	chIsbn <- totalResult
}

func (s *Searcher) getItemIdByIsbn(isbn string) (string, error) {
	if itemInfo, err := s.getAladinInfo(isbn); err != nil {
		return "", err
	} else {
		itemId := strconv.Itoa(int(itemInfo.ItemId))
		return itemId, nil
	}
}

func (s *Searcher) getAladinInfo(isbn string) (*ItemLookUpResult, error) {
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
	// 참고: qry.Add("itemIdType", "ISBN") // (default) 이걸로 해도 ISBN13 처리 가능, 반대는 안됨
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

func (s *Searcher) extractFromPage(itemId string, page int, chPage chan<- Proposals) {
	pageUrl := fmt.Sprintf("%s%s?TabType=0&SortOrder=9&ItemId=%s&page=%d", s.apiHost, PATH_USED_ITEM_MALL, itemId, page)
	pageResult := Proposals{}
	chTr := make(chan Proposals)

	// TODO 로거 만들어서 develop 모드에서는 출력하자
	// log.Println("Requesting... ", pageUrl)
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
		trResult := <-chTr
		for sName, seller := range trResult {
			if s, isExist := pageResult[sName]; isExist {
				s.Books = append(s.Books, seller.Books...)
			} else {
				pageResult[sName] = seller
			}
		}
	}
	chPage <- pageResult
}

func extractFromTr(itemId string, tr *goquery.Selection, ch chan<- Proposals) {
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

	seller.Books = []Book{book}
	ch <- Proposals{
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
	if pageNum == 0 {
		pageNum = 1
	}
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
		fmt.Fprintf(os.Stderr, "Request failed with Status: %d", resp.StatusCode)
	}
}

// TODO goroutine 내에서의 에러와 recover 어떻게 처리할지 생각해보기
func checkErr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
