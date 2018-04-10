package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/flaccid/sitemap-crawler/util"
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
			Name:  "async,a",
			Usage: "do http requests in asynchronous mode",
		},
		cli.IntFlag{
			Name:   "throttle,t",
			Usage:  "number of http requests to do at once in async mode",
			EnvVar: "CRAWL_THROTTLE",
			Value:  5,
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

	for {
		if c.Bool("async") {
			log.Info("async mode enabled")
			util.AsyncCrawl(smap, c.Int("throttle"), c.String("host"), c.String("user"), c.String("pass"))
		} else {
			util.SyncCrawl(smap, c.Int("throttle"), c.String("host"), c.String("user"), c.String("pass"))
		}
	}

	return nil
}
