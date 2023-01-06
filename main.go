package main

import (
	"fmt"

	"github.com/lelemita/who_sells_all/searcher"
)

func main() {
	genie := searcher.NewSearcher("https://www.aladin.co.kr")
	isbns := []string{"9791164136452", "9788970128856"}
	output := genie.GetOrderedList(isbns)

	for i, shop := range output {
		fmt.Printf("%d) %s: %d (%s) - %s\n", i+1, shop.Name, shop.TotalPrice, shop.DeliveryFee, "https://www.aladin.co.kr"+shop.Link)
	}
}
