package main

import (
	"fmt"

	searcher "github.com/lelemita/who_sells_all/searcher/aladdin"
)

func main() {
	isbn := "9772799628000"
	genie := searcher.NewSearcher()
	itemLookUpResult, err := genie.GetByIsbn(isbn)
	if err != nil {
		panic(err)
	}

	proposals, err := genie.CrawlProposals(*itemLookUpResult)
	if err != nil {
		panic(err)
	}

	for k, v := range proposals.Proposals {
		fmt.Println(k, v)
	}
}
