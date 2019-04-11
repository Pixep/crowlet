package main

import (
	"os"
	"time"

	exec "github.com/Pixep/crowlet/internal/pkg"
	"github.com/Pixep/crowlet/pkg/crawler"

	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	// VERSION stores the current version as string
	VERSION = "v0.1.0"
)

func beforeApp(c *cli.Context) error {
	if c.GlobalBool("debug") {
		log.SetLevel(log.DebugLevel)
	} else if c.GlobalBool("quiet") || c.GlobalBool("summary-only") {
		log.SetLevel(log.FatalLevel)
	}

	if c.GlobalBool("json") {
		log.SetFormatter(&log.JSONFormatter{})
	}

	if c.NArg() < 1 {
		log.Fatal("sitemap url required")
	}

	if len(c.GlobalString("pre-cmd")) > 1 {
		ok := exec.Exec(c.GlobalString("pre-cmd"))
		if !ok {
			log.Fatal("Failed to execute command")
		}
	}

	return nil
}

func afterApp(c *cli.Context) error {
	if len(c.GlobalString("post-cmd")) > 1 {
		ok := exec.Exec(c.GlobalString("post-cmd"))
		if !ok {
			log.Fatal("Failed to execute command")
		}
	}

	return nil
}

var exitCode int

func main() {
	app := cli.NewApp()
	app.Name = "crowlet"
	app.Version = VERSION
	app.Usage = "a basic sitemap.xml crawler"
	app.Action = start
	app.Before = beforeApp
	app.After = afterApp
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "host",
			Usage:  "override the hostname used in sitemap urls",
			EnvVar: "CRAWL_HOST",
		},
		cli.StringFlag{
			Name:   "user,u",
			Usage:  "username for http basic authentication",
			EnvVar: "CRAWL_HTTP_USER",
		},
		cli.StringFlag{
			Name:   "pass,p",
			Usage:  "password for http basic authentication",
			EnvVar: "CRAWL_HTTP_PASSWORD",
		},
		cli.BoolFlag{
			Name:  "forever,f",
			Usage: "crawl the sitemap's URLs forever... or until stopped",
		},
		cli.IntFlag{
			Name:  "iterations,i",
			Usage: "number of crawling iterations for the whole sitemap",
			Value: 1,
		},
		cli.IntFlag{
			Name:   "wait-interval,w",
			Usage:  "wait interval in seconds between sitemap crawling iterations",
			EnvVar: "CRAWL_WAIT_INTERVAL",
			Value:  0,
		},
		cli.IntFlag{
			Name:   "throttle,t",
			Usage:  "number of http requests to do at once",
			EnvVar: "CRAWL_THROTTLE",
			Value:  5,
		},
		cli.IntFlag{
			Name:  "timeout,y",
			Usage: "timeout duration for requests, in milliseconds",
			Value: 20000,
		},
		cli.BoolFlag{
			Name:  "quiet,silent,q",
			Usage: "suppress all normal output",
		},
		cli.BoolFlag{
			Name:  "json,j",
			Usage: "output using JSON format (experimental)",
		},
		cli.IntFlag{
			Name: "non-200-error,e",
			Usage: "error code to use if any non-200 response if" +
				" encountered",
			Value: 1,
		},
		cli.IntFlag{
			Name: "response-time-error,l",
			Usage: "error code to use if the maximum response time" +
				" is overrun",
			Value: 1,
		},
		cli.IntFlag{
			Name: "response-time-max,m",
			Usage: "maximum response time of URLs, in milliseconds, before" +
				" considered an error",
			Value: 0,
		},
		cli.BoolFlag{
			Name:  "summary-only",
			Usage: "print only the summary",
		},
		cli.StringFlag{
			Name:  "pre-cmd",
			Usage: "command(s) to run before starting crawler",
		},
		cli.StringFlag{
			Name:  "post-cmd",
			Usage: "command(s) to run after crawler finishes",
		},
		cli.BoolFlag{
			Name:  "debug",
			Usage: "run in debug mode",
		},
	}

	app.Run(os.Args)
	os.Exit(exitCode)
}

func start(c *cli.Context) error {
	sitemapURL := c.Args().Get(0)
	log.Info("Crawling ", sitemapURL)

	urls, err := crawler.GetSitemapUrlsAsStrings(sitemapURL)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Found ", len(urls), " URL(s)")

	config := crawler.CrawlConfig{
		Throttle: c.Int("throttle"),
		Host:     c.String("host"),
		HTTP: crawler.HTTPConfig{
			User:    c.String("user"),
			Pass:    c.String("pass"),
			Timeout: time.Duration(c.Int("timeout")) * time.Millisecond,
		},
		HTTPGetter: &crawler.BaseConcurrentHTTPGetter{
			Get: crawler.HTTPGet,
		},
	}

	var stats crawler.CrawlStats
	for i := 0; i < c.Int("iterations") || c.Bool("forever"); i++ {
		if i != 0 {
			time.Sleep(time.Duration(c.Int("wait-interval")) * time.Second)
		}

		itStats, stop, err := crawler.AsyncCrawl(urls, config)

		stats = crawler.MergeCrawlStats(stats, itStats)

		if err != nil {
			log.Warn(err)
		}

		if stop {
			break
		}
	}

	if !c.GlobalBool("quiet") {
		if c.GlobalBool("json") {
			crawler.PrintJSONSummary(stats)
		} else {
			crawler.PrintSummary(stats)
		}

		if c.Bool("summary-only") {
			log.SetLevel(log.FatalLevel)
		}
	}

	if stats.Total != stats.StatusCodes[200] {
		exitCode = c.Int("non-200-error")
		return nil
	}

	maxResponseTime := c.Int("response-time-max")
	if maxResponseTime > 0 && int(stats.Max200Time/time.Millisecond) > maxResponseTime {
		log.Warn("Max response time (", maxResponseTime, "ms) was exceeded")
		exitCode = c.Int("response-time-error")
	}

	return nil
}
