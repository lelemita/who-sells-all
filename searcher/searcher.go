package searcher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"sort"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

const (
	PATH_SEARCH_ITEMS   = "/ttb/api/ItemSearch.aspx"
	PATH_ITEM_LOOK_UP   = "/ttb/api/ItemLookUp.aspx"
	PATH_USED_ITEM_MALL = "/shop/UsedShop/wuseditemall.aspx"
)

var regex_only_num = regexp.MustCompile("[^0-9]")

type BookMetaList struct {
	Books []BookMeta `json:"item"` // item: aladin API response parsing 을 위한 이름
}

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
	Books       map[string]UsedBook // isbn13: UsedBook
}

// 책의 고유 정보: 키워드로 ISBN을 찾기 위한 구조
type BookMeta struct {
	Itemid uint   `json:"itemId"`
	Isbn13 string `json:"isbn13"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Cover  string `json:"cover"`
}

// 중고 매물 정보
type UsedBook struct {
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

// Search book list by keyword
func (s *Searcher) Search(ctx context.Context, keyword string) (*BookMetaList, error) {
	// TODO pagination
	return s.searchWithPagination(ctx, keyword, 1, 20)
}

func (s *Searcher) searchWithPagination(ctx context.Context, keyword string, pageNum int, pageSize int) (*BookMetaList, error) {
	url := s.apiHost + PATH_SEARCH_ITEMS
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	qry := req.URL.Query()
	qry.Add("ttbkey", s.ttbkey)
	qry.Add("Query", keyword)
	qry.Add("output", "js")
	qry.Add("Start", intToStr(pageNum))
	qry.Add("MaxResults", intToStr(pageSize))
	qry.Add("version", "20131101")
	qry.Add("Cover", "Big")
	req.URL.RawQuery = qry.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to execute search query request", slog.String("keyword", keyword), slog.Any("error", err))
		return nil, err
	}
	defer resp.Body.Close()
	checkCode(ctx, resp)

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to read search response body", slog.String("keyword", keyword), slog.Any("error", err))
		return nil, err
	}
	respInfo := &BookMetaList{}
	if err := json.Unmarshal(respBytes, respInfo); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal search response", slog.String("keyword", keyword), slog.Any("error", err))
		return nil, err
	}
	if len(respInfo.Books) == 0 {
		return nil, errors.New("fail to find item")
	}
	return respInfo, nil
}

// Scrape Aladin used bookstore by isbn list
func (s *Searcher) GetOrderedList(ctx context.Context, isbns []string) ShopList {
	output := ShopList{}
	proposals := s.getProposals(ctx, isbns)
	for sName, seller := range proposals {
		shop := Shop{
			Name:        sName,
			Link:        s.apiHost + seller.Link,
			DeliveryFee: seller.DeliveryFee,
		}
		for _, book := range seller.Books {
			shop.TotalPrice += book.Price
		}
		output = append(output, shop)
	}

	sort.Sort(output)
	return output
}

func (s *Searcher) getProposals(ctx context.Context, isbns []string) Proposals {
	sellers := Proposals{}
	chIsbn := make(chan Proposals)
	for _, isbn := range isbns {
		go s.getForOneIsbn(ctx, isbn, chIsbn)
	}

	for i := 0; i < len(isbns); i++ {
		isbn := isbns[i]
		proposals := <-chIsbn
		for sName, s := range proposals {
			if seller, isExist := sellers[sName]; isExist {
				// TODO 아래 부분 테스트 필요
				seller.Books[isbn] = s.Books[isbn]
				sellers[sName] = seller
			} else {
				sellers[sName] = s
			}
		}
	}

	// SOMEDAY 현재는 다 가진 셀러만 보여줌, 일부도 보여주려면 기준 필요, 가중치?
	result := Proposals{}
	for sName, seller := range sellers {
		if len(seller.Books) >= len(isbns) {
			result[sName] = seller
		}
	}
	return result
}

func (s *Searcher) getForOneIsbn(ctx context.Context, isbn string, chIsbn chan<- Proposals) {
	totalResult := Proposals{}
	itemId, err := s.getItemIdByIsbn(ctx, isbn)
	if err != nil {
		chIsbn <- totalResult
		return
	}
	// TabType=0: 전체 목록 / SortOrder=9: 저가격순
	baseUrl := fmt.Sprintf("%s%s?TabType=0&SortOrder=9&ItemId=%s", s.apiHost, PATH_USED_ITEM_MALL, itemId)
	totalPage := getPages(ctx, baseUrl)
	chPage := make(chan Proposals)
	for page := 1; page <= totalPage; page++ {
		go s.extractFromPage(ctx, isbn, itemId, page, chPage)
	}
	for page := 1; page <= totalPage; page++ {
		pageResult := <-chPage
		for sName, seller := range pageResult {
			if s, isExist := totalResult[sName]; isExist {
				appendBooks(s.Books, seller.Books)
			} else {
				totalResult[sName] = seller
			}
		}
	}
	chIsbn <- totalResult
}

// 저렴한 가격 기준으로 중복 제거하며 합치기
func appendBooks(target map[string]UsedBook, appending map[string]UsedBook) {
	for isbn, new := range appending {
		if _, isExist := target[isbn]; !isExist {
			target[isbn] = new
		} else if target[isbn].Price > new.Price {
			target[isbn] = new
		}
	}
}

func (s *Searcher) getItemIdByIsbn(ctx context.Context, isbn string) (string, error) {
	if itemInfo, err := s.getAladinInfo(ctx, isbn); err != nil {
		return "", err
	} else {
		itemId := strconv.Itoa(int(itemInfo.ItemId))
		return itemId, nil
	}
}

func (s *Searcher) getAladinInfo(ctx context.Context, isbn string) (*ItemLookUpResult, error) {
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
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get Aladin info request", slog.String("isbn", isbn), slog.Any("error", err))
		return nil, err
	}
	defer resp.Body.Close()
	checkCode(ctx, resp)

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to read Aladin info response body", slog.String("isbn", isbn), slog.Any("error", err))
		return nil, err
	}
	respInfo := &ItemLookUpList{}
	if err := json.Unmarshal(respBytes, respInfo); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal Aladin info response", slog.String("isbn", isbn), slog.Any("error", err))
		return nil, err
	}
	if len(respInfo.Item) == 0 {
		return nil, errors.New("fail to find item")
	}
	return &respInfo.Item[0], nil
}

// TODO input parameter 줄이기
func (s *Searcher) extractFromPage(ctx context.Context, isbn string, itemId string, page int, chPage chan<- Proposals) {
	pageUrl := fmt.Sprintf("%s%s?TabType=0&SortOrder=9&ItemId=%s&page=%d", s.apiHost, PATH_USED_ITEM_MALL, itemId, page)
	pageResult := Proposals{}
	chTr := make(chan Proposals)

	slog.DebugContext(ctx, "Requesting page details",
		slog.String("itemId", itemId),
		slog.Int("page", page),
		slog.String("url", pageUrl),
	)

	resp, err := http.Get(pageUrl)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to request page URL", slog.String("url", pageUrl), slog.Any("error", err))
		chPage <- pageResult
		return
	}
	defer resp.Body.Close()
	checkCode(ctx, resp)

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to parse document from body", slog.String("url", pageUrl), slog.Any("error", err))
		chPage <- pageResult
		return
	}

	trTag := doc.Find(".Ere_usedsell_table > table > tbody > tr")
	trLength := trTag.Length()
	if trLength <= 1 {
		chPage <- pageResult
		return
	}

	trTag.Each(func(i int, tr *goquery.Selection) {
		if i == 0 {
			return
		}
		go extractFromTr(isbn, tr, chTr)
	})

	for i := 0; i < trLength-1; i++ {
		trResult := <-chTr
		for sName, seller := range trResult {
			if s, isExist := pageResult[sName]; isExist {
				appendBooks(s.Books, seller.Books)
				pageResult[sName] = s
			} else {
				pageResult[sName] = seller
			}
		}
	}
	chPage <- pageResult
}

func extractFromTr(isbn string, tr *goquery.Selection, ch chan<- Proposals) {
	var sName SellerName
	seller := Seller{}
	book := UsedBook{}
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

	seller.Books = map[string]UsedBook{isbn: book}
	ch <- Proposals{
		sName: seller,
	}
}

func getPages(ctx context.Context, url string) int {
	resp, err := http.Get(url)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get page count from URL", slog.String("url", url), slog.Any("error", err))
		return 1
	}
	defer resp.Body.Close()
	checkCode(ctx, resp)

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to parse page count document", slog.String("url", url), slog.Any("error", err))
		return 1
	}
	pageNum := doc.Find(".Ere_usedsell_num_box > div > div > ul > li").Length()
	if pageNum == 0 {
		pageNum = 1
	}
	return pageNum
}

func parsePrice(strPrice string) uint {
	price := regex_only_num.ReplaceAllString(strPrice, "")
	result, err := strconv.ParseUint(price, 0, 64)
	if err != nil {
		slog.Error("Failed to parse price", slog.String("value", strPrice), slog.Any("error", err))
	}
	return uint(result)
}

func checkCode(ctx context.Context, resp *http.Response) {
	if resp == nil {
		slog.ErrorContext(ctx, "Response is nil")
		return
	}
	if resp.StatusCode != http.StatusOK {
		urlStr := ""
		if resp.Request != nil && resp.Request.URL != nil {
			urlStr = resp.Request.URL.String()
		}
		slog.WarnContext(ctx, "Request failed",
			slog.Int("status", resp.StatusCode),
			slog.String("url", urlStr),
		)
	}
}
