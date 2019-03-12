package main

import (
	"os"
	"time"

	"github.com/Pixep/sitemap-crawler/util"
	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/yterajima/go-sitemap"
)

var (
	VERSION = "v0.0.0-dev"
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

func main() {
	app := cli.NewApp()
	app.Name = "smapcrawl"
	app.Version = VERSION
	app.Usage = "a basic sitemap.xml crawler"
	app.Action = start
	app.Before = beforeApp
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
			Usage: "reads the sitemap once keep crawling all urls until stopped",
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
			Usage: "suppresses all normal output",
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
}

func start(c *cli.Context) error {
	log.Info("starting up")

	smap, err := sitemap.Get(c.Args().Get(0), nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Info(len(smap.URL), " urls to crawl")

	var stats util.CrawlStats
	for i := 0; i < 1 || c.Bool("forever"); i++ {
		if i != 0 {
			time.Sleep(time.Duration(c.Int("wait-interval")) * time.Second)
		}

		var stop bool
		var err error
		stats, stop, err = util.AsyncCrawl(smap, c.Int("throttle"), c.String("host"), c.String("user"), c.String("pass"))

		util.PrintSummary(stats)

		if err != nil {
			log.Warn(err)
		}

		if stop {
			break
		}
	}

	return nil
}
