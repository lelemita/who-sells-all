package main

import (
	"fmt"

	searcher "github.com/lelemita/who_sells_all/searcher/aladin"
)

func main() {
	genie := searcher.NewSearcher()
	isbns := []string{"9791164136452", "9788970128856"}
	proposals := genie.GetProposals(isbns)

	fmt.Println("total seller: ", len(proposals))
	for sName, seller := range proposals {
		var totalPrice uint
		for _, book := range seller.Proposal {
			totalPrice += book.Price
		}
		fmt.Printf("%s: %d (%s)\n", sName, totalPrice, seller.DeliveryFee)
	}

}
