package crawler

import (
	"golang.org/x/net/context"
	"golang.org/x/net/proxy"
	"net"
	"net/http"
)

func createHttpClient(proxyAddr string) (*http.Client, error) {
	if proxyAddr == "" {
		return http.DefaultClient, nil
	}

	// 创建 socks5 代理拨号器
	dialer, err := proxy.SOCKS5("tcp", proxyAddr,
		nil, // 认证信息，如果不需要认证则为nil
		proxy.Direct,
	)
	if err != nil {
		return nil, err
	}

	// 创建自定义 Transport
	httpTransport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		},
	}

	// 返回配置了代理的 HTTP 客户端
	return &http.Client{
		Transport: httpTransport,
	}, nil
}
