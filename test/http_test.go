package crawler

import (
	"net/http"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/Pixep/crowlet/pkg/crawler"
)

var waitMutex = &sync.Mutex{}
var fetchedUrls []string

func TestRunConcurrentGet(t *testing.T) {
	resultChan := make(chan *crawler.HTTPResponse)
	quitChan := make(chan struct{})
	maxConcurrency := 3
	urls := []string{
		"url1",
		"url2",
		"url3",
		"url4",
		"url5",
	}

	waitMutex.Lock()
	go crawler.RunConcurrentGet(mockHTTPGet, urls, crawler.HTTPConfig{}, maxConcurrency, resultChan, quitChan)
	time.Sleep(2 * time.Second)

	if len(fetchedUrls) != maxConcurrency {
		t.Fatal("Incorrect channel length of", len(fetchedUrls))
		t.Fail()
	}

	waitMutex.Unlock()

	resultChanOpen := true
	for resultChanOpen == true {
		select {
		case _, resultChanOpen = <-resultChan:
		}
	}

	sort.Strings(fetchedUrls)
	if !testEq(fetchedUrls, urls) {
		t.Fatal("Expected to crawl ", urls, " but crawled ", fetchedUrls, " instead.")
		t.Fail()
	}
}

func mockHTTPGet(client *http.Client, url string, config crawler.HTTPConfig) *crawler.HTTPResponse {
	fetchedUrls = append(fetchedUrls, url)
	waitMutex.Lock()
	waitMutex.Unlock()

	return &crawler.HTTPResponse{URL: url}
}

func testEq(a, b []string) bool {

	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
