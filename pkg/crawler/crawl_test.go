package crawler

import (
	"net/url"
	"reflect"
	"slices"
	"testing"
	"time"
)

func TestMergeCrawlStats(t *testing.T) {
	tests := []struct {
		name           string
		statsA         CrawlStats
		statsB         CrawlStats
		expectedTotal  int
		expectedCodes  map[int]int
		expectedAvg200 time.Duration
		expectedMax200 time.Duration
		expectedNon200 []CrawlResult
	}{
		{
			name: "Basic merge with 200 and 404",
			statsA: CrawlStats{
				Total:          10,
				StatusCodes:    map[int]int{200: 10},
				Average200Time: time.Second,
				Max200Time:     2 * time.Second,
			},
			statsB: CrawlStats{
				Total:          4,
				StatusCodes:    map[int]int{200: 2, 404: 2},
				Average200Time: 7 * time.Second,
				Max200Time:     9 * time.Second,
				Non200Urls: []CrawlResult{
					{URL: "http://example.com", StatusCode: 404, LinkingURLs: []string{"http://example.com/1"}},
					{URL: "http://example.com/2", StatusCode: 404}},
			},
			expectedTotal:  14,
			expectedCodes:  map[int]int{200: 12, 404: 2},
			expectedAvg200: 2 * time.Second,
			expectedMax200: 9 * time.Second,
			expectedNon200: []CrawlResult{
				{URL: "http://example.com", StatusCode: 404, LinkingURLs: []string{"http://example.com/1"}},
				{URL: "http://example.com/2", StatusCode: 404}},
		},
		{
			name: "One empty stats input",
			statsA: CrawlStats{
				Total:          0,
				StatusCodes:    nil,
				Average200Time: 0,
				Max200Time:     0,
			},
			statsB: CrawlStats{
				Total:          5,
				StatusCodes:    map[int]int{200: 5},
				Average200Time: 3 * time.Second,
				Max200Time:     4 * time.Second,
			},
			expectedTotal:  5,
			expectedCodes:  map[int]int{200: 5},
			expectedAvg200: 3 * time.Second,
			expectedMax200: 4 * time.Second,
		},
		{
			name: "Disjoint status codes",
			statsA: CrawlStats{
				Total:          8,
				StatusCodes:    map[int]int{301: 3, 500: 5},
				Average200Time: 0,
				Max200Time:     0,
				Non200Urls: []CrawlResult{
					{URL: "http://example.com", StatusCode: 301}},
			},
			statsB: CrawlStats{
				Total:          7,
				StatusCodes:    map[int]int{200: 7},
				Average200Time: 2 * time.Second,
				Max200Time:     3 * time.Second,
				Non200Urls: []CrawlResult{
					{URL: "http://other.com", StatusCode: 500}},
			},
			expectedTotal:  15,
			expectedCodes:  map[int]int{200: 7, 301: 3, 500: 5},
			expectedAvg200: 2 * time.Second,
			expectedMax200: 3 * time.Second,
			expectedNon200: []CrawlResult{
				{URL: "http://example.com", StatusCode: 301},
				{URL: "http://other.com", StatusCode: 500}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := MergeCrawlStats(tt.statsA, tt.statsB)

			if stats.Total != tt.expectedTotal {
				t.Fatalf("Total mismatch: expected %d, got %d", tt.expectedTotal, stats.Total)
			}

			for code, count := range tt.expectedCodes {
				if stats.StatusCodes[code] != count {
					t.Fatalf("Status code %d mismatch: expected %d, got %d", code, count, stats.StatusCodes[code])
				}
			}

			if stats.Average200Time != tt.expectedAvg200 {
				t.Fatalf("Average200Time mismatch: expected %v, got %v", tt.expectedAvg200, stats.Average200Time)
			}

			if stats.Max200Time != tt.expectedMax200 {
				t.Fatalf("Max200Time mismatch: expected %v, got %v", tt.expectedMax200, stats.Max200Time)
			}

			if len(stats.Non200Urls) != len(tt.expectedNon200) {
				t.Fatalf("Non200Urls length mismatch: expected %d, got %d", len(tt.expectedNon200), len(stats.Non200Urls))
			}
			for i, result := range stats.Non200Urls {
				if result.URL != tt.expectedNon200[i].URL {
					t.Fatalf("Non200Urls URL mismatch at index %d: expected %s, got %s", i, tt.expectedNon200[i].URL, result.URL)
				}
				if result.StatusCode != tt.expectedNon200[i].StatusCode {
					t.Fatalf("Non200Urls StatusCode mismatch at index %d: expected %d, got %d", i, tt.expectedNon200[i].StatusCode, result.StatusCode)
				}
				if !slices.Equal(result.LinkingURLs, tt.expectedNon200[i].LinkingURLs) {
					t.Fatalf("Non200Urls LinkingURLs mismatch at index %d: expected %v, got %v", i, tt.expectedNon200[i].LinkingURLs, result.LinkingURLs)
				}
			}
		})
	}
}

func Test_getLinksToCrawl(t *testing.T) {
	type args struct {
		sourceResults map[string]*HTTPResponse
		sourceConfig  CrawlConfig
	}
	tests := []struct {
		name string
		args args
		want map[string][]string
	}{
		{
			name: "Hyperlinks disabled",
			args: args{
				sourceResults: map[string]*HTTPResponse{
					"//example.com": {
						URL: "//example.com",
						Links: []Link{
							{Type: Hyperlink, TargetURL: url.URL{Host: "example.com", Path: "/1"}, IsExternal: false},
							{Type: Hyperlink, TargetURL: url.URL{Host: "example.com", Path: "/2"}, IsExternal: false},
						},
					},
				},
				sourceConfig: CrawlConfig{
					Links: CrawlPageLinksConfig{
						CrawlExternalLinks: false,
						CrawlHyperlinks:    false,
						CrawlImages:        false,
					},
				},
			},
			want: map[string][]string{},
		},
		{
			name: "Hyperlinks enabled",
			args: args{
				sourceResults: map[string]*HTTPResponse{
					"//example.com": {
						URL: "//example.com",
						Links: []Link{
							{Type: Hyperlink, TargetURL: url.URL{Host: "example.com", Path: "/1"}, IsExternal: false},
							{Type: Hyperlink, TargetURL: url.URL{Host: "example.com", Path: "/2"}, IsExternal: false},
						},
					},
				},
				sourceConfig: CrawlConfig{
					Links: CrawlPageLinksConfig{
						CrawlExternalLinks: false,
						CrawlHyperlinks:    true,
						CrawlImages:        false,
					},
				},
			},
			want: map[string][]string{
				"//example.com/1": {"//example.com"},
				"//example.com/2": {"//example.com"},
			},
		},
		{
			name: "External links disabled",
			args: args{
				sourceResults: map[string]*HTTPResponse{
					"//example.com": {
						URL: "//example.com",
						Links: []Link{
							{Type: Hyperlink, TargetURL: url.URL{Host: "example.com", Path: "/1"}, IsExternal: false},
							{Type: Hyperlink, TargetURL: url.URL{Host: "example.com", Path: "/2"}, IsExternal: false},
							{Type: Hyperlink, TargetURL: url.URL{Host: "external.com"}, IsExternal: true},
						},
					},
				},
				sourceConfig: CrawlConfig{
					Links: CrawlPageLinksConfig{
						CrawlExternalLinks: false,
						CrawlHyperlinks:    true,
						CrawlImages:        false,
					},
				},
			},
			want: map[string][]string{
				"//example.com/1": {"//example.com"},
				"//example.com/2": {"//example.com"},
			},
		},
		{
			name: "External links enabled",
			args: args{
				sourceResults: map[string]*HTTPResponse{
					"//example.com": {
						URL: "//example.com",
						Links: []Link{
							{Type: Hyperlink, TargetURL: url.URL{Host: "example.com", Path: "/1"}, IsExternal: false},
							{Type: Hyperlink, TargetURL: url.URL{Host: "example.com", Path: "/2"}, IsExternal: false},
							{Type: Hyperlink, TargetURL: url.URL{Host: "external.com"}, IsExternal: true},
						},
					},
				},
				sourceConfig: CrawlConfig{
					Links: CrawlPageLinksConfig{
						CrawlExternalLinks: true,
						CrawlHyperlinks:    true,
						CrawlImages:        false,
					},
				},
			},
			want: map[string][]string{
				"//example.com/1": {"//example.com"},
				"//example.com/2": {"//example.com"},
				"//external.com":  {"//example.com"},
			},
		},
		{
			name: "Images disabled",
			args: args{
				sourceResults: map[string]*HTTPResponse{
					"//example.com": {
						URL: "//example.com",
						Links: []Link{
							{Type: Hyperlink, TargetURL: url.URL{Host: "example.com", Path: "/1"}, IsExternal: false},
							{Type: Hyperlink, TargetURL: url.URL{Host: "example.com", Path: "/2"}, IsExternal: false},
							{Type: Image, TargetURL: url.URL{Host: "example.com", Path: "/image.jpg"}, IsExternal: false},
						},
					},
				},
				sourceConfig: CrawlConfig{
					Links: CrawlPageLinksConfig{
						CrawlExternalLinks: false,
						CrawlHyperlinks:    true,
						CrawlImages:        false,
					},
				},
			},
			want: map[string][]string{
				"//example.com/1": {"//example.com"},
				"//example.com/2": {"//example.com"},
			},
		},
		{
			name: "Images enabled",
			args: args{
				sourceResults: map[string]*HTTPResponse{
					"//example.com": {
						URL: "//example.com",
						Links: []Link{
							{Type: Hyperlink, TargetURL: url.URL{Host: "example.com", Path: "/1"}, IsExternal: false},
							{Type: Hyperlink, TargetURL: url.URL{Host: "example.com", Path: "/2"}, IsExternal: false},
							{Type: Image, TargetURL: url.URL{Host: "example.com", Path: "/image.jpg"}, IsExternal: false},
						},
					},
				},
				sourceConfig: CrawlConfig{
					Links: CrawlPageLinksConfig{
						CrawlExternalLinks: false,
						CrawlHyperlinks:    true,
						CrawlImages:        true,
					},
				},
			},
			want: map[string][]string{
				"//example.com/1":         {"//example.com"},
				"//example.com/2":         {"//example.com"},
				"//example.com/image.jpg": {"//example.com"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getLinksToCrawl(tt.args.sourceResults, tt.args.sourceConfig); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getLinksToCrawl() = %v, want %v", got, tt.want)
			}
		})
	}
}

type RequestTime struct {
	startTime time.Time
}

// Make RequestDuration implement the RequestTimer interface
func (r *RequestTime) Total(endTime time.Time) time.Duration {
	return endTime.Sub(r.startTime)
}

func Test_populateCrawlStats(t *testing.T) {
	type args struct {
		result *HTTPResponse
		stats  *CrawlStats
	}

	tests := []struct {
		name string
		args args
		want *CrawlStats
	}{
		{
			name: "first 200 status code",
			args: args{
				result: &HTTPResponse{
					URL:        "http://example.com",
					StatusCode: 200,
					EndTime:    time.UnixMilli(11000),
					Result: &RequestTime{
						startTime: time.UnixMilli(10000),
					},
				},
				stats: &CrawlStats{
					StatusCodes: make(map[int]int),
				},
			},
			want: &CrawlStats{
				Total:          1,
				StatusCodes:    map[int]int{200: 1},
				Total200Time:   1000 * time.Millisecond,
				Average200Time: 1000 * time.Millisecond,
				Max200Time:     1000 * time.Millisecond,
			},
		},
		{
			name: "update with 200 status code",
			args: args{
				result: &HTTPResponse{
					URL:        "http://example.com",
					StatusCode: 200,
					EndTime:    time.UnixMilli(10500),
					Result: &RequestTime{
						startTime: time.UnixMilli(10000),
					},
				},
				stats: &CrawlStats{
					Total:          1,
					StatusCodes:    map[int]int{200: 1},
					Total200Time:   1000 * time.Millisecond,
					Average200Time: 1000 * time.Millisecond,
					Max200Time:     1000 * time.Millisecond,
				},
			},
			want: &CrawlStats{
				Total:          2,
				StatusCodes:    map[int]int{200: 2},
				Total200Time:   1500 * time.Millisecond,
				Average200Time: 750 * time.Millisecond,
				Max200Time:     1000 * time.Millisecond,
			},
		},
		{
			name: "update with 200 status code with higher time",
			args: args{
				result: &HTTPResponse{
					URL:        "http://example.com",
					StatusCode: 200,
					EndTime:    time.UnixMilli(12000),
					Result: &RequestTime{
						startTime: time.UnixMilli(10000),
					},
				},
				stats: &CrawlStats{
					Total:          1,
					StatusCodes:    map[int]int{200: 1},
					Total200Time:   1000 * time.Millisecond,
					Average200Time: 1000 * time.Millisecond,
					Max200Time:     2000 * time.Millisecond,
				},
			},
			want: &CrawlStats{
				Total:          2,
				StatusCodes:    map[int]int{200: 2},
				Total200Time:   3000 * time.Millisecond,
				Average200Time: 1500 * time.Millisecond,
				Max200Time:     2000 * time.Millisecond,
			},
		},
		{
			name: "update with 404 status code",
			args: args{
				result: &HTTPResponse{
					URL:        "http://example.com",
					StatusCode: 404,
					EndTime:    time.UnixMilli(12000),
					Result: &RequestTime{
						startTime: time.UnixMilli(10000),
					},
				},
				stats: &CrawlStats{
					StatusCodes: make(map[int]int),
					Non200Urls:  []CrawlResult{},
				},
			},
			want: &CrawlStats{
				Total:       1,
				StatusCodes: map[int]int{404: 1},
				Non200Urls:  []CrawlResult{{URL: "http://example.com", StatusCode: 404, Time: 2000 * time.Millisecond}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateCrawlStats(tt.args.result, tt.args.stats)
			if !reflect.DeepEqual(tt.args.stats, tt.want) {
				t.Errorf("populateCrawlStats() = %v, want %v", tt.args.stats, tt.want)
			}
		})
	}
}
