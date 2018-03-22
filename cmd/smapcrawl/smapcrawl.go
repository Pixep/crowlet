package main

import (
	"net/http"
	"net/url"
	"io"
	"io/ioutil"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/yterajima/go-sitemap"
	"github.com/tcnksm/go-httpstat"
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
			Usage: "username for basic authentication",
		},
		cli.StringFlag{
			Name:  "pass,p",
			Usage: "password for basic authentication",
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
			"resp": res.StatusCode,
			"server": int(result.ServerProcessing/time.Millisecond),
			"content": int(result.ContentTransfer(time.Now())/time.Millisecond),
		}).Info("GET: " + u.String())

		log.WithFields(log.Fields{
			"resp": res.StatusCode,
			"dns": int(result.DNSLookup/time.Millisecond),
			"tcpconn": int(result.TCPConnection/time.Millisecond),
			"tls": int(result.TLSHandshake/time.Millisecond),
			"server": int(result.ServerProcessing/time.Millisecond),
			"content": int(result.ContentTransfer(time.Now())/time.Millisecond),
			"close": end,
		}).Debug("GET: " + u.String())
	}

	return nil
}
