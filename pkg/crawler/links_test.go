package crawler

import (
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestRewriteURLHost(t *testing.T) {
	tests := []struct {
		name         string
		inputURLs    []string
		newHost      string
		expectedURLs []string
	}{
		{
			name:         "Valid URLs",
			inputURLs:    []string{"http://example.com/path", "https://another.com/otherpath"},
			newHost:      "newhost.com",
			expectedURLs: []string{"http://newhost.com/path", "https://newhost.com/otherpath"},
		},
		{
			name:         "Localhost",
			inputURLs:    []string{"https://example.com/path"},
			newHost:      "localhost",
			expectedURLs: []string{"https://localhost/path"},
		},
		{
			name:         "Invalid URL",
			inputURLs:    []string{"http://example.com/path", "://bad_url"},
			newHost:      "newhost.com",
			expectedURLs: []string{"http://newhost.com/path"}, // Only valid URL should be rewritten
		},
		{
			name:         "Empty Input",
			inputURLs:    []string{},
			newHost:      "newhost.com",
			expectedURLs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RewriteURLHost(tt.inputURLs, tt.newHost)
			if len(result) != len(tt.expectedURLs) {
				t.Errorf("expected %d URLs, got %d", len(tt.expectedURLs), len(result))
			}
			for i := range result {
				expectedURL, _ := url.Parse(tt.expectedURLs[i])
				resultURL, _ := url.Parse(result[i])
				if resultURL.Scheme != expectedURL.Scheme ||
					resultURL.Host != expectedURL.Host ||
					resultURL.Path != expectedURL.Path {
					t.Errorf("expected URL %s, got %s", tt.expectedURLs[i], result[i])
				}
			}
		})
	}
}

func Test_extractLink(t *testing.T) {
	type args struct {
		urlString string
	}
	tests := []struct {
		name string
		args args
		want *Link
	}{
		{
			name: "Valid URL",
			args: args{
				urlString: "http://example.com/path",
			},
			want: &Link{
				TargetURL: url.URL{Scheme: "http", Host: "example.com", Path: "/path"},
			},
		},
		{
			name: "Invalid URL",
			args: args{
				urlString: "://bad_url",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractLink(tt.args.urlString); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractLink() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_extractALinks(t *testing.T) {
	type args struct {
		doc *goquery.Document
	}
	tests := []struct {
		name      string
		args      args
		wantLinks []Link
	}{
		{
			name: "Valid URL",
			args: args{
				doc: func() *goquery.Document {
					doc, _ := goquery.NewDocumentFromReader(strings.NewReader("<a href='http://example.com/path'></a>"))
					return doc
				}(),
			},
			wantLinks: []Link{
				{
					TargetURL: url.URL{Scheme: "http", Host: "example.com", Path: "/path"},
					Type:      Hyperlink,
				},
			},
		},
		{
			name: "Valid URL with anchor",
			args: args{
				doc: func() *goquery.Document {
					doc, _ := goquery.NewDocumentFromReader(strings.NewReader("<a href='#anchor'></a>"))
					return doc
				}(),
			},
			wantLinks: nil,
		},
		{
			name: "Multiple URLs",
			args: args{
				doc: func() *goquery.Document {
					doc, _ := goquery.NewDocumentFromReader(strings.NewReader("<a href='http://example.com/path'></a><a href='http://another.com/otherpath'></a>"))
					return doc
				}(),
			},
			wantLinks: []Link{
				{
					TargetURL: url.URL{Scheme: "http", Host: "example.com", Path: "/path"},
					Type:      Hyperlink,
				},
				{
					TargetURL: url.URL{Scheme: "http", Host: "another.com", Path: "/otherpath"},
					Type:      Hyperlink,
				},
			},
		},
		{
			name: "Invalid URL",
			args: args{
				doc: func() *goquery.Document {
					doc, _ := goquery.NewDocumentFromReader(strings.NewReader("<a href='://bad_url'></a>"))
					return doc
				}(),
			},
			wantLinks: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotLinks := extractALinks(tt.args.doc); !reflect.DeepEqual(gotLinks, tt.wantLinks) {
				t.Errorf("extractALinks() = %v, want %v", gotLinks, tt.wantLinks)
			}
		})
	}
}

func Test_extractImageLinks(t *testing.T) {
	type args struct {
		doc *goquery.Document
	}
	tests := []struct {
		name      string
		args      args
		wantLinks []Link
	}{
		{
			name: "Invalid URL",
			args: args{
				doc: func() *goquery.Document {
					doc, _ := goquery.NewDocumentFromReader(strings.NewReader("<img src='://bad_url'></img>"))
					return doc
				}(),
			},
			wantLinks: nil,
		},
		{
			name: "Multiple image URLs",
			args: args{
				doc: func() *goquery.Document {
					doc, _ := goquery.NewDocumentFromReader(strings.NewReader("<img src='http://example.com/image'></img><img src='http://another.com/anotherimage'></img>"))
					return doc
				}(),
			},
			wantLinks: []Link{
				{
					TargetURL: url.URL{Scheme: "http", Host: "example.com", Path: "/image"},
					Type:      Image,
				},
				{
					TargetURL: url.URL{Scheme: "http", Host: "another.com", Path: "/anotherimage"},
					Type:      Image,
				},
			},
		},
		{
			name: "Valid URL with data",
			args: args{
				doc: func() *goquery.Document {
					doc, _ := goquery.NewDocumentFromReader(strings.NewReader("<img src='data:image/png;base64,base64'></img>"))
					return doc
				}(),
			},
			wantLinks: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotLinks := extractImageLinks(tt.args.doc); !reflect.DeepEqual(gotLinks, tt.wantLinks) {
				t.Errorf("extractImageLinks() = %v, want %v", gotLinks, tt.wantLinks)
			}
		})
	}
}
