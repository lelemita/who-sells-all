package whosellsall_test

import (
	"fmt"
	"os"
	"testing"

	whosellsall "github.com/lelemita/who_sells_all"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	fmt.Println("Set up stuff for tests here")
	exitVal := m.Run()
	fmt.Println("Clean up stuff after test here")
	os.Exit(exitVal)
}

func TestGetByIsbn(t *testing.T) {
	searcher := whosellsall.NewSearcher()

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
			result, err := searcher.GetByIsbn("9772799628000")
			assert.Nil(err)
			assert.NotEmpty(result)
		})
	}

}
