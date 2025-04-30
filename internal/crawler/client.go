package crawler

import (
	"context"
	"net"
	"net/http"
	"net/url"

	"golang.org/x/net/proxy"
)

type httpOption func(*http.Client)

// set http proxy
func SetProxy(proxyUrl string) (httpOption, error) {
	proxyURL, err := url.Parse(proxyUrl)
	if err != nil {
		return nil, err
	}
	tr := &http.Transport{}
	switch proxyURL.Scheme {
	case "http", "https":
		tr.Proxy = http.ProxyURL(proxyURL)
		return func(client *http.Client) {
			client.Transport = tr
		}, nil
	case "socks5":
		auth := &proxy.Auth{}
		if proxyURL.User != nil {
			auth.User = proxyURL.User.Username()
			password, ok := proxyURL.User.Password()
			if ok {
				auth.Password = password
			}
		}
		dialer, err := proxy.SOCKS5("tcp", proxyURL.Host, auth, proxy.Direct)
		if err != nil {
			return nil, err
		}
		tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		}
	default:
		panic("unsupported proxy scheme")
	}
	return func(client *http.Client) {
		client.Transport = tr
	}, nil
}

func NewHttpClient(opts ...httpOption) *http.Client {
	baseClient := &http.Client{}
	for _, opt := range opts {
		opt(baseClient)
	}
	return baseClient
}
