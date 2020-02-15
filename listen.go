package client

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"github.com/gliderlabs/ssh"
	"github.com/gorilla/websocket"
	"github.com/hashicorp/yamux"
	"github.com/pkg/errors"
	"github.com/soheilhy/cmux"
	"github.com/sshst/sshststuff/pb"
	"github.com/sshst/sshststuff/wsconn"
	gossh "golang.org/x/crypto/ssh"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"strings"
	"time"
)

var WebTerminalFingerprint = "SHA256:AgX6IK2m9OgKt54/33gZmCxrMLvtXjMWnJy7j38c2zI"

func Listen(ctx context.Context, config ListenConfig) error {
	trusted := config.SshFingerprints
	if config.WebOk {
		trusted = append(trusted, WebTerminalFingerprint)
	}

	headers := http.Header{}
	header, _ := json.Marshal(config)
	headers.Add("Sshst-Config", string(header))
	headers.Add("Sshst-Commit", config.Version)

	conn, err := connection(config.ApiUrl, headers)
	if err != nil {
		return err
	}

	sess, err := yamux.Server(conn, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	mux := cmux.New(sess)
	grpcL := mux.Match(cmux.HTTP2())
	sshL := mux.Match(cmux.Any())

	server := server(time.Duration(config.MaxTimeout)*time.Second, trusted)

	l := &listener{
		server:        server,
		command:       config.Command,
		activechanged: make(chan int),
	}

	control := grpc.NewServer()
	pb.RegisterListenerControlServer(control, &controlChannel{hk: l})
	go control.Serve(grpcL)

	server.Handler = l.handleSsh
	go server.Serve(sshL)

	go l.timeout(config.IdleTimeout)
	fmt.Println("Established connection to HQ.")

	errch := make(chan error)
	go func() {
		errch <- errors.WithStack(mux.Serve())
	}()

	select {
	case err := <-errch:
		return err
	case <-ctx.Done():
		_ = conn.Close()
		return ctx.Err()
	}
}

func server(timeout time.Duration, trustedFingerprints []string) ssh.Server {
	return ssh.Server{
		HostSigners: []ssh.Signer{newSigner()},
		MaxTimeout:  timeout,
		PublicKeyHandler: func(ctx ssh.Context, key ssh.PublicKey) bool {
			fingerprint := effectiveFingerprint(key)
			for _, trusted := range trustedFingerprints {
				if trusted == fingerprint {
					return true
				}
			}

			return false
		},
		LocalPortForwardingCallback: func(ctx ssh.Context, destinationHost string, destinationPort uint32) bool {
			return true
		},
		ReversePortForwardingCallback: func(ctx ssh.Context, bindHost string, bindPort uint32) bool {
			return true
		},
	}
}

func connection(url string, headers http.Header) (net.Conn, error) {
	if strings.HasPrefix(url, "https") {
		url = strings.Replace(url, "https", "wss", 1)
	}

	wsc, resp, err := websocket.DefaultDialer.Dial(url, headers)
	if err != nil {
		if resp != nil {
			if resp.StatusCode == 307 || resp.StatusCode == 308 {
				url = resp.Header.Get("Location")
				return connection(url, headers)
			}
		}

		return nil, errors.WithStack(err)
	}

	c := wsconn.New(wsc)
	return c, nil
}

func fingerprinter(s ssh.PublicKey) string {
	fingerprint := gossh.FingerprintSHA256(s)

	if cert, ok := s.(*gossh.Certificate); ok {
		if gossh.FingerprintSHA256(cert.SignatureKey) == WebTerminalFingerprint {
			fingerprint = fmt.Sprintf("GitHub user %s", cert.KeyId)
		} else {
			fingerprint = fmt.Sprintf("%s (%s)", cert.KeyId, fingerprint)
		}
	}
	return fingerprint
}

func newSigner() ssh.Signer {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	signer, _ := gossh.NewSignerFromKey(privateKey)
	return signer
}

func effectiveFingerprint(key ssh.PublicKey) string {
	if cert, ok := key.(*gossh.Certificate); ok {
		return gossh.FingerprintSHA256(cert.SignatureKey)
	}
	return gossh.FingerprintSHA256(key)
}

