package util

import (
	"errors"
	"math"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/yterajima/go-sitemap"
)

// CrawlStats stores crawling status codes and
// total number of crawled URLs
type CrawlStats struct {
	Total          int
	StatusCodes    map[int]int
	Average200Time time.Duration
	Max200Time     time.Duration
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

func addInterruptHandlers(stop chan struct{}) {
	osSignal := make(chan os.Signal)
	signal.Notify(osSignal, os.Interrupt, syscall.SIGTERM)
	signal.Notify(osSignal, os.Interrupt, syscall.SIGINT)

	go func() {
		<-osSignal
		log.Info("Signal received, stopping...")
		stop <- struct{}{}
	}()
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
func AsyncCrawl(urls []string, throttle int, host string,
	user string, pass string) (stats CrawlStats, stopped bool, err error) {
	stats.StatusCodes = make(map[int]int)
	defer func() {
		if stats.Total == 0 {
			err = errors.New("No URL crawled")
		} else if stats.Total != stats.StatusCodes[200] {
			err = errors.New("Some URLs had a different status code than 200")
		}
	}()

	if throttle <= 0 {
		log.Warn("Invalid throttle value, defaulting to 1.")
		throttle = 1
	}

	stop := make(chan struct{})
	addInterruptHandlers(stop)

	numUrls := len(urls)
	numIter := int(math.Ceil(float64(numUrls) / float64(throttle)))

	log.WithFields(log.Fields{
		"url count":  numUrls,
		"throttle":   throttle,
		"iterations": numIter,
	}).Debug("loop summary")

	var low int
	var serverTimeSum time.Duration
	defer func() {
		total200 := stats.StatusCodes[200]
		if total200 > 0 {
			stats.Average200Time = serverTimeSum / time.Duration(total200)
		}
	}()
	for i := 0; i < numIter; i++ {
		low = i * throttle
		high := (low + throttle) - 1

		// do not let high exceed total (last batch/upper limit)
		if high > numUrls {
			high = numUrls - 1
		}

		log.WithFields(log.Fields{
			"iteration": i,
			"low":       low,
			"high":      high,
		}).Debug("loop position")

		urlRange := urls[low : high+1]
		results := AsyncHttpGets(urlRange, user, pass)
		for range urlRange {
			var result *HttpResponse
			select {
			case result = <-results:
			case <-stop:
				stopped = true
				return
			}

			stats.Total++
			stats.StatusCodes[result.Response.StatusCode]++

			if result.Response.StatusCode == 200 {
				serverTime := result.Result.Total(result.EndTime)
				serverTimeSum += serverTime

				if serverTime > stats.Max200Time {
					stats.Max200Time = serverTime
				}
			}
		}
		log.Debug("batch ", low, ":", high, " done")
	}

	return
}
