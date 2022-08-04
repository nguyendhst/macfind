package macfind_test

import (
	"fmt"
	"testing"

	mf "github.com/nguyendhst/macfind"
)

func TestSearch(t *testing.T) {
	vendor, err := mf.Search("FC:FB:FB:01:FA:21")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(vendor)
}
