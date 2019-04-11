package crawler

import (
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
	time.Sleep(time.Second)

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

func mockHTTPGet(url string, config crawler.HTTPConfig) *crawler.HTTPResponse {
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

func TestMergeCrawlStats(t *testing.T) {
	statsA := crawler.CrawlStats{
		Total:          10,
		StatusCodes:    map[int]int{200: 10},
		Average200Time: time.Duration(1) * time.Second,
		Max200Time:     time.Duration(2) * time.Second,
	}

	statsB := crawler.CrawlStats{
		Total:          6,
		StatusCodes:    map[int]int{200: 2, 404: 4},
		Average200Time: time.Duration(7) * time.Second,
		Max200Time:     time.Duration(9) * time.Second,
	}

	stats := crawler.MergeCrawlStats(statsA, statsB)

	if stats.Total != 16 {
		t.Fatal("Invalid total", stats.Total)
		t.Fail()
	}

	if stats.StatusCodes[200] != 12 ||
		stats.StatusCodes[404] != 4 {
		t.Fatal("Invalid status codes count")
		t.Fail()
	}

	if stats.Average200Time != time.Duration(2)*time.Second {
		t.Fatal("Invalid average 200 time:", stats.Average200Time)
		t.Fail()
	}

	if stats.Max200Time != time.Duration(9)*time.Second {
		t.Fatal("Invalid maximum 200 time:", stats.Max200Time)
		t.Fail()
	}
}
