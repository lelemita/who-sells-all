package whosellsall

import (
	"os"
)

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

type Searcher struct {
	ttbkey string
}

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

func NewSearcher() Searcher {
	ttbkey := os.Getenv("ttbkey")
	return Searcher{ttbkey}
}

func (s *Searcher) GetByIsbn(isbn string) (OneResult, error) {
	return OneResult{}, nil
}
