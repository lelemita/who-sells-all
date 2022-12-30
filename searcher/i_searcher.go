package searcher

const (
	PATH_ITEM_LOOK_UP   = "/ttb/api/ItemLookUp.aspx"
	PATH_USED_ITEM_MALL = "/shop/UsedShop/wuseditemall.aspx"
)

type ItemLookUpList struct {
	Item []ItemLookUpResult `json:"item"`
}

type ItemLookUpResult struct {
	ItemId  uint `json:"itemId"`
	SubInfo struct {
		UsedList struct {
			AladinUsed UsedInfo `json:"aladinUsed"`
			UserUsed   UsedInfo `json:"userUsed"`
			SpaceUsed  UsedInfo `json:"spaceUsed"`
		} `json:"usedList"`
	} `json:"subInfo"`
}

type UsedInfo struct {
	ItemCount int    `json:"itemCount"`
	MinPrice  int    `json:"minPrice"`
	Link      string `json:"link"`
}

type SellerName string
type Bidding map[SellerName]Seller

type Seller struct {
	Link        string
	DeliveryFee string
	Proposal    []Book
}

type Book struct {
	ItemId string
	Price  uint
	Status string
	Link   string
}

type Searcher interface {
	GetProposals(isbns []string) Bidding
}
