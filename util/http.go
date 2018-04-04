package util

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
)

type HttpResponse struct {
	Url      string
	Response *http.Response
	Err      error
}

func AsyncHttpGets(urls []string, user string, pass string) <-chan *HttpResponse {
	ch := make(chan *HttpResponse, len(urls)) // buffered
	for _, url := range urls {
		go func(url string) {
			log.Info("Fetching " + url)

			req, err := http.NewRequest("GET", url, nil)
			if len(user) > 0 {
				req.SetBasicAuth(user, pass)
			}

			client := &http.Client{}

			resp, err := client.Do(req)
			if (err == nil) {
      	resp.Body.Close()
      }

			ch <- &HttpResponse{url, resp, err}
		}(url)
	}
	return ch
}
