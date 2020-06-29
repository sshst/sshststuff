package client

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/hashicorp/yamux"
	"github.com/pkg/errors"
	"github.com/soheilhy/cmux"
	"github.com/sshst/sshststuff/pb"
	"github.com/sshst/sshststuff/wsconn"
	gossh "golang.org/x/crypto/ssh"
	"google.golang.org/grpc"
)

var WebTerminalFingerprint = "SHA256:AgX6IK2m9OgKt54/33gZmCxrMLvtXjMWnJy7j38c2zI"
var SshstPubKey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDZFVovbC6jV44UTyFhEfg861v4ddetMlC+IIRB3cPMotlCZg5MhL2sD3Y2FQVaSVtQfVQ4+wyyJxhs3QzH2Qhhmm7thhs+D7uMJHMb+2jvCT3JdYLA/7II6oS+wmh5i5dMBMJtWO/h81pbZt8LzadEGFzEbIsUU9s12m/Tq26g0petIPt5gzgiXjQ3wjN5lzt6NK5/iUVPxjgLxV6BY1coDaC5ZQvW02naU2v7V01kL4MVHfIpy/9sMSt/2zFxg7lstSy9GQqwn4o/nHq/5yhlskjY/tpkBbAqRxp9hx9R+Ci3J5jDvks5l/eyUTyJVNa1ASnnqV2dWWGqB32FqIPn"

func Listen(ctx context.Context, config ListenConfig) error {
	trusted := config.SSHFingerprints
	if config.WebOk {
		trusted = append(trusted, WebTerminalFingerprint)
	}

	headers := http.Header{}
	header, _ := json.Marshal(config)
	headers.Add("Sshst-Config", string(header))
	headers.Add("Sshst-Commit", config.Version)

	conn, err := wsconn.DialContext(ctx, config.APIURL, headers)
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

	server.Handler = l.handleSSH
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

func server(timeout time.Duration, trustedFingerprints []string) *ssh.Server {
	forwardHandler := &ssh.ForwardedTCPHandler{}

	return &ssh.Server{
		HostSigners: []ssh.Signer{newSigner()},
		MaxTimeout:  timeout,
		ChannelHandlers: map[string]ssh.ChannelHandler{
			"session":      ssh.DefaultSessionHandler,
			"direct-tcpip": ssh.DirectTCPIPHandler,
		},
		RequestHandlers: map[string]ssh.RequestHandler{
			"tcpip-forward":        forwardHandler.HandleSSHRequest,
			"cancel-tcpip-forward": forwardHandler.HandleSSHRequest,
		},
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
