package main

import (
	"fmt"
	"github.com/spf13/cobra"
	client "github.com/sshst/sshststuff"
)

func setupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Run this once to configure your ~/.ssh/config",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf(`
### To run this, you can type "sshst setup | sh" in your terminal.

set -eux

mkdir -p ~/.ssh

cat << EOF >> ~/.ssh/config
Host *.ssh.st
  ProxyCommand sshst connect --sni %%h
  UserKnownHostsFile ~/.ssh/sshst_known_hosts
EOF

cat << EOF > ~/.ssh/sshst_known_hosts
@cert-authority * %s 
EOF
`, client.SshstPubKey)
		},
	}

	return cmd
}
