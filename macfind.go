package macfind

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"runtime"
	"time"
)

const (
	API_URL  = "http://api.macvendors.com/%s" // change to your own API URL as needed
	TIMEOUT  = time.Second * 5                // change to your own timeout as needed
	LOCAL_DB = "macfind_local"
)

var DB_AVAIL = true
var LOCAL_DB_PATH = ""

func initialize() {
	// check db availability
	_, callerPath, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Println("runtime.Caller() failed -- exiting")
		os.Exit(1)
	}
	LOCAL_DB_PATH = fmt.Sprintf("%s/%s", path.Dir(callerPath), LOCAL_DB)
	if _, err := os.Stat(LOCAL_DB_PATH); err != nil {
		DB_AVAIL = false
	}
}

// Search searches the remote API for the vendor name of the given MAC address.
func Search(hw string) (string, error) {
	initialize()
	oui, err := parse(hw)
	if err != nil {
		return "", err
	}
	// perform a local search first
	if DB_AVAIL {
		if localRes, err := searchDB(oui); err == nil {
			return localRes + " --from db", nil
		}
	}
	// perform a remote search
	var res *http.Response

	url := fmt.Sprintf(API_URL, url.QueryEscape(hw))
	done := make(chan struct{})
	defer close(done)

	go func() {
		res, err = http.Get(url)
		done <- struct{}{}
	}()

	select {
	case <-done:
	case <-time.After(TIMEOUT):
		return searchDB(hw)
	}

	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return "?(randomized MAC)", nil
	} else if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	var vendor []byte
	vendor, err = io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(vendor) + " --from api", nil
}

// SearchDB searches the local database (macfind_local) for the vendor name of the given MAC address.
func searchDB(oui string) (string, error) {
	var res string
	f, err := os.Open(LOCAL_DB_PATH)
	if err != nil {
		return "", err
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		// example format:
		//FC:F1:CD	Optex-Fa	Optex-Fa Co.,Ltd.
		if line[0:8] == oui {
			res = line[9:]
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return res, nil
}

// Format checking and returns OUI (Organizationally Unique Identifier) of the given MAC address.
func parse(hw string) (string, error) {
	validMAC, err := regexp.Compile(`^[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}$`)
	if err != nil {
		return "", err
	}
	if !validMAC.MatchString(hw) {
		return "", fmt.Errorf("invalid MAC address: %s", hw)
	}
	// get OUI from MAC address
	return hw[0:8], nil
}
