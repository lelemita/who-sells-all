package searcher_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/lelemita/who_sells_all/mock"
	mock_searcher "github.com/lelemita/who_sells_all/searcher/mock"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	go mock.RunAladdinApiMock()
	exitVal := m.Run()
	fmt.Println("Clean up stuff after test here")
	os.Exit(exitVal)
}

func TestGetByIsbn(t *testing.T) {
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
			result, err := searcher.GetByIsbn(tc.isbn)
			if tc.err == nil {
				assert.Nil(err)
				assert.NotEmpty(result)
				assert.Len(result.Item, 1)
				assert.Equal(result.Item[0].ItemId, uint(284863481))
				assert.Equal(result.Item[0].SubInfo.UsedList.AladinUsed.ItemCount, 1)
				assert.Equal(result.Item[0].SubInfo.UsedList.UserUsed.ItemCount, 14)
				assert.Equal(result.Item[0].SubInfo.UsedList.SpaceUsed.ItemCount, 13)
			}
		})
	}

}
