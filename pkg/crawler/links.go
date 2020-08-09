package crawler

import (
	"io"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
)

// LinkType represent the type of link to crawl
type LinkType int

const (
	// Hyperlink is html 'a' tag
	Hyperlink LinkType = 0
	// Image is html 'img' tag
	Image LinkType = 1
)

// Link type holds information of URL links
type Link struct {
	Type       LinkType
	Name       string
	TargetURL  url.URL
	IsExternal bool
}

// ExtractLinks returns links found in the html page provided and currentURL.
// The URL is used to differentiate between internal and external links
func ExtractLinks(htmlBody io.ReadCloser, currentURL url.URL) ([]Link, error) {
	doc, err := goquery.NewDocumentFromReader(htmlBody)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	links := extractALinks(doc)
	links = append(links, extractImageLinks(doc)...)

	for index := range links {
		links[index].IsExternal = links[index].TargetURL.IsAbs() &&
			links[index].TargetURL.Host != currentURL.Host

		if !links[index].TargetURL.IsAbs() {
			links[index].TargetURL = *currentURL.ResolveReference(&links[index].TargetURL)
		}
	}
	return links, nil
}

func extractALinks(doc *goquery.Document) (links []Link) {
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		targetURL, _ := s.Attr("href")

		// TODO: Check that # exists
		if strings.HasPrefix(targetURL, "#") {
			return
		}

		link := extractLink(targetURL)
		if link == nil {
			return
		}

		link.Type = Hyperlink
		links = append(links, *link)
	})

	return
}

func extractImageLinks(doc *goquery.Document) (links []Link) {
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		targetURL, _ := s.Attr("src")

		if strings.HasPrefix(targetURL, "data:") {
			return
		}

		link := extractLink(targetURL)
		if link == nil {
			return
		}

		link.Type = Image
		links = append(links, *link)
	})

	return
}

func extractLink(urlString string) *Link {
	url, err := url.Parse(urlString)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &Link{
		Name:      "<a> link", //strings.TrimSpace(s.Text()),
		TargetURL: *url}
}
