package main

import (
	"crypto/tls"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
	"github.com/sshst/sshststuff/tlslog"
	"github.com/sshst/sshststuff/wsconn"
	"io"
	"net/http"
	"os"
)

func connectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect",
		Short: "SSH ProxyCommand to connect to ssh.st-proxied services",
		Run:   connect,
	}

	cmd.PersistentFlags().String("sni", "", "")
	cmd.PersistentFlags().String("port", "443", "")

	return cmd
}

func connect(cmd *cobra.Command, args []string) {
	sni, err := cmd.PersistentFlags().GetString("sni")
	if err != nil || len(sni) == 0 {
		fmt.Fprintln(os.Stderr, "Must provide one --sni argument")
		os.Exit(1)
	}

	port, _ := cmd.PersistentFlags().GetString("port")
	url := fmt.Sprintf("wss://%s:%s/api/clients", sni, port)

	dialer := websocket.Dialer{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			ServerName:   sni,
			KeyLogWriter: tlslog.Writer,
		},
	}
	wscon, _, err := dialer.Dial(url, http.Header{})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error establishing TLS connection: %+v", err)
		os.Exit(1)
	}

	conn := wsconn.New(wscon)
	go io.Copy(conn, os.Stdin)
	io.Copy(os.Stdout, conn)
}
