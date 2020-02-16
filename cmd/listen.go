package main

import (
	"context"
	"gopkg.in/src-d/go-git.v4"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/sshst/sshststuff"
)

var listenConfig client.ListenConfig

func listenCmd() *cobra.Command {
	listenCmd := &cobra.Command{
		Use:   "listen",
		Short: "Listen for inbound connections",
		Run:   listen,
	}

	listenCmd.PersistentFlags().StringVar(&listenConfig.APIURL, "api", "https://api.ssh.st/api/listeners", "")
	listenCmd.PersistentFlags().MarkHidden("api")

	listenCmd.PersistentFlags().IntVar(&listenConfig.IdleTimeout, "idle-timeout", 30, "")
	listenCmd.PersistentFlags().BoolVar(&listenConfig.WebOk, "web-ok", true, "")
	// TODO command
	listenCmd.PersistentFlags().StringVar(&listenConfig.NotifyUser, "notify-user", "", "")
	listenCmd.PersistentFlags().StringVar(&listenConfig.NotifyTitle, "notify-title", "", "")
	listenCmd.PersistentFlags().StringSliceVarP(&listenConfig.GithubUsers, "github", "g", []string{}, "")
	listenCmd.PersistentFlags().StringSliceVarP(&listenConfig.SSHFingerprints, "fingerprint", "f", []string{}, "")

	codebuildCmd := &cobra.Command{Use: "codebuild", Run: codebuild}
	codebuildCmd.PersistentFlags().Bool("always", false, "")
	listenCmd.AddCommand(codebuildCmd)

	githubCmd := &cobra.Command{Use: "github", Run: github}
	listenCmd.AddCommand(githubCmd)

	return listenCmd
}

func listen(cmd *cobra.Command, args []string) {
	if len(listenConfig.NotifyUser) == 0 {
		listenConfig.NotifyUser = headCommitAuthor()
	}

	listenConfig.Version = version
	
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

	if len(listenConfig.NotifyTitle) == 0 {
		id := os.Getenv("CODEBUILD_BUILD_ID")
		split := strings.SplitN(id, ":", 2)
		listenConfig.NotifyTitle = split[0]
	}

	listen(cmd, args)
}

func headCommitAuthor() string {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return ""
	}

	ref, err := repo.Head()
	if err != nil {
		return ""
	}

	cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return ""
	}

	commit, err := cIter.Next()
	if err != nil {
		return ""
	}

	return commit.Author.Email
}

func github(cmd *cobra.Command, args []string) {
	listen(cmd, args)
}
