package wsconn

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/sshst/sshststuff/tlslog"
)

// DialContext wraps websocket.DefaultDialer with a KeyLogWriter that respects
// the SSLKEYLOGFILE environment variable and follows HTTP redirects (up to a
// limit).
func DialContext(ctx context.Context, url string, headers http.Header) (net.Conn, error) {
	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 10 * time.Second,
		TLSClientConfig: &tls.Config{
			KeyLogWriter: tlslog.Writer,
		},
	}

	limit := 10
redirect:
	if strings.HasPrefix(url, "http") {
		url = strings.Replace(url, "http", "ws", 1)
	}

	wsc, resp, err := dialer.DialContext(ctx, url, headers)
	if err != nil {
		if resp != nil {
			if limit > 0 && (resp.StatusCode == 307 || resp.StatusCode == 308) {
				limit--

				url = resp.Header.Get("Location")

				goto redirect
			}
		}

		return nil, errors.WithStack(err)
	}

	return NetConn(wsc), nil
}
