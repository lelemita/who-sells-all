package mock_test

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"testing"

	"github.com/lelemita/who_sells_all/mock"
	"github.com/lelemita/who_sells_all/searcher"
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

func TestItemLookUp(t *testing.T) {
	tests := []struct {
		title      string
		statusCode int
		err        error
	}{
		{"OK", http.StatusOK, nil},
	}

	for _, tc := range tests {
		assert := assert.New(t)
		t.Run(tc.title, func(t *testing.T) {
			resp, err := http.Get("http://localhost:8081/ttb/api/ItemLookUp.aspx")
			assert.Nil(err)
			assert.NotNil(resp)
			defer resp.Body.Close()
			assert.Equal(tc.statusCode, resp.StatusCode)

			jsonByte, err := io.ReadAll(resp.Body)
			assert.Nil(err)
			respInfo := searcher.ItemLookUpResult{}
			if err := json.Unmarshal(jsonByte, &respInfo); err != nil {
				log.Fatalf("error during parsing json: %v", err)
			}

			mockInfo := searcher.ItemLookUpResult{}
			if err := json.Unmarshal([]byte(mock.RESP_ITEMLOOKUP), &mockInfo); err != nil {
				log.Fatalf("error during parsing mock data: %v", err)
			}
			assert.Equal(mockInfo, respInfo)
		})
	}
}
