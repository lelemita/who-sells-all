package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/lelemita/who_sells_all/searcher"
)

var templates = template.Must(template.ParseGlob("templates/*.html"))

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
		err := templates.ExecuteTemplate(w, "index.html", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
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

	http.HandleFunc("/v1/search", func(w http.ResponseWriter, req *http.Request) {
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
			fmt.Fprintf(w, `{"message": "error in ParseQuery"}`)
			return
		}
		q, isExist := qry["q"]
		if !isExist || len(q) == 0 || q[0] == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"message": "empty query q"}`)
			return
		}
		output, err := genie.Search(q[0])
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `{"message": "%s"}`, err.Error())
			return
		}
		jsonByte, err := json.Marshal(output)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"message": "error in json.Marshal"}`)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, string(jsonByte))
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
