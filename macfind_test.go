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

func TestRandomizedMac(t *testing.T) {
	ven, err := mf.Search("6A:D5:DC:A5:F9:1B")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(ven)
}

func TestBadMACAddress(t *testing.T) {
	ven, err := mf.Search("6A:D5:DC:Avv:F9:1B:1B")
	if err != nil {
		fmt.Println(err)
	} else {
		t.Error(ven)
	}
}
