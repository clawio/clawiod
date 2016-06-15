package helpers

import (
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeURL(t *testing.T) {
	u := "http://example.com/?access_token=somesecrettoken"
	testURL, err := url.Parse(u)
	assert.Nil(t, err)
	got := SanitizeURL(testURL)
	assert.True(t, strings.Contains(got, "REDACTED"))
}

func TestSanitizeURL_withNil(t *testing.T) {
	got := SanitizeURL(nil)
	assert.Empty(t, got)
}

func TestSanitizeURL_withoutToken(t *testing.T) {
	u := "http://example.com/"
	testURL, err := url.Parse(u)
	assert.Nil(t, err)
	got := SanitizeURL(testURL)
	assert.Equal(t, testURL.String(), got)
}
