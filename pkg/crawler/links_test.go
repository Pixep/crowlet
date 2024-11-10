package crawler

import (
	"net/url"
	"testing"
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
