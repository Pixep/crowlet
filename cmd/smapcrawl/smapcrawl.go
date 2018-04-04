package main

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/flaccid/sitemap-crawler/util"
	"github.com/tcnksm/go-httpstat"
	"github.com/urfave/cli"
	"github.com/yterajima/go-sitemap"
)

var (
	VERSION = "v0.0.0-dev"
)

func beforeApp(c *cli.Context) error {
	if c.GlobalBool("debug") {
		log.SetLevel(log.DebugLevel)
	}

	if c.NArg() < 1 {
		log.Fatal("sitemap url required")
	}

	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "smapcrawl"
	app.Version = VERSION
	app.Usage = "a basic sitemap.xml crawler"
	app.Action = start
	app.Before = beforeApp
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "host",
			Usage: "override the hostname used in sitemap urls",
		},
		cli.StringFlag{
			Name:  "user,u",
			Usage: "username for http basic authentication",
		},
		cli.StringFlag{
			Name:  "pass,p",
			Usage: "password for http basic authentication",
		},
		cli.BoolFlag{
			Name:  "async,a",
			Usage: "do http requests in asynchronous mode",
		},
		cli.IntFlag{
			Name:  "throttle,t",
			Usage: "number of http requests to do at once",
			Value: 5,
		},
		cli.BoolFlag{
			Name:  "debug,d",
			Usage: "run in debug mode",
		},
	}
	app.Run(os.Args)
}

func start(c *cli.Context) error {
	log.Info("starting up")

	smap, err := sitemap.Get(c.Args().Get(0), nil)
	if err != nil {
		log.Fatal(err)
	}

	if c.Bool("async") {
		log.Debug("async mode enabled")

		// place all the urls into an array
		var urls []string
		for _, URL := range smap.URL {
			u, err := url.Parse(URL.Loc)
			if err != nil {
				panic(err)
			}
			if len(c.String("host")) > 0 {
				u.Host = c.String("host")
			}
			urls = append(urls, u.String())
		}

		throttle := c.Int("throttle")
		numUrls := len(urls)
		numIter := numUrls / throttle

		log.WithFields(log.Fields{
			"url count":  numUrls,
			"throttle":   throttle,
			"iterations": numIter,
		}).Debug("loop summary")

		/// LOOP START
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

			results := util.AsyncHttpGets(urls[low:high], c.String("user"), c.String("pass"))
			log.Debug("batch ", low, ":", high, " sending")
			for _ = range urls[low:high] {
				result := <-results
				log.Info(result.Url, result)
			}
			log.Debug("batch ", low, ":", high, " done")
			time.Sleep(1 * time.Second)
		}
		/// LOOP END
	} else {
		log.Info(len(smap.URL), " urls")

		// each in sitemap
		for _, URL := range smap.URL {
			u, err := url.Parse(URL.Loc)
			if err != nil {
				panic(err)
			}

			if len(c.String("host")) > 0 {
				u.Host = c.String("host")
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
			if len(c.String("user")) > 0 {
				req.SetBasicAuth(c.String("user"), c.String("pass"))
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

	return nil
}
