package main

import (
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{Use: "sshst"}
	root.AddCommand(listenCmd())
	root.AddCommand(connectCmd())
	root.AddCommand(setupCmd())

	err := root.Execute()
	if err != nil {
		panic(err)
	}
}
