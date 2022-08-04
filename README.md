# Go MAC Address Lookup Package 
### The database (macfind_local) is taken from:

[Wireshark manufacturer database](https://gitlab.com/wireshark/wireshark/-/raw/master/manuf)
### Example:
```golang
package main

import (
	"log"

	mf "github.com/nguyendhst/macfind"
)

func main() {
	m1, err := mf.Search("18:65:90:dc:c0:cb")
	if err != nil {
		log.Fatal(err)
	}
	log.Print(m1) //Apple, Inc.
}
```