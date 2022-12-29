package main

import (
	"fmt"

	searcher "github.com/lelemita/who_sells_all/searcher/aladdin"
)

func main() {
	isbn := "9772799628000"
	genie := searcher.NewSearcher()
	itemId := genie.GetIdByIsbn(isbn)
	books := genie.CrawlProposals(itemId)

	for k, v := range books {
		fmt.Println(k, v)
	}
}
