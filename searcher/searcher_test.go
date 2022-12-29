package searcher_test

import (
	"os"
	"sync"
	"testing"

	"github.com/lelemita/who_sells_all/mock"
	mock_searcher "github.com/lelemita/who_sells_all/searcher/mock"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	var wg sync.WaitGroup
	wg.Add(1)
	mock.RunAladdinApiMock(&wg)
	wg.Wait()
	exitVal := m.Run()
	os.Exit(exitVal)
}

func TestFirstItemLookUp(t *testing.T) {
	searcher := mock_searcher.NewSearcher()

	testIsbn := "9772799628000"
	tests := []struct {
		title string
		isbn  string
		err   error
	}{
		{"OK", testIsbn, nil},
	}

	for _, tc := range tests {
		assert := assert.New(t)
		t.Run(tc.title, func(t *testing.T) {
			result, err := searcher.FirstItemLookUp(tc.isbn)
			if tc.err == nil {
				assert.Nil(err)
				assert.NotEmpty(result)
				assert.Equal(result.ItemId, uint(284863481))
				assert.Equal(result.SubInfo.UsedList.AladinUsed.ItemCount, 1)
				assert.Equal(result.SubInfo.UsedList.UserUsed.ItemCount, 14)
				assert.Equal(result.SubInfo.UsedList.SpaceUsed.ItemCount, 13)
			}
		})
	}
}

// TODO: 테스트 추가하기. WSL 에서 go tool cover -html=cover.out 보기
