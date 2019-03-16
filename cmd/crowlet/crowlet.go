package main

import (
	"os"
	"time"

	"github.com/Pixep/crowlet/util"
	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	VERSION = "v0.0.1"
)

func beforeApp(c *cli.Context) error {
	if c.GlobalBool("debug") {
		log.SetLevel(log.DebugLevel)
	} else if c.GlobalBool("quiet") {
		log.SetLevel(log.FatalLevel)
	}

	if c.NArg() < 1 {
		log.Fatal("sitemap url required")
	}

	if len(c.GlobalString("pre-cmd")) > 1 {
		util.Exec(c.GlobalString("pre-cmd"), "pre")
	}

	return nil
}

func afterApp(c *cli.Context) error {
	if len(c.GlobalString("post-cmd")) > 1 {
		util.Exec(c.GlobalString("post-cmd"), "post")
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
		cli.BoolFlag{
			Name:  "quiet,silent,q",
			Usage: "suppress all normal output",
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
		cli.StringFlag{
			Name:  "pre-cmd",
			Usage: "command(s) to run before starting crawler",
		},
		cli.StringFlag{
			Name:  "post-cmd",
			Usage: "command(s) to run after crawler finishes",
		},
		cli.BoolFlag{
			Name:  "debug,d",
			Usage: "run in debug mode",
		},
	}

	app.Run(os.Args)
	os.Exit(exitCode)
}

func start(c *cli.Context) error {
	sitemapURL := c.Args().Get(0)
	log.Info("Crawling ", sitemapURL)

	urls, err := util.GetSitemapUrlsAsStrings(sitemapURL)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Found ", len(urls), " URL(s)")

	var stats util.CrawlStats
	for i := 0; i < 1 || c.Bool("forever"); i++ {
		if i != 0 {
			time.Sleep(time.Duration(c.Int("wait-interval")) * time.Second)
		}

		itStats, stop, err := util.AsyncCrawl(urls, c.Int("throttle"),
			c.String("host"), c.String("user"), c.String("pass"))

		stats = util.MergeCrawlStats(stats, itStats)

		if err != nil {
			log.Warn(err)
		}

		if stop {
			break
		}
	}

	util.PrintSummary(stats)

	if stats.Total != stats.StatusCodes[200] {
		exitCode = c.Int("non-200-error")
	}

	maxResponseTime := c.Int("response-time-max")
	if maxResponseTime > 0 && int(stats.Max200Time/time.Millisecond) > maxResponseTime {
		log.Warn("Max response time (", maxResponseTime, "ms) was exceeded")
		exitCode = c.Int("response-time-error")
	}

	return nil
}