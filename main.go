package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/lelemita/who_sells_all/searcher"
)

// TODO write test code
// // 1페이지 결과, 여러페이지 결과, 없는 결과
func main() {
	// TODO ttbkey 있는지 확인하고 없으면 exit
	ttbkey := os.Getenv("ttbkey")
	if len(ttbkey) == 0 {
		log.Fatal("ttbkey value is required")
	}
	genie := searcher.NewSearcher("https://www.aladin.co.kr", ttbkey)

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message": "hello"}`)
	})

	http.HandleFunc("/v1/proposals", func(w http.ResponseWriter, req *http.Request) {
		// TODO recover 대책이 필요하지 않나... 이걸로 되나... 점검 필요..
		// TODO os.Stderr 도 더 알아보기
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, `{"message": "error in process"}`)
				fmt.Fprintln(os.Stderr, err)
			}
		}()

		w.Header().Add("Content-Type", "application/json; charset=utf-8")
		qry, err := url.ParseQuery(req.URL.RawQuery)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			// TODO logger 추가하기
			fmt.Fprintf(w, `{"message": "error in ParseQuery"}`)
			return
		}
		isbns, isExist := qry["isbn"]
		if !isExist || len(isbns) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"message": "empty query isbn"}`)
			return
		}
		output := genie.GetOrderedList(isbns)
		jsonByte, err := json.Marshal(map[string]searcher.ShopList{"result": output})
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"message": "error in json.Marshal"}`)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, string(jsonByte))
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
