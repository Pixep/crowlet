package util

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

// CrawlStats holds crawling related information: status codes, time
// and totals
type CrawlStats struct {
	Total          int
	StatusCodes    map[int]int
	Average200Time time.Duration
	Max200Time     time.Duration
}

// CrawlConfig holds crawling configuration.
type CrawlConfig struct {
	Throttle int
	Host     string
	HTTP     HTTPConfig
}

// MergeCrawlStats merges two sets of crawling statistics together.
// The average time will be an average of the two averages, and not an average
// of all individual times.
func MergeCrawlStats(statsA, statsB CrawlStats) (stats CrawlStats) {
	stats.Total = statsA.Total + statsB.Total
	if statsA.Total == 0 && statsB.Total != 0 {
		stats.Average200Time = statsB.Average200Time
	} else if statsB.Total == 0 {
		stats.Average200Time = statsA.Average200Time
	} else {
		// TODO: This is actually *not* an average anymore...
		stats.Average200Time = (statsA.Average200Time + statsB.Average200Time) / 2
	}

	if statsA.Max200Time > statsB.Max200Time {
		stats.Max200Time = statsA.Max200Time
	} else {
		stats.Max200Time = statsB.Max200Time
	}

	if statsA.StatusCodes != nil {
		stats.StatusCodes = statsA.StatusCodes
	} else {
		stats.StatusCodes = make(map[int]int)
	}

	if statsB.StatusCodes != nil {
		for key, value := range statsB.StatusCodes {
			stats.StatusCodes[key] = stats.StatusCodes[key] + value
		}
	}

	return
}

// PrintSummary prints a summary of HTTP response codes
func PrintSummary(stats CrawlStats) {
	log.Info("-------- Summary -------")
	log.Info("general:")
	log.Info("    crawled: ", stats.Total)
	log.Info("")
	log.Info("status:")
	for code, count := range stats.StatusCodes {
		log.Info("    status-", code, ": ", count)
	}
	log.Info("")
	log.Info("server-time: ")
	log.Info("    avg-time: ", int(stats.Average200Time/time.Millisecond), "ms")
	log.Info("    max-time: ", int(stats.Max200Time/time.Millisecond), "ms")
	log.Info("------------------------")
}

func addInterruptHandlers() chan struct{} {
	stop := make(chan struct{})
	osSignal := make(chan os.Signal)
	signal.Notify(osSignal, os.Interrupt, syscall.SIGTERM)
	signal.Notify(osSignal, os.Interrupt, syscall.SIGINT)

	go func() {
		<-osSignal
		log.Info("Interrupt signal received")
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
	results := ConcurrentHTTPGets(urls, config.HTTP, config.Throttle, quit)
	for {
		select {
		case result, channelOpen := <-results:
			if !channelOpen {
				return
			}

			stats.Total++
			if result.Err != nil {
				stats.StatusCodes[0]++
			} else {
				stats.StatusCodes[result.Response.StatusCode]++

				if result.Response.StatusCode == 200 {
					serverTime := result.Result.Total(result.EndTime)
					serverTimeSum += serverTime

					if serverTime > stats.Max200Time {
						stats.Max200Time = serverTime
					}
				}
			}
		}
	}
}
