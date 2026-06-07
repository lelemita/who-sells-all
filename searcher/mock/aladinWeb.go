package mock

import (
	_ "embed"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// 동일한 책을 2권이상 가진 판매자가 있는 페이지
//
//go:embed page-used-item-mall.html
var PAGE_USED_ITEM_MALL string

func RunAladinWeb(wg *sync.WaitGroup) {
	defer wg.Done()
	srv := &http.Server{Addr: ":8082"}
	http.HandleFunc("/shop/UsedShop/wuseditemall.aspx", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, PAGE_USED_ITEM_MALL)
	})
	go srv.ListenAndServe()
	time.Sleep(time.Millisecond * 100)
	fmt.Println("aladin WEB mock server is running")
}
