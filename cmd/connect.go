package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/sshst/sshststuff/wsconn"
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

	header := http.Header{}
	header.Set("Sshst-Commit", commit)

	conn, _, err := wsconn.DialContext(context.Background(), url, header)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error establishing TLS connection: %+v", err)
		os.Exit(1)
	}

	go io.Copy(conn, os.Stdin)
	io.Copy(os.Stdout, conn)
}
