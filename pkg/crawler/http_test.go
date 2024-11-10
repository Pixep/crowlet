package crawler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"sync"
	"testing"
	"time"
)

var waitMutex = &sync.Mutex{}
var resultMutex = &sync.Mutex{}
var fetchedUrls []string

func TestRunConcurrentGet(t *testing.T) {
	resultChan := make(chan *HTTPResponse)
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
	go RunConcurrentGet(mockHTTPGet, urls, HTTPConfig{}, maxConcurrency, resultChan, quitChan)
	time.Sleep(2 * time.Second)

	if len(fetchedUrls) != maxConcurrency {
		t.Fatal("Incorrect channel length of", len(fetchedUrls))
		t.Fail()
	}

	waitMutex.Unlock()

	resultChanOpen := true
	for resultChanOpen == true {
		_, resultChanOpen = <-resultChan
	}

	sort.Strings(fetchedUrls)
	if !testEq(fetchedUrls, urls) {
		t.Fatal("Expected to crawl ", urls, " but crawled ", fetchedUrls, " instead.")
		t.Fail()
	}
}

func mockHTTPGet(client *http.Client, url string, config HTTPConfig) *HTTPResponse {
	resultMutex.Lock()
	fetchedUrls = append(fetchedUrls, url)
	resultMutex.Unlock()
	waitMutex.Lock()
	waitMutex.Unlock()

	return &HTTPResponse{URL: url}
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

// TestHTTPGet tests the HTTPGet function
func TestHTTPGet(t *testing.T) {
	tests := []struct {
		name          string
		urlStr        string
		config        HTTPConfig
		statusCode    int
		body          string
		expectedLinks []Link
		doNotServe    bool
	}{
		{
			name:          "Successful GET request without link parsing",
			urlStr:        "/test",
			config:        HTTPConfig{ParseLinks: false, Timeout: 5 * time.Second},
			statusCode:    http.StatusOK,
			body:          `<html><a href="https://example.com"></a></html>`,
			expectedLinks: []Link{}, // No links expected since ParseLinks is false
			doNotServe:    false,
		},
		{
			name:       "Successful GET request with link parsing",
			urlStr:     "/test",
			config:     HTTPConfig{ParseLinks: true, Timeout: 5 * time.Second},
			statusCode: http.StatusOK,
			body:       `<html><a href="https://example.com"></a><img src="https://example.com/image.png"></html>`,
			expectedLinks: []Link{
				{
					Type:       Hyperlink,
					TargetURL:  mustParseURL("https://example.com"),
					IsExternal: true,
				},
				{
					Type:       Image,
					TargetURL:  mustParseURL("https://example.com/image.png"),
					IsExternal: true,
				},
			},
			doNotServe: false,
		},
		{
			name:          "Server error response",
			urlStr:        "/error",
			config:        HTTPConfig{ParseLinks: false, Timeout: 5 * time.Second},
			statusCode:    http.StatusInternalServerError,
			body:          "Internal Server Error",
			expectedLinks: []Link{}, // No links expected on error
			doNotServe:    false,
		},
		{
			name:          "Server error response",
			urlStr:        "/error",
			config:        HTTPConfig{ParseLinks: false, Timeout: 500 * time.Millisecond},
			statusCode:    0,
			body:          "",
			expectedLinks: []Link{}, // No links expected on error
			doNotServe:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urlStr := "http://localhost:61111"
			if !tt.doNotServe {
				// Setup mock HTTP server
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tt.statusCode)
					w.Write([]byte(tt.body))
				}))
				urlStr = server.URL + tt.urlStr
				defer server.Close()
			}

			// Initialize client with a timeout
			client := &http.Client{Timeout: tt.config.Timeout}

			// Call HTTPGet function
			response := HTTPGet(client, urlStr, tt.config)

			// Validate error presence as expected
			if (response.Err != nil) != tt.doNotServe {
				t.Errorf("expected error: %v, got: %v", tt.doNotServe, response.Err)
			}

			// Validate status code
			if response.StatusCode != tt.statusCode {
				t.Errorf("expected status code: %d, got: %d", tt.statusCode, response.StatusCode)
			}

			// Validate parsed links if ParseLinks is enabled
			if tt.config.ParseLinks {
				if len(response.Links) != len(tt.expectedLinks) {
					t.Errorf("expected %d links, got %d", len(tt.expectedLinks), len(response.Links))
				}
				for i, link := range response.Links {
					expectedLink := tt.expectedLinks[i]
					if link.Type != expectedLink.Type {
						t.Errorf("expected link type: %v, got: %v", expectedLink.Type, link.Type)
					}
					if link.TargetURL.String() != expectedLink.TargetURL.String() {
						t.Errorf("expected target URL: %s, got: %s", expectedLink.TargetURL.String(), link.TargetURL.String())
					}
					if link.IsExternal != expectedLink.IsExternal {
						t.Errorf("expected IsExternal: %v, got: %v", expectedLink.IsExternal, link.IsExternal)
					}
				}
			}

			// Validate the Result field is set
			if response.Result == nil {
				t.Errorf("expected Result to be initialized, got nil")
			}

			// Validate EndTime is set to a non-zero value
			if response.EndTime.IsZero() {
				t.Errorf("expected EndTime to be set, got zero value")
			}
		})
	}
}

// Helper function to parse URLs and handle errors inline
func mustParseURL(rawurl string) url.URL {
	parsedURL, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}
	return *parsedURL
}
