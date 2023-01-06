package main

import (
	"fmt"
	"sort"

	"github.com/lelemita/who_sells_all/searcher"
)

type Shop struct {
	Name        searcher.SellerName
	TotalPrice  uint
	DeliveryFee string
}

type ShopList []Shop

func (s ShopList) Len() int {
	return len(s)
}

func (s ShopList) Less(i, j int) bool {
	return s[i].TotalPrice < s[j].TotalPrice
}

func (s ShopList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func main() {
	genie := searcher.NewSearcher("https://www.aladin.co.kr")
	isbns := []string{"9791164136452", "9788970128856"}
	proposals := genie.GetProposals(isbns)
	fmt.Println("total seller: ", len(proposals))

	output := ShopList{}
	for sName, seller := range proposals {
		shop := Shop{Name: sName, DeliveryFee: seller.DeliveryFee}
		for _, book := range seller.Proposal {
			shop.TotalPrice += book.Price
		}
		// Think 배송비 반영 해야 하려나? 얼마이상 무료배송 이런거 까지 고려해야 할지도..
		output = append(output, shop)
	}

	sort.Sort(output)
	for i, shop := range output {
		fmt.Printf("%d) %s: %d (%s)\n", i+1, shop.Name, shop.TotalPrice, shop.DeliveryFee)
	}
}
