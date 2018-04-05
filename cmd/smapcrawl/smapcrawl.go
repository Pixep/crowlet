package main

import (
	"net/url"
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
	log.Info(len(smap.URL), " urls to crawl")

	if err != nil {
		log.Fatal(err)
	}

	if c.Bool("async") {
		log.Debug("async mode enabled")
		util.AsyncCrawl(smap, c.Int("throttle"), c.String("host"), c.String("user"), c.String("pass"))
	} else {
		util.SyncCrawl(smap, c.Int("throttle"), c.String("host"), c.String("user"), c.String("pass"))
	}

	return nil
}
