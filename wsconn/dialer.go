package wsconn

import (
	"context"
	"crypto/tls"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/sshst/sshststuff/tlslog"
	"net"
	"net/http"
	"strings"
	"time"
)

func DialContext(ctx context.Context, url string, headers http.Header) (net.Conn, *http.Response, error) {
	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 10 * time.Second,
		TLSClientConfig: &tls.Config{
			KeyLogWriter: tlslog.Writer,
		},
	}

redirect:
	if strings.HasPrefix(url, "http") {
		url = strings.Replace(url, "http", "ws", 1)
	}

	wsc, resp, err := dialer.DialContext(ctx, url, headers)
	if err != nil {
		if resp != nil {
			if resp.StatusCode == 307 || resp.StatusCode == 308 {
				url = resp.Header.Get("Location")
				goto redirect
			}
		}

		return nil, resp, errors.WithStack(err)
	}

	return NetConn(wsc), resp, nil
}
