package mock_test

import (
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/lelemita/who_sells_all/mock"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	aliddinMock := mock.RunMockServer()
	exitVal := m.Run()
	if err := aliddinMock.Close(); err != nil {
		log.Fatalf("error during aladdin mock server stop: %v", err)
	}
	os.Exit(exitVal)
}

func TestItemLookUp(t *testing.T) {
	testIsbn := "9772799628000"
	tests := []struct {
		title string
		isbn  string
		err   error
	}{
		{testIsbn, "OK", nil},
	}

	for _, tc := range tests {
		assert := assert.New(t)
		t.Run(tc.title, func(t *testing.T) {
			resp, err := http.Get("http://localhost:8081/api/itemLookUp.aspx")
			assert.Nil(err)
			assert.NotNil(resp)
			// TODO json 파싱해서 비교해야 하나...
			assert.Equal(mock.RESP_ITEMLOOKUP, resp.Body)
		})
	}

}
