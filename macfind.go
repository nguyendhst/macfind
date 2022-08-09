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

func initializeDB() {
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

// Search searches the local database and in the case of not finding one, queries the remote API for the vendor name of the given MAC address.
func Search(hw string) (string, error) {
	initializeDB()
	oui, err := parse(hw)
	if err != nil {
		return "", fmt.Errorf("Search: failed to parse MAC address; %w", err)
	}
	// perform a local search first
	if DB_AVAIL {
		if localRes, err := searchDB(oui); err == nil {
			return localRes, nil
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
		return "", fmt.Errorf("Search: API connection timeout (5 secs); %w", err)
	}

	if err != nil {
		return "", fmt.Errorf("Search: failed to connect to API; %w", err)
	}
	defer res.Body.Close()

	// some devices using Android Q, iOS 14 and Windows 10 use randomized MAC addresses for enhanced privacy
	// status codes may differ depending on the API provider
	if res.StatusCode == http.StatusNotFound {
		return "?(randomized MAC)", nil
	} else if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Search: status code error: %d %s", res.StatusCode, res.Status)
	}

	var vendor []byte
	vendor, err = io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("main: failed to read response body; %w", err)
	}
	return string(vendor), nil
}

// SearchDB searches the local database (macfind_local) for the vendor name of the given MAC address.
func searchDB(oui string) (string, error) {
	var res string
	f, err := os.Open(LOCAL_DB_PATH)
	if err != nil {
		return "", fmt.Errorf("searchDB: failed to open local database; %w", err)
	}
	defer f.Close()
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
		return "", fmt.Errorf("searchDB: failed to scan local database; %w", err)
	} else if res == "" {
		return "?(randomized MAC)", nil
	}
	return res, nil
}

// Format checking and returns OUI (Organizationally Unique Identifier) of the given MAC address.
func parse(hw string) (string, error) {
	validMAC, err := regexp.Compile(`^[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}$`)
	if err != nil {
		return "", fmt.Errorf("parse: regex compilation failed; %w", err)
	}
	if !validMAC.MatchString(hw) {
		return "", fmt.Errorf("parse: invalid MAC address: %s; %w", hw, err)
	}
	// get OUI from MAC address
	return hw[0:8], nil
}
