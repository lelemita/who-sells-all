package searcher

const (
	PATH_ITEM_LOOK_UP   = "/ttb/api/ItemLookUp.aspx"
	PATH_USED_ITEM_MALL = "/shop/UsedShop/wuseditemall.aspx"
	TABTYPE_ALL         = 0
	TABTYPE_USER        = 1
	TABTYPE_ALADDIN     = 2
	TABTYPE_SPACE       = 3
	SORTORDER_LOW_PRICE = 9
)

type ItemLookUpResult struct {
	Item []struct {
		ItemId  uint `json:"itemId"`
		SubInfo struct {
			UsedList struct {
				AladinUsed UsedInfo `json:"aladinUsed"`
				UserUsed   UsedInfo `json:"userUsed"`
				SpaceUsed  UsedInfo `json:"spaceUsed"`
			} `json:"usedList"`
		} `json:"subInfo"`
	} `json:"item"`
}

type UsedInfo struct {
	ItemCount int    `json:"itemCount"`
	MinPrice  int    `json:"minPrice"`
	Link      string `json:"link"`
}

type SellerId string

type Proposal struct {
	Isbn   string
	Price  int
	Status string
	Link   string
}

type OneResult struct {
	Proposals map[SellerId]Proposal
}

type Searcher interface {
	GetByIsbn(isbn string) (*ItemLookUpResult, error)
	CrawlProposals(itemInfo ItemLookUpResult) (*OneResult, error)
}
