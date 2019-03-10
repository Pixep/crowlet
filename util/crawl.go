package util

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/tcnksm/go-httpstat"
	"github.com/yterajima/go-sitemap"
)

type CrawlStats struct {
	Resp200    int
	RespNon200 int
}

func logTotals(stats CrawlStats) {
	log.Info("total 200 responses: ", stats.Resp200)
	log.Info("total non-200 responses: ", stats.RespNon200)
}

func AsyncCrawl(smap sitemap.Sitemap, throttle int, host string, user string, pass string) {
	var stats CrawlStats

	// support ctrl-c
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		log.Info("sigterm triggered")
		logTotals(stats)
		os.Exit(1)
	}()

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
	numIter := numUrls / throttle

	log.WithFields(log.Fields{
		"url count":  numUrls,
		"throttle":   throttle,
		"iterations": numIter,
	}).Debug("loop summary")

	var low int
	for i := 0; i <= numIter; i++ {
		if i == 0 {
			low = 0
		} else {
			low = i * throttle
		}
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

		results := AsyncHttpGets(urls[low:high], user, pass)
		log.Debug("batch ", low, ":", high, " sending")
		for _ = range urls[low:high] {
			result := <-results

			// look at removal once true async is done
			//log.Debug(result.Url, result)

			// stats collection
			if result.Response.StatusCode == 200 {
				stats.Resp200++
			}
			if result.Response.StatusCode != 200 {
				stats.RespNon200++
			}
		}
		log.Debug("batch ", low, ":", high, " done")
		log.Debug("sleep 1")
		time.Sleep(1 * time.Second)
	}
}

func SyncCrawl(smap sitemap.Sitemap, throttle int, host string, user string, pass string) {
	var stats CrawlStats

	// support ctrl-c
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		log.Info("sigterm triggered")
		logTotals(stats)
		os.Exit(1)
	}()

	// each in sitemap
	for _, URL := range smap.URL {
		u, err := url.Parse(URL.Loc)
		if err != nil {
			panic(err)
		}

		if len(host) > 0 {
			u.Host = host
		}

		// create a new http request
		req, err := http.NewRequest("GET", u.String(), nil)
		if err != nil {
			log.Fatal(err)
		}

		// create a httpstat powered context
		var result httpstat.Result
		ctx := httpstat.WithHTTPStat(req.Context(), &result)
		req = req.WithContext(ctx)

		// add basic auth if user is provided
		if len(user) > 0 {
			req.SetBasicAuth(user, pass)
		}

		// send request by default http client
		client := http.DefaultClient
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := io.Copy(ioutil.Discard, resp.Body); err != nil {
			log.Fatal(err)
		}
		resp.Body.Close()
		end := time.Now()

		// stats collection
		if resp.StatusCode == 200 {
			stats.Resp200++
		}
		if resp.StatusCode != 200 {
			stats.RespNon200++
		}

		// logging
		if log.GetLevel() == log.DebugLevel {
			log.WithFields(log.Fields{
				"resp":    resp.StatusCode,
				"dns":     int(result.DNSLookup / time.Millisecond),
				"tcpconn": int(result.TCPConnection / time.Millisecond),
				"tls":     int(result.TLSHandshake / time.Millisecond),
				"server":  int(result.ServerProcessing / time.Millisecond),
				"content": int(result.ContentTransfer(time.Now()) / time.Millisecond),
				"close":   end,
			}).Debug("GET: " + u.String())
		} else {
			log.WithFields(log.Fields{
				"resp":    resp.StatusCode,
				"server":  int(result.ServerProcessing / time.Millisecond),
				"content": int(result.ContentTransfer(time.Now()) / time.Millisecond),
			}).Info("GET: " + u.String())
		}
	}
}
