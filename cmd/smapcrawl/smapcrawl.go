package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

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
		fmt.Println(err)
	}

	for _, URL := range smap.URL {
		u, err := url.Parse(URL.Loc)
		if err != nil {
			panic(err)
		}

		if len(c.String("host")) > 0 {
			u.Host = c.String("host")
		}

		log.Info("GET ", u.String())

		body := strings.NewReader(``)
		req, err := http.NewRequest("GET", u.String(), body)
		if err != nil {
			fmt.Println(err)
		}
		if len(c.String("user")) > 0 {
			req.SetBasicAuth(c.String("user"), c.String("pass"))
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()

	}

	return nil
}
