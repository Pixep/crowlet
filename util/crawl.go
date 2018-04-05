package util

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/tcnksm/go-httpstat"
	"github.com/yterajima/go-sitemap"
)

func AsyncCrawl(smap sitemap.Sitemap, throttle int, host string, user string, pass string) {
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

		log.WithFields(log.Fields{
			"iteration": i,
			"low":       low,
			"high":      high,
		}).Debug("loop position")

		results := AsyncHttpGets(urls[low:high], user, pass)
		log.Debug("batch ", low, ":", high, " sending")
		for _ = range urls[low:high] {
			result := <-results
			log.Info(result.Url, result)
		}
		log.Debug("batch ", low, ":", high, " done")
		time.Sleep(1 * time.Second)
	}
}

func SyncCrawl(smap sitemap.Sitemap, throttle int, host string, user string, pass string) {
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
		res, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := io.Copy(ioutil.Discard, res.Body); err != nil {
			log.Fatal(err)
		}
		res.Body.Close()
		end := time.Now()

		log.WithFields(log.Fields{
			"resp":    res.StatusCode,
			"server":  int(result.ServerProcessing / time.Millisecond),
			"content": int(result.ContentTransfer(time.Now()) / time.Millisecond),
		}).Info("GET: " + u.String())

		log.WithFields(log.Fields{
			"resp":    res.StatusCode,
			"dns":     int(result.DNSLookup / time.Millisecond),
			"tcpconn": int(result.TCPConnection / time.Millisecond),
			"tls":     int(result.TLSHandshake / time.Millisecond),
			"server":  int(result.ServerProcessing / time.Millisecond),
			"content": int(result.ContentTransfer(time.Now()) / time.Millisecond),
			"close":   end,
		}).Debug("GET: " + u.String())
	}
}
