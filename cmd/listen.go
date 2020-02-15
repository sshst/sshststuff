package main

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/sshst/sshststuff"
	"os"
)

var listenConfig client.ListenConfig

func listenCmd() *cobra.Command {
	listenCmd := &cobra.Command{
		Use:   "listen",
		Short: "Listen for inbound connections",
		Run:   listen,
	}

	listenCmd.PersistentFlags().StringVar(&listenConfig.ApiUrl, "api", "https://api.ssh.st/api/listeners", "")
	listenCmd.PersistentFlags().MarkHidden("api")

	listenCmd.PersistentFlags().IntVar(&listenConfig.IdleTimeout, "idle-timeout", 30, "")
	listenCmd.PersistentFlags().BoolVar(&listenConfig.WebOk, "web-ok", true, "")
	// TODO command
	listenCmd.PersistentFlags().StringVar(&listenConfig.NotifyUser, "notify-user", "", "")
	listenCmd.PersistentFlags().StringVar(&listenConfig.NotifyTitle, "notify-title", "", "")
	listenCmd.PersistentFlags().StringSliceVarP(&listenConfig.GithubUsers, "github", "g", []string{}, "")
	listenCmd.PersistentFlags().StringSliceVarP(&listenConfig.SshFingerprints, "fingerprint", "f", []string{}, "")

	codebuildCmd := &cobra.Command{Use: "codebuild", Run: codebuild}
	codebuildCmd.PersistentFlags().Bool("always", false, "")
	listenCmd.AddCommand(codebuildCmd)

	githubCmd := &cobra.Command{Use: "github", Run: github}
	listenCmd.AddCommand(githubCmd)

	return listenCmd
}

func listen(cmd *cobra.Command, args []string) {
	err := client.Listen(context.Background(), listenConfig)
	if err != nil {
		panic(err)
	}
}

func codebuild(cmd *cobra.Command, args []string) {
	always, _ := cmd.PersistentFlags().GetBool("always")
	if os.Getenv("CODEBUILD_BUILD_SUCCEEDING") == "1" && !always {
		return
	}

	listen(cmd, args)
}

func github(cmd *cobra.Command, args []string) {
	listen(cmd, args)
}
