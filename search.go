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

func NewSearcher() Searcher {
	ttbkey := os.Getenv("ttbkey")
	return Searcher{ttbkey}
}

func (s *Searcher) GetByIsbn(isbn string) (OneResult, error) {
	return OneResult{}, nil
}
