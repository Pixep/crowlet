package crawler

import (
	"errors"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/yterajima/go-sitemap"
)

// CrawlResult is the result from a single crawling
type CrawlResult struct {
	URL        string        `json:"url"`
	StatusCode int           `json:"status-code"`
	Time       time.Duration `json:"server-time"`
}

// CrawlStats holds crawling related information: status codes, time
// and totals
type CrawlStats struct {
	Total          int
	StatusCodes    map[int]int
	Average200Time time.Duration
	Max200Time     time.Duration
	Non200Urls     []CrawlResult
}

// CrawlConfig holds crawling configuration.
type CrawlConfig struct {
	Throttle   int
	Host       string
	HTTP       HTTPConfig
	HTTPGetter ConcurrentHTTPGetter
}

// Crawler provides crawling capabilities
type Crawler struct {
}

// MergeCrawlStats merges two sets of crawling statistics together.
func MergeCrawlStats(statsA, statsB CrawlStats) (stats CrawlStats) {
	stats.StatusCodes = make(map[int]int)
	stats.Total = statsA.Total + statsB.Total

	if statsA.Max200Time > statsB.Max200Time {
		stats.Max200Time = statsA.Max200Time
	} else {
		stats.Max200Time = statsB.Max200Time
	}

	if statsA.StatusCodes != nil {
		for key, value := range statsA.StatusCodes {
			stats.StatusCodes[key] = stats.StatusCodes[key] + value
		}
	}
	if statsB.StatusCodes != nil {
		for key, value := range statsB.StatusCodes {
			stats.StatusCodes[key] = stats.StatusCodes[key] + value
		}
	}

	if statsA.Average200Time != 0 || statsB.Average200Time != 0 {
		total200ns := (statsA.Average200Time.Nanoseconds()*int64(statsA.StatusCodes[200]) +
			statsB.Average200Time.Nanoseconds()*int64(statsB.StatusCodes[200]))
		stats.Average200Time = time.Duration(total200ns/int64(stats.StatusCodes[200])) * time.Nanosecond
	}

	stats.Non200Urls = append(stats.Non200Urls, statsA.Non200Urls...)
	stats.Non200Urls = append(stats.Non200Urls, statsB.Non200Urls...)

	return
}

func addInterruptHandlers() chan struct{} {
	stop := make(chan struct{})
	osSignal := make(chan os.Signal)
	signal.Notify(osSignal, os.Interrupt, syscall.SIGTERM)
	signal.Notify(osSignal, os.Interrupt, syscall.SIGINT)

	go func() {
		<-osSignal
		log.Warn("Interrupt signal received")
		stop <- struct{}{}
	}()

	return stop
}

// GetSitemapUrls returns all URLs found from the sitemap passed as parameter.
// This function will only retrieve URLs in the sitemap pointed, and in
// sitemaps directly listed (i.e. only 1 level deep or less)
func GetSitemapUrls(sitemapURL string) (urls []*url.URL, err error) {
	sitemap, err := sitemap.Get(sitemapURL, nil)

	if err != nil {
		log.Error(err)
		return
	}

	for _, urlEntry := range sitemap.URL {
		newURL, err := url.Parse(urlEntry.Loc)
		if err != nil {
			log.Error(err)
			continue
		}
		urls = append(urls, newURL)
	}

	return
}

// GetSitemapUrlsAsStrings returns all URLs found as string, from in the
// sitemap passed as parameter.
// This function will only retrieve URLs in the sitemap pointed, and in
// sitemaps directly listed (i.e. only 1 level deep or less)
func GetSitemapUrlsAsStrings(sitemapURL string) (urls []string, err error) {
	typedUrls, err := GetSitemapUrls(sitemapURL)
	for _, url := range typedUrls {
		urls = append(urls, url.String())
	}

	return
}

// AsyncCrawl crawls asynchronously URLs from a sitemap and prints related
// information. Throttle is the maximum number of parallel HTTP requests.
// Host overrides the hostname used in the sitemap if provided,
// and user/pass are optional basic auth credentials
func AsyncCrawl(urls []string, config CrawlConfig) (stats CrawlStats,
	stopped bool, err error) {
	stats.StatusCodes = make(map[int]int)
	defer func() {
		if stats.Total == 0 {
			err = errors.New("No URL crawled")
		} else if stats.Total != stats.StatusCodes[200] {
			err = errors.New("Some URLs had a different status code than 200")
		}
	}()

	if config.Throttle <= 0 {
		log.Warn("Invalid throttle value, defaulting to 1.")
		config.Throttle = 1
	}

	var serverTimeSum time.Duration
	defer func() {
		total200 := stats.StatusCodes[200]
		if total200 > 0 {
			stats.Average200Time = serverTimeSum / time.Duration(total200)
		}
	}()

	quit := addInterruptHandlers()
	results := config.HTTPGetter.ConcurrentHTTPGet(urls, config.HTTP, config.Throttle, quit)
	for {
		select {
		case result, channelOpen := <-results:
			if !channelOpen {
				return
			}

			updateCrawlStats(result, &stats, &serverTimeSum)
		}
	}
}

func updateCrawlStats(result *HTTPResponse, stats *CrawlStats, totalTime *time.Duration) {
	stats.Total++
	if result.Err != nil {
		stats.StatusCodes[0]++
	} else {
		stats.StatusCodes[result.Response.StatusCode]++

		serverTime := result.Result.Total(result.EndTime)
		if result.Response.StatusCode == 200 {
			*totalTime += serverTime

			if serverTime > stats.Max200Time {
				stats.Max200Time = serverTime
			}
		} else {
			stats.Non200Urls = append(stats.Non200Urls, CrawlResult{
				URL:        result.URL,
				Time:       serverTime,
				StatusCode: result.Response.StatusCode,
			})
		}
	}
}
