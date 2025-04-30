package crawler_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"v2ex-tui/internal/crawler"
)

func TestCrawler(t *testing.T) {
	for _, proxy := range []string{
		"http://127.0.0.1:7890",
		"socks5://127.0.0.1:7891",
	} {
		t.Run(proxy, func(t *testing.T) {
			c := crawler.New(crawler.WithProxy(proxy))
			ret, err := c.FetchTopics()
			assert.NoError(t, err)
			assert.Greater(t, len(ret), 0)
		})
	}

}
