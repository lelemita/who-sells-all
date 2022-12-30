package main

import (
	"fmt"

	searcher "github.com/lelemita/who_sells_all/searcher/aladin"
)

func main() {
	genie := searcher.NewSearcher()
	// 8970127240
	isbn := "9772799628000"
	itemId := genie.GetIdByIsbn(isbn)
	books := genie.CrawlProposals(itemId)

	for k, v := range books {
		fmt.Println(k, v)
	}

}

// func getProposals(isbn string) *searcher.Bidding {

// }
