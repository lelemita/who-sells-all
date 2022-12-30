package mock

import (
	"fmt"
	"net/http"
	"sync"
)

const RESP_ITEMLOOKUP = `{
	"version": "20131101",
	"logo": "http://image.aladin.co.kr/img/header/2011/aladin_logo_new.gif",
	"title": "알라딘 상품정보 - The Earthian Tales 어션 테일즈 No.1",
	"link": "https://www.aladin.co.kr/shop/wproduct.aspx?ItemId=284863481&amp;partner=openAPI",
	"pubDate": "Thu, 22 Dec 2022 12:32:10 GMT",
	"totalResults": 1,
	"startIndex": 1,
	"itemsPerPage": 1,
	"query": "isbn13=9772799628000",
	"searchCategoryId": 0,
	"searchCategoryName": "",
	"item": [
		{
			"title": "The Earthian Tales 어션 테일즈 No.1 - alone",
			"link": "http://www.aladin.co.kr/shop/wproduct.aspx?ItemId=284863481&amp;partner=openAPI&amp;start=api",
			"author": "김보영, 고호관, 곽재식, 구한나리, 구환회, 김창규, 김효선, 박경만, 박문영, 서바이벌SF키트, 시아란, 심너울, 심완선, 위래, 이경희, 이서영, 이수현, 이지용, 이하진, 전삼혜, 전혜진, 정명섭, 정보라, 정이담, 진규, 최의택, 한승태, 해도연, 홍지운, 황인찬, 루토, OOO, 이주혜 (지은이)",
			"pubDate": "2022-01-01",
			"description": "지구에서, 지구인들이, 계절마다 만들어내는 경이로운 이야기, SF 전문 계간 문학잡지. 각자의 시간을, 공간을, 세상을 성실히 다져온 이야기꾼들은 그 빛을 반사해 저 먼 지면에 스민 밤을 밝혀줄 준비가 되어 있다. 일 년에 네 번, 계절이 올 때마다 찾아올 어션 테일즈의 시작을 알린다.",
			"isbn": "K912835128",
			"isbn13": "9772799628000",
			"itemId": 284863481,
			"priceSales": 22500,
			"priceStandard": 25000,
			"mallType": "BOOK",
			"stockStatus": "",
			"mileage": 1130,
			"cover": "https://image.aladin.co.kr/product/28486/34/coversum/k912835128_1.jpg",
			"categoryId": 51234,
			"categoryName": "국내도서>소설/시/희곡>문학 잡지>기타",
			"publisher": "아작",
			"salesPoint": 10942,
			"adult": false,
			"fixedPrice": false,
			"customerReviewRank": 10,
			"subInfo": {
				"usedList": {
					"aladinUsed": {
						"itemCount": 1,
						"minPrice": 16700,
						"link": "https://www.aladin.co.kr/shop/UsedShop/wuseditemall.aspx?ItemId=284863481&amp;TabType=2&amp;partner=openAPI"
					},
					"userUsed": {
						"itemCount": 14,
						"minPrice": 19170,
						"link": "https://www.aladin.co.kr/shop/UsedShop/wuseditemall.aspx?ItemId=284863481&amp;TabType=1&amp;partner=openAPI"
					},
					"spaceUsed": {
						"itemCount": 13,
						"minPrice": 15900,
						"link": "https://www.aladin.co.kr/shop/UsedShop/wuseditemall.aspx?ItemId=284863481&amp;TabType=3&amp;partner=openAPI"
					}
				},
				"subTitle": "alone",
				"originalTitle": "",
				"itemPage": 272
			}
		}
	]
}`

func RunAladinApiMock(wg *sync.WaitGroup) {
	defer wg.Done()
	srv := &http.Server{Addr: ":8081"}
	http.HandleFunc("/ttb/api/ItemLookUp.aspx", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, RESP_ITEMLOOKUP)
	})
	go srv.ListenAndServe()
	fmt.Println("aladinAPI mock server is running")
}
