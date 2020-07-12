package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/sshst/sshststuff"
	"gopkg.in/src-d/go-git.v4"
	"os"
)

var listenConfig client.ListenConfig

func listenCmd() *cobra.Command {
	listenCmd := &cobra.Command{
		Use:   "listen",
		Short: "Listen for inbound connections",
		Run:   listen,
	}

	pf := listenCmd.PersistentFlags()
	pf.StringVar(&listenConfig.APIURL, "api", "https://api.ssh.st/api/listeners", "")
	pf.MarkHidden("api")

	pf.IntVar(&listenConfig.IdleTimeout, "idle-timeout", 30, "")
	pf.BoolVar(&listenConfig.WebOk, "web-ok", false, "")
	pf.StringVar(&listenConfig.NotifyUser, "notify-user", "", "")
	pf.StringVar(&listenConfig.NotifyTitle, "notify-title", "", "")
	pf.StringSliceVarP(&listenConfig.GithubUsers, "github", "g", []string{}, "")
	pf.StringSliceVarP(&listenConfig.SSHFingerprints, "fingerprint", "f", []string{}, "")

	return listenCmd
}

func listen(cmd *cobra.Command, args []string) {
	if len(listenConfig.NotifyUser) == 0 {
		listenConfig.NotifyUser = headCommitAuthor()
	}

	listenConfig.Version = version
	addGithubSshKeys(&listenConfig)

	if len(listenConfig.SSHFingerprints) == 0 {
		fmt.Fprintln(os.Stderr, "Error: No SSH keys provided (either directly or via GitHub) so it doesn't make sense to listen for no one. Exiting now.")
		os.Exit(1)
	}

	err := client.Listen(context.Background(), listenConfig)
	if err != nil {
		panic(err)
	}
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
