package util

import (
	"errors"
	"math"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/yterajima/go-sitemap"
)

// CrawlStats stores crawling status codes and
// total number of crawled URLs
type CrawlStats struct {
	Total       int
	StatusCodes map[int]int
}

// PrintSummary prints a summary of HTTP response codes
func PrintSummary(stats CrawlStats) {
	log.Info("---------------")
	log.Info("Summary:")
	log.Info("  HTTP Status    Count")
	for code, count := range stats.StatusCodes {
		log.Info("    ", code, "          ", count)
	}
	log.Info("  Total          ", stats.Total)
	log.Info("---------------")
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

// AsyncCrawl crawls synchronously URLs from a sitemap and prints related
// information. Throttle is the maximum number of parallel HTTP requests.
// Host overrides the hostname used in the sitemap if provided,
// and user/pass are optional basic auth credentials
func AsyncCrawl(smap sitemap.Sitemap, throttle int, host string,
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

	// place all the urls into an array
	var urls []string
	for _, URL := range smap.URL {
		u, err := url.Parse(URL.Loc)
		if err != nil {
			panic(err)
		}
		if len(host) > 0 {
			u.Host = host
		}
		urls = append(urls, u.String())
	}

	numUrls := len(urls)
	numIter := int(math.Ceil(float64(numUrls) / float64(throttle)))

	log.WithFields(log.Fields{
		"url count":  numUrls,
		"throttle":   throttle,
		"iterations": numIter,
	}).Debug("loop summary")

	var low int
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
		log.Debug("batch ", low, ":", high, " sending")
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
		}
		log.Debug("batch ", low, ":", high, " done")
	}

	return
}
