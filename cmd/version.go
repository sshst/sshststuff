package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var commit = ""
var version = ""
var date = ""

func versionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version and build info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf(`
sshst
Version: %s
Commit:  %s
Date:    %s
`, version, commit, date)
		},
	}

	return cmd
}
