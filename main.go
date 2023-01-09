package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/lelemita/who_sells_all/searcher"
)

// TODO write test code
// // 1페이지 결과, 여러페이지 결과, 없는 결과
func main() {
	// TODO ttbkey 있는지 확인하고 없으면 exit
	genie := searcher.NewSearcher("https://www.aladin.co.kr")

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message": "hello"}`)
	})

	http.HandleFunc("/v1/proposals", func(w http.ResponseWriter, req *http.Request) {
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
